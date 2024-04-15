package utils

import "strconv"

func IsInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func HasDuplicates(arr []int) bool {
	counter := make(map[int]int)
	for _, el := range arr {
		counter[el]++
		if counter[el] > 1 {
			return true
		}
	}
	return false
}
