package registry

import (
    "github.com/seletskiy/grapeyard/iface"
)

type GenFunc func() iface.Ensurer
var Registry = make(map[string]GenFunc)


