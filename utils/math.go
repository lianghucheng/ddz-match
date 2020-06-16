package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
)

func MD5Encrypt(data string) string {
	m := md5.New()
	m.Write([]byte(data))
	return hex.EncodeToString(m.Sum(nil))
}

func MD5Encrypt2(data string) string {
	m := md5.New()
	io.WriteString(m, data)
	return fmt.Sprintf("%X", m.Sum(nil))
}

// 四舍五入，保留n位小数
func Round(f float64, n int) float64 {
	pow10N := math.Pow10(n)
	return math.Trunc((f+0.5/pow10N)*pow10N) / pow10N
}
