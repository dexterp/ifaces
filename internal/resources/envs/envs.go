package envs

import (
	"os"
	"strconv"
)

func Gofile() string {
	return os.Getenv("GOFILE")
}

func Goline() int {
	l, err := strconv.Atoi(os.Getenv("GOLINE"))
	if err != nil {
		return -1
	}
	return l
}
