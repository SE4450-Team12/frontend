package main

import (
	"fmt"
	"encoding/base64"
	"encoding/hex"
	"crypto/sha256"
	"hash"
	"io"
	"golang.org/x/crypto/hkdf"
)

func main() {
	secret := []byte{0x00, 0x01, 0x02, 0x03} // this should be the root key?
	salt := []byte("salt") // this should be some shared constant?
	info := []byte("hskdf example") // this should be the dh output?

	keys := hkdf_output(2, 32, sha256.New, secret, salt, info)

	for i := range keys {
		fmt.Println(keys[i]) //byte array
		fmt.Println(b16(keys[i])) //hex representation
		fmt.Println(b64(keys[i])) //base64 representation
	}

}

func hkdf_output (numKeys int, keySize int, hash func() hash.Hash, secret, salt, info []byte) [][]byte{

	hkdf_step :=hkdf.New(hash, secret, salt, info)

	var keys [][]byte
	for i := 0; i < numKeys; i++ {
		key := make([]byte, keySize)
		if _, err := io.ReadFull(hkdf_step, key); err != nil {
			panic(err)
		}
		keys = append(keys, key)
	}

	return keys
}

func b16(msg []byte) string {
	return hex.EncodeToString(msg)
}
func b64(msg []byte) string {
	return base64.StdEncoding.EncodeToString(msg)
}

