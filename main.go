//usr/bin/env go run $0 $@; exit $?
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func toUtf8(str string, t transform.Transformer) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(str), t))
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func unzip(src, dest string, listOnly bool, t transform.Transformer) error {
	zc, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zc.Close()

	for _, item := range zc.File {
		fname, err := toUtf8(item.Name, t)
		if err != nil {
			fname = item.Name
		}
		path := filepath.Join(dest, fname)

		// a directory this concrete file is in could be stored
		// in the ZIP file index after the file, so create
		// any intermediate paths ahead of time
		dir, _ := filepath.Split(path)
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}

		// item itself is a directory
		if item.FileInfo().IsDir() {
			log.Printf("Directory '%v'", path)
			if !listOnly {
				if err := os.MkdirAll(path, 0755); err != nil {
					return err
				}
			}
		} else {
			// otherwise, it's a concrete file
			log.Printf("File '%v'", path)
			if !listOnly {
				output, err := os.Create(path)
				if err != nil {
					return err
				}
				defer output.Close()
				fp, err := item.Open()
				if err != nil {
					return err
				}
				defer fp.Close()
				if _, err := io.Copy(output, fp); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func main() {
	dest := "./"
	listOnly := false
	flag.StringVar(&dest, "d", dest, "destination folder")
	flag.BoolVar(&listOnly, "l", listOnly, "just list the contents of the zip file")
	flag.Parse()
	fmt.Println("dest:", dest)
	err := unzip(flag.Arg(0), dest, listOnly, japanese.ShiftJIS.NewDecoder())
	if err != nil {
		log.Fatal(err)
	}
}
