package modinfo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/dexterp/ifaces/internal/resources/envs"
	"github.com/dexterp/ifaces/internal/resources/match"
	"golang.org/x/mod/modfile"
)

var ErrNotFound = errors.New(`module not found`)
var ErrLatestNotSupported = errors.New(`@latest version currently not supported`)

type ModInfo struct {
	modfile *modfile.File
}

func Load(file string, data []byte) (*ModInfo, error) {
	f, err := modfile.Parse(file, data, nil)
	if err != nil {
		return nil, err
	}
	return &ModInfo{
		modfile: f,
	}, nil
}

func LoadFromParents(srcpath string) (*ModInfo, error) {
	modpath, err := findMod(srcpath)
	if err != nil {
		return nil, err
	}
	mod, err := os.ReadFile(modpath)
	if err != nil {
		return nil, err
	}
	return Load(modpath, mod)
}

func (m ModInfo) GetVersion(module string) (ver string, err error) {
	for _, r := range m.modfile.Require {
		if match.Match(r.Mod.Path, module) {
			return r.Mod.Version, nil
		}
	}
	return ``, ErrNotFound
}

// GetPath obtains the package path to a file path on disk if it exists. The
// module name is passed in as an argument and the path to the parsed source
// file.
func (m ModInfo) GetPath(mod string) (realPath string, err error) {
	switch {
	case strings.HasSuffix(mod, `@latest`):
		return ``, errors.New(`@latest version currently not supported`)
	case strings.Contains(mod, `@`):
		realPath = filepath.Join(envs.Gopath(), `pkg`, `mod`, filepath.FromSlash(mod))
	case m.modfile != nil:
		v, err := m.GetVersion(mod)
		if err != nil {
			return ``, err
		}
		if v == `` {
			return ``, ErrNotFound
		}
		realPath = filepath.Join(envs.Gopath(), `pkg`, `mod`, filepath.FromSlash(mod)+`@`+v)
	default:
		return ``, ErrNotFound
	}
	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		return ``, ErrNotFound
	}
	return
}

// GetImport returns the import string for srcpath by combining the path of the
// go.mod file with the go source path to generate the import path. The contents
// of the go.mod file are passed in as the byte array. If gomodpath and/or data
// is empty, the parent directories of srcpath are scanned for the go.mod file.
func GetImport(gomodpath string, data []byte, srcpath string) (importpath string, err error) {
	if srcpath == `` {
		return ``, errors.New(`no srcpath in arguments`)
	}
	data, gomodpath, err = getModInputs(data, gomodpath, srcpath)
	if err != nil {
		return importpath, err
	}
	p := modfile.ModulePath(data)
	parentpath := strings.TrimSuffix(gomodpath, `go.mod`)
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
	modpath = filepath.Join(srcpath, `go.mod`)
	if _, err := os.Stat(modpath); err != nil {
		newpath := filepath.Dir(srcpath)
		if newpath == srcpath {
			return ``, errors.New(`no go.mod file found in parent directory`)
		}
		return findMod(newpath)
	} else if err != nil {
		return ``, err
	}
	return modpath, nil
}
