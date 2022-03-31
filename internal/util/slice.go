package util

func find(s []string, target string) int {
	for i, str := range s {
		if str == target {
			return i
		}
	}
	return -1
}

func Remove(origin []string, target string) []string {
	length := len(origin)
	i := find(origin, target)
	if i < 0 {
		return origin
	}

	if i == 0 {
		return origin[1:]
	} else if i == length-1 {
		return origin[:i]
	} else {
		return append(origin[:i], origin[i+1:]...)
	}
}
