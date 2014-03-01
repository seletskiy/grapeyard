package installer

import (
	"fmt"
	"os/exec"
)

func Install(pkg string) (result int) {
	cmd := exec.Command("pacman", "--noconfirm", "-S", pkg)
	sobuf := make([]byte, 512)
	sebuf := make([]byte, 512)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("failed to get stdout pipe; error:", err)
		return 1
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("failed to get stderr pipe; error:", err)
		return 1
	}

	if err := cmd.Start(); err != nil {
		stdout.Read(sobuf)
		stderr.Read(sebuf)
		fmt.Println("failed to install", pkg, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return 1
	}

	fmt.Println("waiting pacman to install", pkg)

	stdout.Read(sobuf)
	stderr.Read(sebuf)
	if err := cmd.Wait(); err != nil {
		fmt.Println("failed to install", pkg, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return 1
	}

	return 0
}
