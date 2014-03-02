package yard

import (
	"github.com/BurntSushi/Toml"
)

type Yard struct {
	Hostname string
	Port     int
	Runlist  []string
}

func GetYard(yard *Yard, path string) error {
	if _, err := toml.DecodeFile(path, &yard); err != nil {
		return err
	}
	return nil
}
