package gps

import (
	"errors"
	"strconv"
)

func stringToInt(s string) (int, error) {
	N, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.New("bad integer")
	}
	return N, nil
}


func stringToFloat(s string) (float64, error) {
	N, err := strconv.ParseFloat(s,64)
	if err != nil {
		return 0, err
	}
	return N, nil
}
