package config

import (
	"errors"
	"strings"
)

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
	Values []string
	Block  Block
}

type QueryResult struct {
	Single   string
	Property *Property
}

var (
	ErrLCannotReadFile  = errors.New("load cfg error: cannot read file")
	ErrLIncludeFailed   = errors.New("load cfg error: cannot include file")
	ErrQNotFound        = errors.New("query cfg error: property not found")
	ErrQDifferentFormat = errors.New("query cfg error: format mismatch")
)

func LoadConfig() (Config, error) {
	return Config{}, nil
}

func (cfg Config) Query(path string) (QueryResult, error) {
	parts := strings.Split(path, " ")

	curNode := &cfg.Data
	for _, v := range parts {
		//TODO
	}
}

func (cfg Config) QuerySingle(path string) (string, error) {
	prop, err := cfg.Query(path)
	return "", err
}
