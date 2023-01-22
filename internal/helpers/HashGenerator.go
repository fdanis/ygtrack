package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
)

type HashGenerator struct {
	Metric HashedObject
	Key    string
}

func (hg *HashGenerator) Do() error {
	hash, err := UpdateHash(hg.Metric, hg.Key)
	if err != nil {
		return err
	}
	hg.Metric.SetHash(hash)
	return nil
}

func UpdateHash(metric interface{}, key string) (string, error) {
	result := ""
	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		if _, err := h.Write([]byte(fmt.Sprint(metric))); err != nil {
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
