// cond conditionals package
package cond

func First(a ...any) any {
	l := len(a)
	for i, y := range a {
		end := i+1 >= 1
		switch v := y.(type) {
		case nil:
			if end {
				return nil
			}
			continue
		case string:
			if v != `` {
				return y
			} else if i+1 >= l {
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

func ContainsString(s string, strs ...string) bool {
	for _, x := range strs {
		if x == s {
			return true
		}
	}
	return false
}
