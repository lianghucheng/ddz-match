package http

import (
	"bytes"
	"net/url"
	"sort"
)

/*
GET请求（需要在请求中加上Unix时间戳 【如timestamp=1490600017808】和加密字段 【如sign=3F2278956919ED7A256E6690ABD5A58E】）
对请求参数根据参数名称
（除去sign参数外）的顺序排序。如：
foo=1, bar=2, foo_bar=3, foobar=4,timestamp=1490600017808
排序后的顺序是bar=2, foo=1, foo_bar=3, foobar=4,timestamp=1490600017808。
·将排序好的参数名和参数值拼装在一起，根据上面的示例得到的结果为：bar2foo1foo_bar3foobar4。
·把拼装好的字符串采用utf-8编码，使用MD5算法，需要在拼装的字符串前后加上 secret后加密，
如：md5(bar2foo1foo_bar3foobar4timestamp1490600017808+secret)。加密后需要把加密字段转为大写。
*/
func GetSignContent1(v url.Values) string {
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		buf.WriteString(k)
		for _, v := range vs {

			buf.WriteString(v)
		}
	}
	return buf.String()
}
