package iface

import (
    "github.com/seletskiy/grapeyard/yard"
)

type Ensurer interface {
    Ensure(yard.Yard, map[string]string) error
}
