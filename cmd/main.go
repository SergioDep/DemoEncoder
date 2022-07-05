package main

import (
	iparser "github.com/hx-w/minidemo-encoder/internal/parser"
	"io/ioutil"
	"log"
	"fmt"
	"strings"
)

func main() {
	files, err := ioutil.ReadDir("./demofiles/")
	if err != nil {
    log.Fatal(err)
	} else {
		for _, f := range files {
			filename := f.Name()
			if (strings.HasSuffix(filename, ".dem")) {
				fmt.Println("Parsing " + filename + "...")
				iparser.Start(filename)
			}
		}
	}
}
