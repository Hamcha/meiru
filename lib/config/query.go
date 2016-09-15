package config

import (
	"errors"
	"strconv"
	"strings"
)

var (
	QueryErrInvalidParamConstraint = errors.New("query cfg error: invalid param constraint")
	QueryErrSingleNonNumFilter     = errors.New("query cfg error: non numeric single filter")
	QueryErrSingleTooFewResults    = errors.New("query cfg error: too few results")
	QueryErrSingleTooFewValues     = errors.New("query cfg error: too few values")
)

type QueryResult []Property

func (cfg Config) Query(path string) (QueryResult, error) {
	parts := strings.Split(path, " ")
	return queryPath(parts, cfg.Data)
}

func (cfg Config) QuerySingle(path string) (string, error) {
	// Separate between generic and specific part
	sep := strings.LastIndexByte(path, ' ')

	// Call query on generic path
	results, err := cfg.Query(path[:sep])
	if err != nil {
		return "", err
	}

	//
	// Parse specific path
	//

	parts := strings.SplitN(path[sep+1:], ":", 2)
	resultID := 0
	paramID := 0
	if len(parts) == 1 {
		// "n" will be the Nth value of the first result
		paramID, err = strconv.Atoi(parts[0])
		if err != nil {
			return "", QueryErrSingleNonNumFilter
		}
	} else {
		// "n:m" will the be the Mth value of the Nth result
		resultID, err = strconv.Atoi(parts[0])
		if err != nil {
			return "", QueryErrSingleNonNumFilter
		}
		paramID, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", QueryErrSingleNonNumFilter
		}
	}

	//
	// Filter through results using the specific path
	//

	// Check for out of bound errors
	if resultID >= len(results) {
		return "", QueryErrSingleTooFewResults
	}
	if paramID >= len(results[resultID].Values) {
		return "", QueryErrSingleTooFewValues
	}

	return results[resultID].Values[paramID], nil
}

type constraint struct {
	ParamID int
	Value   string
}

func queryPath(path []string, block Block) ([]Property, error) {
	var found []Property

	if len(path) == 1 {
		// Already at leaf nodes, only fetch matching
		for _, property := range block {
			isOk, err := verifyConstraints(path[0], property)
			if err != nil {
				return found, err
			}
			if isOk {
				found = append(found, property)
			}
		}
	} else {
		// Root or middle nodes, recurse to matching leaves
		for _, property := range block {
			if property.Block != nil {
				isOk, err := verifyConstraints(path[0], property)
				if err != nil {
					return found, err
				}
				if isOk {
					items, err := queryPath(path[1:], property.Block)
					if err != nil {
						return found, err
					}

					found = append(found, items...)
				}
			}
		}
	}

	return found, nil
}

func getConstraintList(str string) ([]constraint, error) {
	var list []constraint

	// Add query constraint terminator if not present
	if !strings.HasSuffix(str, ",") {
		str += ","
	}

	// Iterate through string to find constraints
	currentConstraint := constraint{}
	firstChar := 0
	trimQuotes := false
	insideQuotes := false
	for i, chr := range str {
		switch chr {
		case '"':
			trimQuotes = true
			if i > 0 && str[i-1] != '\\' {
				insideQuotes = !insideQuotes
			}
		case ',':
			if insideQuotes {
				continue
			}

			currentConstraint.Value = str[firstChar:i]

			if trimQuotes {
				currentConstraint.Value = strings.Trim(currentConstraint.Value, "\"")
			}

			list = append(list, currentConstraint)
		case '=':
			if insideQuotes {
				continue
			}

			// Get param id
			num, err := strconv.Atoi(str[firstChar:i])
			if err != nil {
				return list, QueryErrInvalidParamConstraint
			}

			currentConstraint.ParamID = num
			firstChar = i + 1
			trimQuotes = false
		}
	}

	return list, nil
}

func verifyConstraints(path string, property Property) (bool, error) {
	parts := strings.SplitN(path, ":", 2)
	pathname := parts[0]

	// Parse constraints if any
	var constraints []constraint
	if len(parts) > 1 {
		var err error
		constraints, err = getConstraintList(parts[1])
		if err != nil {
			return false, err
		}
	}

	// Check path
	if pathname != property.Key {
		return false, nil
	}

	// Check constraints
	numvals := len(property.Values)
	for _, constraint := range constraints {
		if constraint.ParamID >= numvals || property.Values[constraint.ParamID] != constraint.Value {
			return false, nil
		}
	}

	return true, nil
}
