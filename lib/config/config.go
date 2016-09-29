package config

import "io/ioutil"

/*

An attempt to make a relodable and flexible yet dead simple configuration file format

Example config:

bind localhost local.domain 127.0.0.1

default:
	box /mail/:name
	limit 100M

user admin:
	limit none

user ext:
	@include users/ext.conf

@include rest.conf

*/

type Config struct {
	Data Block
}

type Block []Property

type Property struct {
	Key    string
	Values []string
	Block  Block
}

func LoadConfig(path string) (Config, error) {
	cfg := Config{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	cfg.Data, err = parseConfig(path, string(data))
	if err != nil {
		return cfg, err
	}

	cfg.Data, err = processConfig(path, cfg.Data)
	return cfg, err
}
