package api

import (
	"fmt"
	"os/exec"

    "github.com/seletskiy/grapeyard/yard"
)

type Service string

func (p *Service) Ensure(y yard.Yard, cfg map[string]string) error {
	service := cfg["service"]
	cmd := exec.Command("systemctl", "restart", service)
	sobuf := make([]byte, 512)
	sebuf := make([]byte, 512)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("failed to get stdout pipe; error:", err)
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("failed to get stderr pipe; error:", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		stdout.Read(sobuf)
		stderr.Read(sebuf)
		fmt.Println("failed to start service", service, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return err
	}

	fmt.Println("waiting systemctl to start service", service)

	stdout.Read(sobuf)
	stderr.Read(sebuf)
	if err := cmd.Wait(); err != nil {
		// TODO log instead of print
		fmt.Println("failed to start service", service, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return err
	}

	return nil
}
