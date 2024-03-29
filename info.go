package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/fxamacker/cbor/v2"
)

type SpiderInfo struct {
	StartPage      int            `cbor:"startPage"`
	GID            int            `cbor:"gid"`
	Token          string         `cbor:"token"`
	PreviewPages   int            `cbor:"previewPages"`
	PreviewPerPage int            `cbor:"previewPerPage"`
	Pages          int            `cbor:"pages"`
	TokenMap       map[int]string `cbor:"pTokenMap"`
}

func (i *SpiderInfo) ToPlainText() (string, error) {
	var sb = strings.Builder{}
	sb.WriteString("VERSION2\n")
	sb.WriteString(fmt.Sprintf("%08x\n", i.StartPage))
	sb.WriteString(fmt.Sprintf("%d\n", i.GID))
	sb.WriteString(i.Token)
	sb.WriteRune('\n')
	sb.WriteString("1\n")
	sb.WriteString(fmt.Sprintf("%d\n", i.PreviewPages))
	sb.WriteString(fmt.Sprintf("%d\n", i.PreviewPerPage))
	sb.WriteString(fmt.Sprintf("%d\n", i.Pages))

	keys := make([]int, 0)
	for k := range i.TokenMap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(lhs, rhs int) bool {
		return lhs < rhs
	})

	for k := range keys {
		sb.WriteString(fmt.Sprintf("%d %s\n", k, i.TokenMap[k]))
	}

	return sb.String(), nil
}

func (i *SpiderInfo) ToCbor() ([]byte, error) {
	return cbor.Marshal(i)
}

func NewInfoFromCbor(data io.Reader) (*SpiderInfo, error) {
	bytes, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}

	var info SpiderInfo
	_, err = cbor.UnmarshalFirst(bytes, &info)
	return &info, err
}

var errInvalidFormat = errors.New("invalid format")

func NewInfoFromPlain(data io.Reader) (*SpiderInfo, error) {
	var info SpiderInfo

	rd := bufio.NewScanner(data)
	rd.Split(bufio.ScanLines)

	// VERSION2
	if !rd.Scan() {
		return nil, errInvalidFormat
	}
	if rd.Text() != "VERSION2" {
		return nil, errInvalidFormat
	}

	// startPage
	if !rd.Scan() {
		return nil, errInvalidFormat
	}
	pageNum, err := strconv.ParseUint(rd.Text(), 16, 64)
	if err != nil {
		return nil, err
	}
	info.StartPage = int(pageNum)

	// gid
	if !rd.Scan() {
		return nil, errInvalidFormat
	}
	gid, err := strconv.ParseUint(rd.Text(), 10, 64)
	if err != nil {
		return nil, err
	}
	info.GID = int(gid)

	// token
	if !rd.Scan() {
		return nil, errInvalidFormat
	}
	info.Token = rd.Text()

	// mode, deprecated
	if !rd.Scan() {
		return nil, errInvalidFormat
	}

	// previewPages
	if !rd.Scan() {
		return nil, errInvalidFormat
	}
	previewPages, err := strconv.ParseUint(rd.Text(), 10, 64)
	if err != nil {
		return nil, err
	}
	info.PreviewPages = int(previewPages)

	// previewPerPage
	if !rd.Scan() {
		return nil, errInvalidFormat
	}
	previewPerPage, err := strconv.ParseUint(rd.Text(), 10, 64)
	if err != nil {
		return nil, err
	}
	info.PreviewPerPage = int(previewPerPage)

	// pages
	if !rd.Scan() {
		return nil, errInvalidFormat
	}
	pages, err := strconv.ParseUint(rd.Text(), 10, 64)
	if err != nil {
		return nil, err
	}
	info.Pages = int(pages)

	info.TokenMap = make(map[int]string)
	for rd.Scan() {
		text := rd.Text()
		arr := strings.Split(text, " ")
		if len(arr) != 2 {
			return nil, errInvalidFormat
		}

		id, err := strconv.ParseUint(arr[0], 10, 64)
		if err != nil {
			return nil, err
		}

		info.TokenMap[int(id)] = arr[1]
	}

	return &info, nil
}

func NewInfo(data io.ReadSeeker) (*SpiderInfo, error) {
	head := make([]byte, 1)
	_, err := data.Read(head)
	if err != nil {
		return nil, err
	}

	_, err = data.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	if head[0] == 'V' {
		return NewInfoFromPlain(data)
	}
	if head[0]&0xE0 == 0b10100000 {
		return NewInfoFromCbor(data)
	}
	return nil, errInvalidFormat
}
