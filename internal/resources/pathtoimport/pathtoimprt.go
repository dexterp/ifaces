package pathtoimport

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dexterp/ifaces/internal/resources/envs"
	"github.com/dexterp/ifaces/internal/resources/modinfo"
)

func PathToImport(path string) (imp string, err error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ``, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return ``, err
	}
	gosrc := filepath.Join(envs.Goroot(), `src`)
	if strings.HasPrefix(abs, gosrc) {
		if !info.IsDir() {
			imp = filepath.Dir(abs)
		} else {
			imp = abs
		}
		imp = strings.TrimPrefix(imp, gosrc)
		if imp == `` {
			return ``, fmt.Errorf(`invalid go source path: %s`, path)
		}
		if os.PathSeparator != rune('/') {
			imp = strings.ReplaceAll(imp, string(os.PathSeparator), `/`)
		}
		return imp, nil
	}

	// TODO - Get paths for go modules that have no go.mod file and are not
	// descendants of GOROOT.

	return modinfo.GetImport(``, nil, path)
}
