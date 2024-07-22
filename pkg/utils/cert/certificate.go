package utils

import (
	"encoding/base64"
)

func DecodeBase64(data map[string]byte) (map[string]string, error) {
	var parsed = make(map[string]string)
	for k, v := range data {
		decodeBase64, err := base64.StdEncoding.DecodeString(string(v))
		if err != nil {
			parsed[k] = ""
		} else {
			parsed[k] = string(decodeBase64)
		}
	}
	return parsed, nil
}
