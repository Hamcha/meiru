package config

import "strings"

type QueryResult []Property

func queryPath(path []string, block Block) []Property {
	var found []Property

	if len(path) == 1 {
		// Already at leaf nodes, only fetch matching
		for _, property := range block {
			if property.Key == path[0] {
				found = append(found, property)
			}
		}
	} else {
		// Root or middle nodes, recurse to matching leaves
		for pathIndex, pathName := range path {
			for _, property := range block {
				if property.Key == pathName && property.Block != nil {
					found = append(found, queryPath(path[pathIndex+1:], property.Block)...)
				}
			}
		}
	}

	return found
}

func (cfg Config) Query(path string) QueryResult {
	parts := strings.Split(path, " ")
	return queryPath(parts, cfg.Data)
}
