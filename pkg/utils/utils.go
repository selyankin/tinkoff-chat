package utils



func Unique(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func ArrayIn(strSlice []string, element string) bool{
	for _, e := range strSlice{
		if e == element{
			return true
		}
	}
	return false
}
