package main

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"github.com/docopt/docopt.go"

	"github.com/seletskiy/grapeyard/lib/gossip"
	"github.com/seletskiy/grapeyard/lib/httpapi"
)

const (
	VERSION = "Grapeyard 0.9, Mar 2014"
)

func main() {
	usage := `Grapeyard

Usage:
	gyard rape <version> <nodescache> [--web-port=<webport>] [--gossip-port=<gossipport>]
	gyard -h | --help
	gyard -v | --version

Options:
	--web-port=<webport>         Port to get binary packages from node [default: 8081].
	--gossip-port=<gossipport>   Port for communication between nodes using gossip protocol. [default: 2001]
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
			RootNodes: nodesList,
			LocalPort: gossipPort,
			LocalVersion: int64(ver),
			Name: fmt.Sprintf("%s:%d", hostname, gossipPort),
		}

		net := gossip.NewGossipNetwork(conf, &gossip.ImmediateExecutor{args})
		net.SendUpdateMsg(int64(ver), api.GetImageURI())

		for {
			for _, m := range net.GetMembers() {
				fmt.Printf("[node] %s\n", m.Name)
			}

			time.Sleep(5 * time.Second)
		}
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
