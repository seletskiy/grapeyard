package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
	"github.com/seletskiy/grapeyard/lib/gossip"
	"github.com/seletskiy/grapeyard/lib/httpapi"
	"github.com/seletskiy/grapeyard/yard"
	"github.com/seletskiy/grapeyard/registry"
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
	gyard service-test
	gyard -h | --help
	gyard -v | --version

Options:
	--web-port=<webport>         Port to get binary packages from node [default: 8081].
	--gossip-port=<gossipport>   Port for communication between nodes using gossip protocol. [default: 2001].
	--extract-repo=<repooffset>  Offset in binary where repo begins. If flag is not specified or zero, no extract will be done. [default: 0].
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

        println("")
		nodesList := readNodesCache(args["<nodescache>"].(string))
		hostname, _ := os.Hostname()
		ver, _ := strconv.Atoi(args["<version>"].(string))
		gossipPort, _ := strconv.Atoi(args["--gossip-port"].(string))
		webPort, _ := strconv.Atoi(args["--web-port"].(string))

		api := httpapi.Start(webPort)

        println("http api running")
		api.UploadImage(ver, os.Args[0])
        println("image uploeaded")

		conf := gossip.Config{
			RootNodes:    nodesList,
			LocalPort:    gossipPort,
			LocalVersion: int64(ver),
			Name:         fmt.Sprintf("%s:%d", hostname, gossipPort),
		}

		if !args["-p"].(bool) {
			yard := getYard(args["-c"].(string))

            println("got yard")
			yardMap := map[string]string{
				"Hostname": yard.Hostname,
				"Port": strconv.Itoa(yard.Port),
				"to": "/etc/nginx/conf",
				"from": "grapes/nginx/nginx.conf.template",
			}

			for _, grape := range yard.Runlist {
				// @TODO config
                fmt.Println(registry.Registry)
				registry.Registry[grape]().Ensure(yardMap)
			}
		}

		net := gossip.NewGossipNetwork(conf, &gossip.ImmediateExecutor{args})
        println("acquired gossip")
        // FIXME: handle error
        extr_repo, _ := strconv.Atoi(args["--extract-repo"].(string))
		net.SendUpdateMsg(int64(ver), api.GetImageURI(),
            int64(extr_repo))
        println("updated")

		for {
			for _, m := range net.GetMembers() {
				fmt.Printf("[node] %s\n", m.Name)
			}

			if (args["-x"].(bool) || args["-p"].(bool)) {
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
