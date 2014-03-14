package registry

import (
    "github.com/seletskiy/grapeyard/iface"
    "github.com/seletskiy/grapeyard/api"
)

func init() {
    Registry["api.File"] = func() iface.Ensurer { return new(api.File); }
    Registry["api.Nginx"] = func() iface.Ensurer { return new(api.Nginx); }

}
