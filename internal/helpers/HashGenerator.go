package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"golang.org/x/sync/errgroup"
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

func SetHash[T HashedObject](key string, datalist []T) error {
	if key != "" {
		g := &errgroup.Group{}
		for _, v := range datalist {
			gen := HashGenerator{Object: v, Key: key}
			g.Go(gen.Do)
		}
		err := g.Wait()
		if err != nil {
			log.Printf("could not set hash  %v", err)
			return err
		}
	}
	return nil
}

type HashedObject interface {
	SetHash(hash string)
}
