package main

import (
	"log"

	"github.com/pathops/pathops-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}