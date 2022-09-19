package modinfo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// GetImport returns the import string for srcpath by combining the path of the
// go.mod file with the go source path to generate the import path. The contents
// of the go.mod file are passed in as the byte array. If modpath and/or modpath
// is empty, the parent directories of srcpath are scanned for the go.mod file.
func GetImport(mod []byte, modpath, srcpath string) (importpath string, err error) {
	if srcpath == `` {
		return ``, errors.New(`no srcpath in arguments`)
	}
	mod, modpath, err = getModInputs(mod, modpath, srcpath)
	if err != nil {
		return importpath, err
	}
	p := modfile.ModulePath(mod)
	parentpath := strings.TrimSuffix(modpath, `go.mod`)
	realpath := strings.Split(filepath.Dir(strings.TrimPrefix(srcpath, parentpath)), string(os.PathSeparator))
	return strings.Join(append([]string{p}, realpath...), "/"), nil
}

func getModInputs(mod []byte, modpath, srcpath string) ([]byte, string, error) {
	var (
		err error
	)
	if mod != nil && modpath != `` {
		return mod, modpath, nil
	}
	if modpath == `` {
		modpath, err = findMod(srcpath)
		if err != nil {
			return nil, ``, err
		}
	}
	if mod == nil {
		mod, err = os.ReadFile(modpath)
		if err != nil {
			return nil, ``, err
		}
	}
	return mod, modpath, nil
}

func findMod(srcpath string) (modpath string, err error) {
	d := filepath.Dir(srcpath)
	if d == srcpath || d == `` {
		return ``, errors.New(`no go.mod file found in parent directory`)
	}
	modpath = filepath.Join(d, `go.mod`)
	if _, err := os.Stat(modpath); os.IsNotExist(err) {
		return findMod(d)
	} else if err != nil {
		return ``, err
	}
	return modpath, nil
}
