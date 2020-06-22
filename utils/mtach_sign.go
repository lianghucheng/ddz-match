package utils

import (
	"crypto/md5"
	"sort"
	"strings"
)

func GenerateEdySign(params []string, secret string) []byte {
	sort.Strings(params)
	str := ``
	for _, v := range params {
		temp := strings.Split(v, "=")
		if len(temp) == 2 {
			str += temp[0] + temp[1]
		}
	}
	str += secret
	m := md5.New()
	m.Write([]byte(strings.ToUpper(str)))
	return m.Sum(nil)
}
