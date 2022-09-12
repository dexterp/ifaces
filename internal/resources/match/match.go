package match

import "unicode"

// Capitalized check if string is capitalized
func Capitalized(str string) bool {
	if str == `` {
		return false
	}
	return unicode.IsUpper(rune(str[0]))
}

// Match return true if pattern matches str
func Match(str, pattern string) bool {
	var (
		input  = []rune(str)
		pat    = []rune(pattern)
		lenIn  = len(input)
		lenPat = len(pat)
		match  = make([][]bool, lenIn+1)
	)
	for i := range match {
		match[i] = make([]bool, lenPat+1)
	}
	match[0][0] = true
	for i := 1; i < lenIn; i++ {
		match[i][0] = false
	}
	if lenPat > 0 {
		if pat[0] == '*' {
			match[0][1] = true
		}
	}
	for j := 2; j <= lenPat; j++ {
		if pat[j-1] == '*' {
			match[0][j] = match[0][j-1]
		}
	}
	for i := 1; i <= lenIn; i++ {
		for j := 1; j <= lenPat; j++ {
			if pat[j-1] == '*' {
				match[i][j] = match[i-1][j] || match[i][j-1]
			}
			if pat[j-1] == '?' || input[i-1] == pat[j-1] {
				match[i][j] = match[i-1][j-1]
			}
		}
	}
	return match[lenIn][lenPat]
}
