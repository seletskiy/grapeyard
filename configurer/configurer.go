package configurer

import (
	"os"
	"github.com/seletskiy/grapeyard/yard"
	"text/template"
)

func Configure(yard yard.Yard, tplPath string, confPath string) error {
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
