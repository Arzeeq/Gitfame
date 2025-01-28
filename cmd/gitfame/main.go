//go:build !solution

package main

import (
	"io"
	"os"

	"gitlab.com/slon/shad-go/gitfame/internal/flag"
	"gitlab.com/slon/shad-go/gitfame/internal/git"
	"gitlab.com/slon/shad-go/gitfame/internal/printer"
)

func main() {
	flag, err := flag.Parse()
	if err != nil {
		_, e := io.WriteString(os.Stderr, err.Error())
		if e != nil {
			panic(err)
		}
		os.Exit(1)
	}

	stats := git.CalculateFiles(flag)
	printer.Print(stats, flag)
}
