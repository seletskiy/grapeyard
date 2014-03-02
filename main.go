package main

import (
	"fmt"
	"strconv"
	"os"
	"log"
	"time"
	"syscall"
	"io"
	"io/ioutil"

	"gossip"
	"httpapi"
)

type ImmediateExecutor struct{}

func (ie *ImmediateExecutor) Run(binStream io.Reader, net *gossip.Network) {
	// @FIXME
	binPath := os.Args[0]
	log.Printf("replacing binary %s", binPath)
	oldInfo, err := os.Stat(binPath)
	if err != nil {
		panic("error retrieving binary mode: " + err.Error())
	}
	tmpBinPath := binPath + ".new"
	binFile, err := os.Create(tmpBinPath)
	if err != nil {
		panic("error creating new binary: " + err.Error())
	}
	for {
		buf := make([]byte, 255)
		n, _ := binStream.Read(buf)
		if n == 0 {
			break
		}
		binFile.Write(buf[:n])
	}

	binFile.Close()

	os.Chmod(tmpBinPath, oldInfo.Mode())

	stateFile, err := ioutil.TempFile(os.TempDir(), "grape-")
	if err != nil {
		panic("can't create state transient file")
	}

	for _, m := range net.GetMembers() {
		fmt.Fprintf(stateFile, "%s:%d\n", m.Addr, m.Port)
	}

	stateFile.Close()

	err = os.Rename(tmpBinPath, binPath)
	if err != nil {
		panic("error replacing old binary: " + err.Error())
	}

	panic(syscall.Exec(binPath,
		[]string{
			os.Args[0], os.Args[1], os.Args[2],
			fmt.Sprintf("%d", net.GetVersion()),
			stateFile.Name(),
		},
		os.Environ()))
}

func main() {
	gossipPort, _ := strconv.Atoi(os.Args[1])
	webPort, _ := strconv.Atoi(os.Args[2])
	ver, _ := strconv.Atoi(os.Args[3])
	log.Println("running at v.", ver)

	conf := gossip.Config{
		RootNodes: []string{
			"127.1:2001",
		},
		LocalPort: gossipPort,
		LocalVersion: int64(ver),
		Name: fmt.Sprintf("%d", gossipPort),
	}

	net := gossip.NewGossipNetwork(conf, &ImmediateExecutor{})
	
	if len(os.Args) > 4 {
		fmt.Println(os.Args)
		stateFile, err := os.Open(os.Args[4])
		if err != nil {
			panic("error opening state file for reading: " + err.Error())
		}
		for {
			var line string
			_, err := fmt.Fscanf(stateFile, "%s\n", &line)
			if err != nil {
				break
			}

			log.Printf("state: %s", line)

			n, _ := net.Join(line)
			if n > 0 {
				log.Println("member joined again: ", line)
			}
		}
	}

	api := httpapi.Start(webPort)

	api.UploadImage(ver, os.Args[0])
	net.SendUpdateMsg(int64(ver), api.GetImageURI())

	for {
		for _, m := range net.GetMembers() {
			fmt.Printf("[%s] %s:%d\n", m.Name, m.Addr, m.Port)
		}

		time.Sleep(5 * time.Second)
	}
}
