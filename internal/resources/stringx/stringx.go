package stringx

func NotEmpty(s ...string) (o []string) {
	for _, x := range s {
		if x == `` {
			continue
		}

		o = append(o, x)
	}
	return
}
