package oauth2

import (
	"fmt"
	"strconv"
	"strings"
)

type hashRule struct {
	algorithm  string
	iterations int
	length     int
	salt       string
	value      string
}

func extractHashRule(hash string) (*hashRule, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid hashed password format")
	}
	subParts := strings.Split(parts[2], ",")
	if len(subParts) != 2 {
		return nil, fmt.Errorf("invalid hashed sub part format")
	}
	iterationsIndex := strings.Index(subParts[0], "=")
	if iterationsIndex == -1 {
		return nil, fmt.Errorf("iterations not found")
	}
	iterations, err := strconv.Atoi(subParts[0][iterationsIndex+1:])
	if err != nil {
		return nil, fmt.Errorf("invalid iterations format")
	}
	lengthIndex := strings.Index(subParts[1], "=")
	if lengthIndex == -1 {
		return nil, fmt.Errorf("length not found")
	}
	length, err := strconv.Atoi(subParts[1][lengthIndex+1:])
	if err != nil {
		return nil, fmt.Errorf("invalid length format")
	}
	return &hashRule{
		algorithm:  parts[0],
		iterations: iterations,
		length:     length,
		salt:       parts[3],
		value:      parts[4],
	}, nil
}
