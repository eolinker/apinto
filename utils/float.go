package utils

import "math/rand"

func RandFloats(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
