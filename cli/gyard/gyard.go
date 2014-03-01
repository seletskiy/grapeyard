package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
)

const (
	VERSION = "Grapeyard 0.9, Mar 2014"
)

func main() {
	usage := `Grapeyard

Usage:
	gyard rape
	gyard -h | --help
	gyard -v | --version

Options:
	-h --help     Show this screen.
	-v --version  Show version.`

	args, _ := docopt.Parse(usage, nil, true, VERSION, false)
	fmt.Println("args:")
	fmt.Println(args)
	return
}
