package api

import (
	"os"
	"text/template"
)

type Config struct {}

func (p *Config) Ensure(cfg map[string]string) error {
    return configure(cfg, cfg["from"], cfg["to"])
}

func configure(cfg, tplPath string, confPath string) error {
	tmpl, err := template.New("template.conf").ParseFiles(tplPath)
	if err != nil {
		return err
	}

	f, err := os.Create(confPath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, cfg)
	if err != nil {
		return err
	}

	f.Close()

	return nil
}

