package api

import (
    "github.com/seletskiy/grapeyard/yard"
)

type Nginx struct{}


func (n *Nginx) Ensure(y yard.Yard, cfg map[string]string) error {
    // install
    var p *Package

    p.install("nginx")
    return nil
}
