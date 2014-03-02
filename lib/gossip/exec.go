package gossip

import (
	"os"
	"io"
	"io/ioutil"
	"log"
	"fmt"
	"syscall"
)

type ImmediateExecutor struct {
	Args map[string]interface{}
}

func (ie *ImmediateExecutor) Run(binStream io.Reader, net *Network) {
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

	log.Println(
		[]string{
			binPath,
			"rape",
			fmt.Sprintf("%d", net.GetVersion()),
			stateFile.Name(),
			fmt.Sprintf("--gossip-port=%s", ie.Args["--gossip-port"]),
			fmt.Sprintf("--web-port=%s", ie.Args["--web-port"]),
		})

	err = syscall.Exec(binPath,
		[]string{
			binPath,
			"rape",
			fmt.Sprintf("%d", net.GetVersion()),
			stateFile.Name(),
			fmt.Sprintf("--gossip-port=%s", ie.Args["--gossip-port"]),
			fmt.Sprintf("--web-port=%s", ie.Args["--web-port"]),
		},
		os.Environ())

	log.Println("[EMERGENCY ALERT] while upgrading binary: " + err.Error())
	log.Println("[EMERGENCY ALERT] binary was NOT upgraded!")
	log.Println("[EMERGENCY ALERT] Р”Р°Р@ЅРЅР°СЏ СЃС‚СЂ@Р°РЅРёС†Р° РґРѕС")
}
