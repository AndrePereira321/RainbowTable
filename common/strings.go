package common

import (
	"math/rand"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const alphabetLen = len(alphabet)
const moduloRange = 1 << 16

func ReduceHash(hash []byte, i uint64, seed string, str []byte) {
	data := append(hash, []byte(seed)...)
	iBytes := intToBytes(i)
	length := len(str)
	for j := 0; j < length; j++ {
		for k := 0; k < len(data); k++ {
			data[k] ^= byte(j)
			for iB := range iBytes {
				if iBytes[iB] > 0 {
					data[k] ^= iBytes[iB]
				}
			}
		}
		str[j] = alphabet[customHash(data)%alphabetLen]
	}

}

func customHash(data []byte) int {
	hash := 0
	for _, b := range data {
		hash = (hash*31 + int(b)) % moduloRange
	}
	return hash
}

func intToBytes(i uint64) []byte {
	b := make([]byte, 8)
	for j := 0; j < 8; j++ {
		b[j] = byte(i >> (56 - 8*j))
	}
	return b
}

func GenerateRandomString(min int, max int) []byte {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	length := r.Intn(max-min+1) + min

	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[r.Intn(alphabetLen)]
	}

	return b
}
