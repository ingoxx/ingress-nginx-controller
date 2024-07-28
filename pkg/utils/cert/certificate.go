package utils

func DecodeBase64(data map[string][]byte) (map[string][]byte, error) {
	var parsed = make(map[string][]byte)
	for k, v := range data {
		parsed[k] = v
	}
	return parsed, nil
}
