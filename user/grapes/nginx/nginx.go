package nginx

import (
	"github.com/seletskiy/api"
)

func (p *Config) Ensure(cfg map[string]string) error {
	p := new(api.Package)
	//c := new(api.Config)
	
	if err := p.Ensure(map[string]string{"package": "nginx"}); err != nil {
		log.Println("error: " + err.Error())
	}

	//if err := c.Ensure(map[string]string{""})
}
