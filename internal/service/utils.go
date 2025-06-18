package service

import (
	"math/rand"
	"time"
)

const (
	lenShortLink int = 8
)

var pseudoRand = rand.New(rand.NewSource(time.Now().Unix()))

// create short link with pseudo-random.
//
// link with len 8 symbols
func generateShortLink() string {
	shortLink := make([]byte, lenShortLink)
	for i := range shortLink {
		if pseudoRand.Intn(2) == 0 {
			shortLink[i] = byte(pseudoRand.Intn(25) + 65)
		} else {
			shortLink[i] = byte(pseudoRand.Intn(25) + 97)
		}
	}
	return string(shortLink)
}
