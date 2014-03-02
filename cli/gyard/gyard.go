package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
	"github.com/seletskiy/grapeyard/api"
	"github.com/seletskiy/grapeyard/configurer"
	"github.com/seletskiy/grapeyard/lib/gossip"
	"github.com/seletskiy/grapeyard/lib/httpapi"
	"github.com/seletskiy/grapeyard/yard"
	"os"
	"strconv"
	"time"
)

const (
	VERSION = "Grapeyard 0.9, Mar 2014"
)

func main() {
	usage := `Grapeyard

Usage:
	gyard [options] rape <version> <nodescache>
	gyard yard-test
	gyard package-test
	gyard conf-test
	gyard -h | --help
	gyard -v | --version

Options:
	--web-port=<webport>         Port to get binary packages from node [default: 8081].
	--gossip-port=<gossipport>   Port for communication between nodes using gossip protocol. [default: 2001].
	--extract-repo=<repooffset>  Offset in binary where repo begins. If flag is not specified, no extract will be done. [default: 0].
	-c <yardpath>                Path to config of the yard [default: ./yard.toml].
	-x                           Exit after raping.
	-p                           Propagate only. Do not configure current node. Implies "-x".
	-h --help                    Show this screen.
	-v --version                 Show version.`

	args, _ := docopt.Parse(usage, nil, true, VERSION, false)

	if args["rape"].(bool) {
		fmt.Println("action: rape")
		fmt.Println(args)

		fmt.Println("RUNNING v." + args["<version>"].(string))

		nodesList := readNodesCache(args["<nodescache>"].(string))
		hostname, _ := os.Hostname()
		ver, _ := strconv.Atoi(args["<version>"].(string))
		gossipPort, _ := strconv.Atoi(args["--gossip-port"].(string))
		webPort, _ := strconv.Atoi(args["--web-port"].(string))

		api := httpapi.Start(webPort)

		api.UploadImage(ver, os.Args[0])

		conf := gossip.Config{
			RootNodes:    nodesList,
			LocalPort:    gossipPort,
			LocalVersion: int64(ver),
			Name:         fmt.Sprintf("%s:%d", hostname, gossipPort),
		}

		if !args["-p"].(bool) {
			yard := getYard(args["-c"].(string))
			for _, grape := range yard.Runlist {
				// @TODO
				panic("apply grape here! " + grape)
			}

			return
		}

		net := gossip.NewGossipNetwork(conf, &gossip.ImmediateExecutor{args})
		net.SendUpdateMsg(int64(ver), api.GetImageURI(), args["--extract-repo"].(int64))

		for {
			for _, m := range net.GetMembers() {
				fmt.Printf("[node] %s\n", m.Name)
			}

			if (args["-x"].(bool)) {
				return
			}

			time.Sleep(5 * time.Second)
		}
	}

	if args["yard-test"].(bool) {
		fmt.Println("action: yard-test")
		var y yard.Yard
		err := yard.GetYard(&y,
			"test/yard/yard.toml")
		if err != nil {
			fmt.Println("err", err)
		}
		fmt.Println(y)
		return
	}

	if args["package-test"].(bool) {
		fmt.Println("action: package-test")
		var p = new(api.Package)
		err := p.Ensure(map[string]string{"package": "ntp"})
		if err != nil {
			fmt.Println("error:", err)
		}
		return
	}

	if args["conf-test"].(bool) {
		fmt.Println("action: conf-test")
		yard := yard.Yard{"localhost", 80, []string{"nginx", "ntp"}}
		err := configurer.Configure(
			yard,
			"test/configurer/template.conf",
			"test/configurer/result.conf")
		if err != nil {
			fmt.Println("err", err)
		}
		return
	}

	fmt.Println("no action selected, args:")
	fmt.Println(args)

	return
}

func readNodesCache(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		panic("can't open nodes list: " + err.Error())
	}
	nodesList := make([]string, 0)
	for {
		var line string

		_, err := fmt.Fscanf(file, "%s\n", &line)
		if err != nil {
			break
		}

		nodesList = append(nodesList, line)
	}

	return nodesList
}

func getYard(path string) yard.Yard {
	var y yard.Yard
	err := yard.GetYard(&y, path)
	if err != nil {
		fmt.Println("err", err)
	}
	return y
}
