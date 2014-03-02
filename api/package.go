package api

import (
	"fmt"
	"os/exec"

    "github.com/seletskiy/grapefruit/yard"
)

type Package struct {}

func (p *Package) Ensure(y yard.Yard, cfg map[string]string) error {
	pkg := cfg["package"]
	installed, err := p.isInstalled(pkg)
	if err != nil {
		return err
	}
	if installed {
		fmt.Println("already installed, nothing to do")
		return nil
	}
	if err := p.install(pkg); err != nil {
		return err
	}
	return nil
}

func (p *Package) install(pkg string) error {
	cmd := exec.Command("pacman", "--noconfirm", "-S", pkg)
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
		fmt.Println("failed to install", pkg, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return err
	}

	fmt.Println("waiting pacman to install", pkg)

	stdout.Read(sobuf)
	stderr.Read(sebuf)
	if err := cmd.Wait(); err != nil {
		// TODO log instead of print
		fmt.Println("failed to install", pkg, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return err
	}

	return nil
}

// FIXME finish the function implementation
func (p *Package) isInstalled(pkg string) (bool, error) {
	cmd := exec.Command("pacman", "-Qi", pkg)
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
		fmt.Println("failed to install", pkg, "; error:", err)
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))
		return false, err
	}

	stdout.Read(sobuf)
	stderr.Read(sebuf)
	if err := cmd.Wait(); err != nil {
		// TODO log instead of print
		fmt.Println("stdout", string(sobuf))
		fmt.Println("stderr", string(sebuf))

		fmt.Println("package", pkg, "not installed")
		return false, nil
	}
	fmt.Println("package", pkg, "already installed")
	return true, nil
}
