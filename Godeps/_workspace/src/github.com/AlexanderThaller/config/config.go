package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/juju/errgo"
)

type Format int

const (
	FormatJSON Format = iota
	FormatTOML
)

type Config interface {
	Default()
	Format() Format
}

func Configure(pa string, co Config) error {
	_, err := os.Stat(pa)

	if !os.IsNotExist(err) {
		if err != nil {
			return errgo.New(err.Error())
		}

		err = Load(pa, co)
		if err != nil {
			return err
		}

		return nil
	}
	co.Default()

	err = Save(pa, co)
	if err != nil {
		return err
	}

	return nil
}

func Load(pa string, co Config) error {
	switch co.Format() {
	case FormatJSON:
		return loadJSON(pa, co)
	case FormatTOML:
		return loadTOML(pa, co)
	default:
		return errgo.New("do not understand this format")
	}
}

func loadJSON(pa string, co Config) error {
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

func loadTOML(pa string, co Config) error {
	data, err := ioutil.ReadFile(pa)
	if err != nil {
		return errgo.New(err.Error())
	}

	_, err = toml.Decode(string(data), co)
	if err != nil {
		return errgo.New(err.Error())
	}

	return nil
}

func Save(pa string, co Config) error {
	switch co.Format() {
	case FormatJSON:
		return saveJSON(pa, co)
	case FormatTOML:
		return saveTOML(pa, co)
	default:
		return errgo.New("do not understand this format")
	}
}

func saveJSON(pa string, co Config) error {
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

func saveTOML(pa string, co Config) error {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(co)
	if err != nil {
		return errgo.New(err.Error())
	}

	err = ioutil.WriteFile(pa, buf.Bytes(), 0644)
	if err != nil {
		return errgo.New(err.Error())
	}
	return nil
}
