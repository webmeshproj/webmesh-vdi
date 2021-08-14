package main

import (
	"log"

	"github.com/kvdi/kvdi/pkg/cmd"
)

func main() {
	if err := cmd.GenerateMarkdownDocs("doc/kvdictl"); err != nil {
		log.Fatal(err)
	}
}
