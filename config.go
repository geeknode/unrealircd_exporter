package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Listen string
	Link   string
	Name   string
	Sid    int
	Cert   string
	Key    string
}

func LoadConfig(configFile string) (*Config, error) {
	var conf Config
	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
