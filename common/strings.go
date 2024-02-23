package common

import (
	"math/rand"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const alphabetLen = len(alphabet)

func ReduceHash(hash []byte, chainIndex uint64, seedScore uint64, str []byte) {
	score := chainIndex * seedScore
	strLen := len(str)
	hashLen := len(hash)
	for i := 0; i < strLen; i++ {
		index := ((((uint64(hash[(i+int(chainIndex))%hashLen]) << 8) |
			(uint64(hash[(i+int(chainIndex)+1)%hashLen]))) +
			(score + chainIndex + seedScore + uint64(i))) % 15485863) % uint64(alphabetLen)

		str[i] = alphabet[index]
	}
}

func GenerateRandomString(r *rand.Rand, min int, max int) []byte {
	length := r.Intn(max-min+1) + min

	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[r.Intn(alphabetLen)]
	}

	return b
}
