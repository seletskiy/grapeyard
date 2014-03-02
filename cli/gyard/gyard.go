package main

import (
	"fmt"
	"github.com/docopt/docopt.go"

	"lib/gossip"
	"lib/httpapi"
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

		nodesCache, err := os.Open(args["nodescache"])
		if err != nil {
			panic("error opening state file for reading: " + err.Error())
		}

		nodesList := readNodesCache(args["nodescache"])
		hostname, err := os.Hostname()
		ver, _ := strconv.Atoi(args["version"])
		gossipPort, _ := strconv.Atoi(args["gossipport"])
		webPort, _ := strconv.Atoi(args["webport"])

		api := httpapi.Start(webPort)

		api.UploadImage(ver, os.Args[0])
		net.SendUpdateMsg(int64(ver), api.GetImageURI())

		conf := gossip.Config{
			RootNodes: nodesList,
			LocalPort: gossipPort,
			LocalVersion: ver,
			Name: fmt.Sprintf("%s:%d", hostname, args["gossipport"]),
		}

		net := gossip.NewGossipNetwork(conf, &gossip.ImmediateExecutor{})

		for {
			for _, m := range net.GetMembers() {
				fmt.Printf("[node] %s:%d\n", m.Name, m.Addr, m.Port)
			}

			time.Sleep(5 * time.Second)
		}
	}

	fmt.Println("no action selected, args:")
	fmt.Println(args)

	return
}

func readNodesCache(path string) []string {
	nodesList := make([]string, 0)
	for {
		var line string

		_, err := fmt.Fscanf(path, "%s\n", &line)
		if err != nil {
			break
		}

		nodesList = append(nodesList, line)
	}

	return nodesList
}
