// Package main provides the Go coverage CLI tool
package main

import (
	"os"

	"github.com/mrz1836/go-coverage/cmd/go-coverage/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
