package generator

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"
)

func GenerateUniqueID(fromID, toID string, amount float64) string {

	data := fromID + toID + strconv.FormatFloat(amount, 'f', -1, 64) + time.Now().String()

	hash := md5.Sum([]byte(data))

	uniqueID := hex.EncodeToString(hash[:])

	return uniqueID
}
