package gossip

import (
	"os"
	"io"
	"io/ioutil"
	"log"
	"fmt"
	"archive/tar"
	"syscall"
)

type ImmediateExecutor struct {
	Args map[string]interface{}
}

func (ie *ImmediateExecutor) Run(binStream io.Reader, repoOffset int64, net *Network) {
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

	if repoOffset > 0 {
		_, err := io.CopyN(binFile, binStream, repoOffset)
		if err != nil {
			panic(err)
		}

		tr := tar.NewReader(binStream)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}

			if err != nil {
				panic(err)
			}

			log.Println("repo file %s", hdr.Name)
		}
	} else {
		_, err := io.Copy(binFile, binStream)
		if err != nil {
			panic(err)
		}
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

	log.Println(
		[]string{
			binPath,
			"rape",
			fmt.Sprintf("%d", net.GetVersion()),
			stateFile.Name(),
			fmt.Sprintf("--gossip-port=%s", ie.Args["--gossip-port"]),
			fmt.Sprintf("--web-port=%s", ie.Args["--web-port"]),
			"-c",
			fmt.Sprintf("%s", ie.Args["-c"]),
		})

	err = syscall.Exec(binPath,
		[]string{
			binPath,
			"rape",
			fmt.Sprintf("%d", net.GetVersion()),
			stateFile.Name(),
			fmt.Sprintf("--gossip-port=%s", ie.Args["--gossip-port"]),
			fmt.Sprintf("--web-port=%s", ie.Args["--web-port"]),
			"-c",
			fmt.Sprintf("%s", ie.Args["-c"]),
		},
		os.Environ())

	log.Println("[EMERGENCY ALERT] while upgrading binary: " + err.Error())
	log.Println("[EMERGENCY ALERT] binary was NOT upgraded!")
	log.Println("[EMERGENCY ALERT] Р”Р°Р@ЅРЅР°СЏ СЃС‚СЂ@Р°РЅРёС†Р° РґРѕС")
}
