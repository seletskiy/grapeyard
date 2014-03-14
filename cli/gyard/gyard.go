package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/docopt/docopt.go"
	"github.com/seletskiy/grapeyard/lib/gossip"
	"github.com/seletskiy/grapeyard/lib/httpapi"
	"github.com/seletskiy/grapeyard/yard"
	//"github.com/seletskiy/grapeyard/registry"
	//"github.com/mechmind/git-go/git"
	"github.com/seletskiy/grapeyard/builder"

	"github.com/mechmind/git-go/git"
	"github.com/seletskiy/grapeyard/fs"
)

const (
	VERSION = `Grapeyard 0.9, Mar 2014`
	USAGE   = `Grapeyard

Usage:
	gyard [options] rape <version> <nodescache>
	gyard [options] build [<ref-spec>]
	gyard [options] extract
	gyard -h | --help
	gyard -v | --version

Options:
	-c <yardpath>                Path to config of the yard [default: ./yard.toml].
	-x                           Exit after raping. Do not launch daemon mode.
	-p                           Propagate and exit. Do not configure current node.
	--web-port=<webport>         Port to get binary packages from node [default: 8081].
	--gossip-port=<gossipport>   Port for communication between nodes using gossip protocol. [default: 2001].
	-h --help                    Show this screen.
	-v --version                 Show version.`
)

func main() {
	args, _ := docopt.Parse(USAGE, nil, true, VERSION, false)

	if args["build"].(bool) {
		err := buildGyard()
		if err != nil {
			fmt.Println("error while building binary: " + err.Error())
		}

		return
	}

	if args["extract"].(bool) {
		// just testing
		self, err := os.Open(os.Args[0])
		if err != nil {
			panic(err)
		}
		emfs, err := fs.OpenEmbedFs(self)
		if err != nil {
			panic(err)
		}
		repo, err := git.OpenFsRepo(emfs)
		if err != nil {
			panic(err)
		}
		head, err := repo.ReadSymbolicRef("HEAD")
		if err != nil {
			panic(err)
		}
		fmt.Println("HEAD is", head)
		return
	}

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

		//if !args["-p"].(bool) {
		//    yard := getYard(args["-c"].(string))
		//    yardMap := map[string]string{
		//        "Hostname": yard.Hostname,
		//        "Port": yard.Port,
		//    }

		//    for _, grape := range yard.Runlist {
		//        // @TODO config
		//        registry.Registry[grape]().Ensure(yardMap)
		//    }

		//    return
		//}

		net := gossip.NewGossipNetwork(conf, &gossip.ImmediateExecutor{args})
		net.SendUpdateMsg(int64(ver), api.GetImageURI(), args["--extract-repo"].(int64))

		for {
			for _, m := range net.GetMembers() {
				fmt.Printf("[node] %s\n", m.Name)
			}

			if args["-x"].(bool) {
				return
			}

			time.Sleep(5 * time.Second)
		}
	}

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

func buildGyard() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = builder.BuildExecutable(cwd)
	if err != nil {
		return err
	}

	return nil
}
