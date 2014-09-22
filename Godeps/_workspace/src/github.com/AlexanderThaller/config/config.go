package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/juju/errgo"
)

type Config interface {
	Default()
}

func Configure(pa string, co Config) error {
	_, err := os.Stat(pa)

	if !os.IsNotExist(err) {
		if err != nil {
			return errgo.New(err.Error())
		}

		err = Load(pa, co)
		if err != nil {
			return errgo.New(err.Error())
		}

		return nil
	}
	co.Default()

	err = Save(pa, co)
	if err != nil {
		return errgo.New(err.Error())
	}

	return nil
}

func Load(pa string, co Config) error {
	i, err := ioutil.ReadFile(pa)
	if err != nil {
		return errgo.New(err.Error())
	}

	var b bytes.Buffer
	json.Compact(&b, i)

	err = json.Unmarshal(b.Bytes(), co)
	if err != nil {
		return errgo.New(err.Error())
	}

	return nil
}

func Save(pa string, co Config) error {
	o, err := json.MarshalIndent(co, "", "  ")
	if err != nil {
		return errgo.New(err.Error())
	}

	err = ioutil.WriteFile(pa, o, 0644)
	if err != nil {
		return errgo.New(err.Error())
	}

	return nil
}
