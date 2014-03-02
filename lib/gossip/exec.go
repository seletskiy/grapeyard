package gossip

type ImmediateExecutor struct{}

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

	panic(syscall.Exec(binPath,
		[]string{
			os.Args[0], os.Args[1], os.Args[2],
			fmt.Sprintf("%d", net.GetVersion()),
			stateFile.Name(),
		},
		os.Environ()))
}
