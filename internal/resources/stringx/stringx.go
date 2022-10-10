package stringx

import (
	"strings"
	"unicode"
)

// ExIdent extract a valid identifier from the beginning of the string. Returns
// an empty string on failure.
func ExIdent(ident string) string {
	if ident == `` {
		return ``
	}
	out := []rune{}
	r := []rune(ident)
	if unicode.IsLetter(r[0]) {
		out = append(out, r[0])
	} else {
		return ``
	}
	for _, c := range r[1:] {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			out = append(out, c)
		} else {
			return string(out)
		}
	}
	return string(out)
}

// ExPkg extract a valid package name from the beginning of the string. Returns
// an empty string on failure.
func ExPkg(ident string) string {
	if ident == `` {
		return ``
	}
	out := []rune{}
	r := []rune(ident)
	if unicode.IsLower(r[0]) && unicode.IsLetter(r[0]) {
		out = append(out, r[0])
	} else {
		return ``
	}
	for _, c := range r[1:] {
		if (unicode.IsLower(c) && unicode.IsLetter(c)) || unicode.IsDigit(c) {
			out = append(out, c)
		} else {
			return string(out)
		}
	}
	return string(out)
}

// ExPkgPath extract a valid package name from an import path. Returns an empty
// string on failure.
func ExPkgPath(path string) string {
	p := path
	s := strings.Split(path, `/`)
	if l := len(s); l > 0 {
		p = s[l-1]
	}
	return ExPkg(p)
}

// IsIdent true if ident is a valid identifier
func IsIdent(ident string) bool {
	if ident == `` {
		return false
	}
	return ExIdent(ident) == ident
}

// IsPkg true if ident is a valid package name
func IsPkg(ident string) bool {
	if ident == `` {
		return false
	}
	return ExPkg(ident) == ident
}

// NotEmpty return the first non empty string
func NotEmpty(s ...string) (o []string) {
	for _, x := range s {
		if x == `` {
			continue
		}

		o = append(o, x)
	}
	return
}

// StripVersion strip version from paths prefixed with GOPATH
func StripVersion(path string) string {
	var (
		before, after string
	)
	for i, c := range []rune(path) {
		if c == rune('@') {
			before = path[:i]
		}
		if before != `` && c == rune('/') {
			after = path[i:]
			break
		}
	}
	if before != `` {
		return before + after
	}
	return path
}
