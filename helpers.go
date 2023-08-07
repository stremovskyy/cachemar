package cachemar

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func HashKey(prefix string, object interface{}) string {
	str := fmt.Sprintf("%v", object)
	hash := md5.Sum([]byte(str))
	hashStr := hex.EncodeToString(hash[:])

	return prefix + ":" + hashStr
}
