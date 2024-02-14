package main

import (
	"hash"
)

type Encoder interface {
	Encode(str []byte, hash []byte) []byte
}

type Sha256Encoder struct {
	Hash hash.Hash
	Encoder
}

func (encoder *Sha256Encoder) Encode(input []byte, hash []byte) []byte {
	defer encoder.Hash.Reset()

	encoder.Hash.Write(input)
	return encoder.Hash.Sum(hash)
}
