package myutils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func GenerateCacheKey(obj any) string {
	if obj == nil {
		return md5Hash("null")
	}

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return md5Hash("null")
	}

	return md5Hash(string(jsonBytes))
}

func md5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}
