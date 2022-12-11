package algs

func FirstNotEmpty(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}

	return ""
}
