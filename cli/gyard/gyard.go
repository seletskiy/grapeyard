package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
	"grapeyard/installer"
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

	if args["rape"].(bool) {
		fmt.Println("rapeme!1")
		fmt.Println("installing ntp")
		installer.Install("ntp")
		fmt.Println("installation finished")
		return
	}

	fmt.Println("no action selected, args:")
	fmt.Println(args)

	return
}
