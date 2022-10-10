package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dexterp/ifaces/internal/resources/envs"
	"github.com/dexterp/ifaces/internal/resources/modinfo"
	"github.com/dexterp/ifaces/internal/resources/stringx"
)

// PathToImport determine import from path name. Requires paths that are
// descendants of GOROOT or GOPATH, or that an ancestor director contains a
// go.mod file.
func PathToImport(path string) (imp string, err error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ``, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return ``, err
	}

	goRootSrc := filepath.Join(envs.Goroot(), "src")
	if strings.HasPrefix(abs, goRootSrc) {
		if info.IsDir() {
			imp = abs
		} else {
			imp = filepath.Dir(abs)
		}
		imp = strings.TrimPrefix(imp, goRootSrc+string(os.PathSeparator))
		if imp == `` {
			return ``, fmt.Errorf(`invalid go source path: %s`, path)
		}
		if os.PathSeparator != rune('/') {
			imp = strings.ReplaceAll(imp, string(os.PathSeparator), `/`)
		}
		return imp, nil
	}

	goPathSrc := filepath.Join(envs.Gopath(), `pkg`, `mod`)
	if strings.HasPrefix(abs, goPathSrc) {
		if !info.IsDir() {
			imp = filepath.Dir(abs)
		} else {
			imp = abs
		}
		imp = strings.TrimPrefix(imp, goPathSrc+string(os.PathSeparator))
		if os.PathSeparator != rune('/') {
			imp = strings.ReplaceAll(imp, string(os.PathSeparator), `/`)
		}
		imp = stringx.StripVersion(imp)
		if imp == `` {
			return ``, fmt.Errorf(`invalid go source path: %s`, path)
		}
		return imp, nil
	}

	return modinfo.GetImport(``, nil, path)
}
