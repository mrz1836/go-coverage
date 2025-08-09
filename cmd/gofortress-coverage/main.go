// Package main provides the GoFortress coverage CLI tool
package main

import (
	"os"

	"github.com/mrz1836/go-coverage/cmd/gofortress-coverage/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
