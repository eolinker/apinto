package utils

func CopyMaps[K comparable, T any](maps map[K]T) map[K]T {

	temp := make(map[K]T, len(maps))
	for k, t := range maps {
		temp[k] = t
	}

	return temp
}
