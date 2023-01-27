package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
)

type HashGenerator struct {
	Object HashedObject
	Key    string
}

func (hg *HashGenerator) Do() error {
	hash, err := GetHash(hg.Object, hg.Key)
	if err != nil {
		return err
	}
	hg.Object.SetHash(hash)
	return nil
}

func GetHash(data any, key string) (string, error) {
	result := ""
	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		if _, err := h.Write([]byte(fmt.Sprint(data))); err != nil {
			log.Println("can not get hash")
			return "", err
		}
		result = hex.EncodeToString(h.Sum(nil))
	}
	return result, nil
}

type HashedObject interface {
	SetHash(hash string)
}
