package config

import (
	"strconv"
	"strings"

	"github.com/hamcha/meiru/lib/errors"
)

var (
	ErrSrcQuery errors.ErrorSource = "query"

	QueryErrInvalidParamConstraint = errors.NewType(ErrSrcQuery, "invalid param constraint")
	QueryErrSingleNonNumFilter     = errors.NewType(ErrSrcQuery, "non numeric single filter")
	QueryErrSingleTooFewResults    = errors.NewType(ErrSrcQuery, "too few results")
	QueryErrSingleTooFewValues     = errors.NewType(ErrSrcQuery, "too few values")
)

type QueryResult []Property

func (cfg Config) Query(path string) (QueryResult, *errors.Error) {
	return cfg.QuerySub(path, cfg.Data)
}

func (cfg Config) QuerySub(path string, start Block) (QueryResult, *errors.Error) {
	parts := strings.Split(path, " ")
	return queryPath(parts, start)
}

func (cfg Config) QuerySingle(path string) (string, *errors.Error) {
	return cfg.QuerySingleSub(path, cfg.Data)
}

func (cfg Config) QuerySingleSub(path string, start Block) (string, *errors.Error) {
	// Separate between generic and specific part
	sep := strings.LastIndexByte(path, ' ')

	// Call query on generic path
	results, err := cfg.QuerySub(path[:sep], start)
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
		var err error
		paramID, err = strconv.Atoi(parts[0])
		if err != nil {
			return "", errors.NewError(QueryErrSingleNonNumFilter).WithError(err)
		}
	} else {
		// "n:m" will the be the Mth value of the Nth result
		var err error
		resultID, err = strconv.Atoi(parts[0])
		if err != nil {
			return "", errors.NewError(QueryErrSingleNonNumFilter).WithError(err)
		}
		paramID, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", errors.NewError(QueryErrSingleNonNumFilter).WithError(err)
		}
	}

	//
	// Filter through results using the specific path
	//

	// Check for out of bound errors
	if resultID >= len(results) {
		return "", errors.NewError(QueryErrSingleTooFewResults)
	}
	if paramID >= len(results[resultID].Values) {
		return "", errors.NewError(QueryErrSingleTooFewValues)
	}

	return results[resultID].Values[paramID], nil
}

type constraint struct {
	ParamID int
	Value   string
}

func queryPath(path []string, block Block) ([]Property, *errors.Error) {
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

func getConstraintList(str string) ([]constraint, *errors.Error) {
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
				return list, errors.NewError(QueryErrInvalidParamConstraint).WithError(err)
			}

			currentConstraint.ParamID = num
			firstChar = i + 1
			trimQuotes = false
		}
	}

	return list, nil
}

func verifyConstraints(path string, property Property) (bool, *errors.Error) {
	parts := strings.SplitN(path, ":", 2)
	pathname := parts[0]

	// Parse constraints if any
	var constraints []constraint
	if len(parts) > 1 {
		var err *errors.Error
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
