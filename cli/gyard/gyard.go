package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docopt/docopt.go"
	"github.com/seletskiy/grapeyard/lib/gossip"
	"github.com/seletskiy/grapeyard/lib/httpapi"
	"github.com/seletskiy/grapeyard/yard"
	//"github.com/seletskiy/grapeyard/registry"
	//"github.com/mechmind/git-go/git"
	"github.com/seletskiy/grapeyard/builder"
)

const BASE_URL = "github.com/seletskiy/grapeyard"
const MAIN_EXE = "cli/gyard"

const RELEASE_DIR = "/tmp/grapeyard"

const (
	VERSION = `Grapeyard 0.9, Mar 2014`
	USAGE   = `Grapeyard

Usage:
	gyard [options] rape <version> <nodescache>
	gyard [options] build [<ref-spec>]
	gyard -h | --help
	gyard -v | --version

Options:
	-c <yardpath>                Path to config of the yard [default: ./yard.toml].
	-x                           Exit after raping. Do not launch daemon mode.
	-p                           Propagate and exit. Do not configure current node.
	-l                           Use local grapes directory, not embedded one.
	--web-port=<webport>         Port to get binary packages from node [default: 8081].
	--gossip-port=<gossipport>   Port for communication between nodes using gossip protocol. [default: 2001].
	-h --help                    Show this screen.
	-v --version                 Show version.`
)

func main() {
	args, _ := docopt.Parse(USAGE, nil, true, VERSION, false)

	if args["build"].(bool) {
		buildGyard(args["-l"].(bool))
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

func buildGyard(local bool) {
	err := deployCurrentBranch(local)
	if err != nil {
		fmt.Println("error building binary: " + err.Error())
		os.Exit(1)
	}
}

func deployCurrentBranch(local bool) error {
	// resolve HEAD and build seed from that branch
	// pwd will be at root of repo

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	//repo, err := git.OpenRepo(cwd)
	//if err != nil {
	//    return err
	//}

	//branch, err := repo.ReadSymbolicRef("HEAD")

	//if err != nil {
	//    return err
	//}

	//tempDir, err := ioutil.TempDir("/tmp", "grape-build.")
	//if err != nil {
	//    return err
	//}
	tempDir := cwd

	log.Println("[build] building into " + tempDir)

	//buildDir := filepath.Join(tempDir, "src", BASE_URL)
	buildDir := cwd

	log.Println("[build] building from " + buildDir)

	//err = builder.ExtractTree(repo, branch, buildDir)
	//if err != nil {
	//    return err
	//}

	// build binary
	registry, err := builder.MakeRegistry(buildDir)
	if err != nil {
		return err
	}

	err = builder.WriteRegistry(registry, buildDir, BASE_URL)
	if err != nil {
		return err
	}

	// invoke go build to make binary
	cmdArgs := []string{"install", filepath.Join(BASE_URL, MAIN_EXE)}

	log.Println("[builder] will run go " + strings.Join(cmdArgs, " "))
	cmd := exec.Command("go", cmdArgs...)
	cmdEnv := make([]string, len(os.Environ()))
	copy(cmdEnv, os.Environ())

	// find GOPATH and replace
	var gopathFound bool

	for idx, vr := range cmdEnv {
		if strings.HasPrefix(vr, "GOPATH=") {
			cmdEnv[idx] = "GOPATH=" + tempDir + ":" + vr[7:]
			gopathFound = true
			break
		}
	}
	if !gopathFound {
		cmdEnv = append(cmdEnv, "GOPATH="+tempDir)
	}

	cmd.Env = cmdEnv

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	output, _ := cmd.CombinedOutput()
	log.Printf("[builder] go install says:\n%s", output)

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	//sourceBinary := filepath.Join(tempDir, "bin", filepath.Base(MAIN_EXE))
	sourceBinary := os.Args[0]

	stat, err := os.Stat(sourceBinary)
	if err != nil {
		return err
	}

	exeLength := stat.Size()
	seedName := filepath.Base(MAIN_EXE) + "." + strconv.Itoa(int(exeLength))
	seedPath := filepath.Join(tempDir, seedName)

	seedFile, err := os.Create(seedPath)
	if err != nil {
		return err
	}
	os.Chmod(seedPath, 0755)

	sourceFile, err := os.Open(sourceBinary)
	if err != nil {
		return err
	}

	io.Copy(seedFile, sourceFile)

	// build tar
	// strip go sources from user/ dir

	userDir := filepath.Join(buildDir, "user", "")
	err = builder.StripSources(userDir)
	if err != nil {
		return err
	}

	// append tarred data
	tarWriter := tar.NewWriter(seedFile)

	var dirs = []string{}

	tlFileInfos, err := ioutil.ReadDir(userDir)
	if err != nil {
		return err
	}

	for _, tlfinfo := range tlFileInfos {
		if tlfinfo.IsDir() {
			dirs = append(dirs, tlfinfo.Name())
		}
	}

	for {
		if len(dirs) == 0 {
			break
		}

		var newDirs []string

		for _, dir := range dirs {
			files, err := ioutil.ReadDir(filepath.Join(userDir, dir))
			if err != nil {
				return err
			}
			for _, file := range files {
				fullPath := filepath.Join(dir, file.Name())
				if file.IsDir() {
					newDirs = append(newDirs, fullPath)
				} else {
					tarHeader, err := tar.FileInfoHeader(file, "")
					if err != nil {
						return err
					}

					tarHeader.Name = filepath.Join(dir, tarHeader.Name)

					err = tarWriter.WriteHeader(tarHeader)
					if err != nil {
						return err
					}

					sourceFile, err := os.Open(filepath.Join(userDir, fullPath))
					if err != nil {
						return err
					}

					_, err = io.Copy(tarWriter, sourceFile)
					if err != nil {
						return err
					}

					sourceFile.Close()
				}
			}
		}
		dirs = newDirs
	}

	// move to RELEASE_DIR
	if _, err := os.Stat(RELEASE_DIR); os.IsNotExist(err) {
		os.Mkdir(RELEASE_DIR, 0777)
	}

	os.Rename(seedPath, filepath.Join(RELEASE_DIR, seedName))
	return nil
}
