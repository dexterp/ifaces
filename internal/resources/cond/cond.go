// cond conditionals package
package cond

import "reflect"

// First return the first non empty value. This values are:
//
//   - string: any string with a length greater then 1
//   - numeric: any non 0 value
//   - bool: true values
//   - anything else that is not nil
//
// If no non-zero values are found, the last variadic variable is returned.
func First(a ...any) any {
	l := len(a) - 1
	for i, y := range a {
		end := i == l
		switch v := y.(type) {
		case nil:
			if end {
				return nil
			}
			continue
		case string:
			if v != `` {
				return y
			} else if end {
				return ``
			}
			continue
		case int, uint8, uint16, uint32, uint64, int8, int16, int32, int64, float32, float64:
			if v != 0 {
				return y
			} else if end {
				return 0
			}
			continue
		case bool:
			if v {
				return y
			} else if end {
				return false
			}
			continue
		}
		return y
	}
	return nil
}

func StringValPos(val any, pos int, strs []string) bool {
	for i, v := range strs {
		if val == v && pos == i+1 {
			return true
		}
	}
	return false
}

func EqualAny(target any, a ...any) bool {
	for _, x := range a {
		if reflect.DeepEqual(target, x) {
			return true
		}
	}
	return false
}

func EqualAnyString(s string, eq ...string) bool {
	for _, x := range eq {
		if s == x {
			return true
		}
	}
	return false
}
