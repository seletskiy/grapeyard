package api

type Nginx struct{}

func (n *Nginx) Ensure(cfg map[string]string) error {
	p := new(Package)
	c := new(Config)
	
	if err := p.Ensure(map[string]string{"package": "nginx"}); err != nil {
		log.Println("nginx package error: " + err.Error())
	}

	if err := c.Ensure(cfg); err != nil {
		log.Println("nginx config error: " + err.Error())
	}
}
