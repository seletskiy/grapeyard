package api

import (
	"fmt"
	"os/exec"
	"syscall"
)

type Package string

func (p Package) Ensure(map[string]string) error {
	installed, err := p.isInstalled()
	if err != nil {
		return err
	}
	if installed {
		return nil
	}
	if err := p.install(); err != nil {
		return err
	}
	return nil
}

func (p Package) install() error {
	cmd := exec.Command("pacman", "--noconfirm", "-S", string(p))
	sobuf := make([]byte, 512)
	sebuf := make([]byte, 512)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		// TODO log instead of print
		fmt.Println("failed to get stdout pipe; error:", err)
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		// TODO log instead of print
		fmt.Println("failed to get stderr pipe; error:", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		stdout.Read(sobuf)
		stderr.Read(sebuf)
		// TODO log instead of print
		fmt.Println("failed to install", p, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return err
	}

	fmt.Println("waiting pacman to install", p)

	stdout.Read(sobuf)
	stderr.Read(sebuf)
	if err := cmd.Wait(); err != nil {
		// TODO log instead of print
		fmt.Println("failed to install", p, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return err
	}

	return nil
}

// FIXME finish the function implementation
func (p Package) isInstalled() (bool, error) {
	cmd := exec.Command(
		"pacman", "-Qsq", string(p), "|", "grep", "^"+string(p)+"$")
	sobuf := make([]byte, 512)
	sebuf := make([]byte, 512)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		// TODO log instead of print
		fmt.Println("failed to get stdout pipe; error:", err)
		return false, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		// TODO log instead of print
		fmt.Println("failed to get stderr pipe; error:", err)
		return false, err
	}

	if err := cmd.Start(); err != nil {
		stdout.Read(sobuf)
		stderr.Read(sebuf)
		// TODO log instead of print
		fmt.Println("failed to install", p, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return false, err
	}

	fmt.Println("waiting pacman to install", p)

	stdout.Read(sobuf)
	stderr.Read(sebuf)
	if err := cmd.Wait(); err != nil {
		// TODO log instead of print
		fmt.Println("failed to install", p, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))

		if msg, ok := err.(*exec.ExitError); ok {
			exitCode := msg.Sys().(syscall.WaitStatus).ExitStatus()
			if exitCode == 1 {
				fmt.Println("package", p, "already installed")
				return false, nil
			}
		}

		return false, err
	}

	return true, nil
}