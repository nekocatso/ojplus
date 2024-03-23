package util

import "strconv"

func isInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
