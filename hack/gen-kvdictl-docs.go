package main

import (
	"log"

	"github.com/tinyzimmer/kvdi/pkg/cmd"
)

func main() {
	if err := cmd.GenerateMarkdownDocs("doc/kvdictl"); err != nil {
		log.Fatal(err)
	}
}
