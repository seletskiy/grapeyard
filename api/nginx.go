package api

import (
	"log"
)

type Nginx struct{}

func (n Nginx) Ensure(cfg map[string]string) error {
	p := new(Package)
	c := new(Config)
	s := new(Service)
	
	if err := p.Ensure(map[string]string{"package": "nginx"}); err != nil {
		log.Println("nginx package error: " + err.Error())
	}

	if err := c.Ensure(map[string]string{"from": "grapes/nginx/nginx.conf.template", "to": "/etc/nginx/nginx.conf"}); err != nil {
		log.Println("nginx config error: " + err.Error())
	}

	if err := c.Ensure(map[string]string{"from": "grapes/nginx/index.html.template", "to": "/etc/nginx/index.html"}); err != nil {
		log.Println("nginx config error: " + err.Error())
	}

	if err := s.Ensure(cfg); err != nil {
		log.Println("nginx service error: " + err.Error())
	}

	return nil
}
