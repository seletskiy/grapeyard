package api

import (
	"os"
	"github.com/seletskiy/grapeyard/yard"
	"text/template"
)

type Config struct {}

func (p *Config) Ensure(y yard.Yard, cfg map[string]string) error {
    return configure(y, cfg["from"], cfg["to"])
}

func configure(yard yard.Yard, tplPath string, confPath string) error {
	tmpl, err := template.New("template.conf").ParseFiles(tplPath)
	if err != nil {
		return err
	}

	f, err := os.Create(confPath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, yard)
	if err != nil {
		return err
	}

	f.Close()

	return nil
}
