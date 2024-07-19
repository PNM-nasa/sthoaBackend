package random

import "math/rand"

func Createtoken16() string {
	lenToken := 16 // no int
	bytes := make([]byte, lenToken)
	for i := 0; i < lenToken; i++ {
		bytes[i] = byte('a' + rand.Intn('z'-'a'))
	}
	return string(bytes)
}
