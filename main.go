package main

import (
	"flag"
	"io"
	"os"
	"path"
)

const (
	fileName     = ".ehviewer"
	backFileName = ".ehviewer.bak"
)

func main() {
	var inputDir = flag.String("input", "", "Dir of ehviewer download")
	var format = flag.Int("format", 1, "Output format. 1 for original text format, 2 for Overhauled binary format.")
	var restore = flag.Bool("restore", false, "Restore original backup")

	flag.Parse()
	if len(os.Args) == 2 {
		inputDir = &os.Args[1]
	}
	if *inputDir == "" || *format <= 0 || *format > 2 {
		println("Usage: [-input] [path] [-format]")
		flag.PrintDefaults()
		return
	}
	if *restore {
		doRestore(*inputDir)
	} else {
		doConvert(*inputDir, *format)
	}
}

func doConvert(input string, format int) {
	entries, err := os.ReadDir(input)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := path.Join(input, entry.Name())
		src := path.Join(dir, fileName)
		srcBak := path.Join(dir, backFileName)

		_, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				println("[ERROR]", entry.Name(), "has no metadata")
			} else {
				println("[ERROR]", entry.Name(), err.Error())
			}
			continue
		}

		// 备份原文件
		_, err = os.Stat(srcBak)
		if os.IsNotExist(err) {
			srcFile, err := os.Open(src)
			if err != nil {
				println("[ERROR]", entry.Name(), "open failed", err.Error())
				continue
			}
			defer srcFile.Close()
			dstFile, err := os.Create(srcBak)
			if err != nil {
				println("[ERROR]", entry.Name(), "backup failed", err.Error())
				continue
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				println("[ERROR]", entry.Name(), "backup failed", err.Error())
				continue
			}
		}

		// 读取信息
		srcFile, err := os.Open(srcBak)
		if err != nil {
			println("[ERROR]", entry.Name(), "open failed", err.Error())
			continue
		}
		defer srcFile.Close()

		info, err := NewInfo(srcFile)
		if err != nil {
			println("[ERROR]", entry.Name(), "parse failed", err.Error())
			continue
		}

		// 写入格式
		var data []byte
		switch format {
		case 1:
			var text string
			text, err = info.ToPlainText()
			data = []byte(text)
		case 2:
			data, err = info.ToCbor()
		default:
			panic("unknown format")
		}
		if err != nil {
			println("[ERROR]", entry.Name(), "convert failed", err.Error())
			continue
		}
		err = os.WriteFile(src, data, 0664)
		if err != nil {
			println("[ERROR]", entry.Name(), "write failed", err.Error())
			continue
		}
		println("[INFO ]", entry.Name(), "converted")
	}
}

func doRestore(input string) {
	entries, err := os.ReadDir(input)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := path.Join(input, entry.Name())
		src := path.Join(dir, fileName)
		srcBak := path.Join(dir, backFileName)

		_, err := os.Stat(srcBak)
		if err != nil {
			if os.IsNotExist(err) {
				println("[ERROR]", entry.Name(), "has no metadata")
			} else {
				println("[ERROR]", entry.Name(), err.Error())
			}
			continue
		}

		err = os.Rename(srcBak, src)
		if err != nil {
			println("[ERROR]", entry.Name(), "restore failed", err.Error())
			continue
		}

		println("[INFO ]", entry.Name(), "restored")
	}
}
