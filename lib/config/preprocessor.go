package config

import (
	"errors"
	"path/filepath"
)

type pScope map[string]string
type pFunction func(scope pScope, prop Property) ([]Property, error)

var (
	PPErrorInexistantFunction = errors.New("preprocess cfg error: unknown preprocess directive")
	PPErrorMissingParameter   = errors.New("preprocess cfg error: missing required parameter")
)

func processConfig(path string, block Block) (Block, error) {
	scope := pScope{
		"_pwd": filepath.Dir(path),
	}

	var out Block
	for _, property := range block {
		if property.Key[0] == '@' {
			var function pFunction
			switch property.Key[1:] {
			case "include":
				function = pInclude
			default:
				return out, PPErrorInexistantFunction
			}
			result, err := function(scope, property)
			if err != nil {
				return out, err
			}
			out = append(out, result...)
		} else {
			out = append(out, property)
		}
	}

	return out, nil
}

func pInclude(scope pScope, prop Property) ([]Property, error) {
	if len(prop.Values) < 1 {
		return nil, PPErrorMissingParameter
	}

	var props []Property
	for _, val := range prop.Values {
		path := filepath.Clean(filepath.Join(scope["_pwd"], val))
		cfg, err := LoadConfig(path)
		if err != nil {
			return nil, err
		}
		props = append(props, cfg.Data...)
	}

	return props, nil
}
