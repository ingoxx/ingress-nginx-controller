package file

import (
	"crypto/sha1"
	"encoding/hex"
	"k8s.io/klog/v2"
	"os"
)

// SHA1 returns the SHA1 of a file.
func SHA1(filename string) string {
	hasher := sha1.New() // #nosec
	s, err := os.ReadFile(filename)
	if err != nil {
		klog.ErrorS(err, "Error reading file", "path", filename)
		return ""
	}

	hasher.Write(s)
	return hex.EncodeToString(hasher.Sum(nil))
}
