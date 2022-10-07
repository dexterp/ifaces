package envs

import (
	"go/build"
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

func Gopath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == `` {
		gopath = build.Default.GOPATH
	}
	return gopath
}

func Goroot() string {
	goroot := os.Getenv("GOROOT")
	if goroot == `` {
		goroot = build.Default.GOROOT
	}
	return goroot
}
