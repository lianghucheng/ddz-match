package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	. "reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/szxby/tools/log"
)

var (
	chars = []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
		"u", "v", "w", "x", "y", "z", "A", "B", "C", "D",
		"E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
		"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
		"Y", "Z", "~", "!", "@", "#", "$", "%", "^", "&",
		"*", "(", ")", "-", "_", "=", "+", "[", "]", "{",
		"}", "|", "<", ">", "?", "/", ".", ",", ";", ":"}

	numberChars = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
)

func Shuffle2(a []string) []string {
	i := len(a) - 1
	for i > 0 {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
		i--
	}
	return a
}

func GetToken(n int) string {
	if n < 1 {
		return ""
	}
	var tokens []string
	for i := 0; i < n; i++ {
		tokens = append(tokens, chars[rand.Intn(90)]) // 90 是 Chars 的长度
	}
	return strings.Join(tokens, "")
}

// id 的第一位从 1 开始
func GetID(n int) int {
	if n < 1 {
		return -1
	}
	min := math.Pow10(n - 1)
	id := int(min) + rand.Intn(int(math.Pow10(n)-min))
	return id
}

func HttpPost(url string, data string) ([]byte, error) {
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		return body, nil
	}
	return nil, err
}
func Atoi(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

var todayCode = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C",
	"D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P",
	"Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

func GetTodayCode(n int) string {
	newWords := ""
	for i := 0; i < n; i++ {
		newWords += todayCode[rand.Intn(len(todayCode))]
	}
	return newWords
}

func OneDay0ClockTimestamp(t time.Time) int64 {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func TimeFormat() string {
	return time.Now().Format("20060102")
}

// 验证是否手机
func PhoneRegexp(phone string) bool {
	b := false
	if phone != "" {
		reg := regexp.MustCompile(`^(86)*0*1\d{10}$`)
		b = reg.FindString(phone) != ""
	}
	return b
}

var numbers = []rune("0123456789")

func RandomNumber(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = numbers[rand.Intn(10)]
	}
	return string(b)
}

// StructCopy 对结构体中相同字段进行拷贝
func StructCopy(dst, src interface{}) {
	// log.Debugf("%#v %#v", dst, src)
	if dst == nil || src == nil {
		log.Error("nil value")
		return
	}

	srcVal := Indirect(ValueOf(src))
	// log.Debug("src:%v", srcVal)
	dstVal := Indirect(ValueOf(dst))
	// log.Debug("dst:%v", dstVal)
	// log.Debugf("srcVal.Kind():%v,dstVal.Kind():%v", srcVal.Kind(), dstVal.Kind())
	if !(srcVal.Kind() == Struct && dstVal.Kind() == Struct) {
		// log.Error("type is not struct ptr,src:%v,dst:%v", srcVal.Kind(), dstVal.Kind())
		return
	}

	srcType := srcVal.Type()
	// dstType := dstVal.Type()
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		name := srcType.Field(i).Name
		dstField := dstVal.FieldByName(name)
		if !dstField.CanSet() {
			continue
		}

		if getKind(srcField) != getKind(dstField) {
			continue
		}
		switch getKind(srcField) {
		case Int64:
			dstField.SetInt(srcField.Int())
		case Uint64:
			dstField.SetUint(srcField.Uint())
		case Float64:
			dstField.SetFloat(srcField.Float())
		case Bool, String, Slice:
			dstField.Set(srcField)
		case Ptr:
			StructCopy(dstField.Interface(), srcField.Interface())
		default:
			// log.Infof("%#v %#v", srcField, dstField)
		}
	}
}

func getKind(v Value) Kind {
	kind := v.Kind()
	switch kind {
	case Int, Int8, Int16, Int32, Int64:
		return Int64
	case Uint, Uint8, Uint16, Uint32, Uint64:
		return Uint64
	case Float32, Float64:
		return Float64
	default:
		return kind
	}
}

// CalculateHash calculate hash
func CalculateHash(data string) string {
	h := sha256.New()
	h.Write([]byte(key + data))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// CalculateMD5 calculate md5
func CalculateMD5(data string) string {
	h := md5.New()
	h.Write([]byte(key + data))
	return hex.EncodeToString(h.Sum(nil))
}

// FormatFloat 取小数点后n位非零小数
func FormatFloat(num float64, decimal int) string {
	if math.Trunc(num) == num || decimal == 0 {
		return fmt.Sprintf("%.f", math.Trunc(num))
	}
	format := "%." + strconv.Itoa(decimal) + "f"
	return fmt.Sprintf(format, num)
}

// RoundFloat 取小数点后n位非零小数,四舍五入
func RoundFloat(num float64, decimal int) string {
	// if math.Trunc(num) == num || decimal == 0 {
	// 	return fmt.Sprintf("%.f", math.Trunc(num))
	// }
	for i := 0; i < decimal; i++ {
		num = num * 10
	}
	return fmt.Sprintf("%.1f", math.Round(num)/10)

	// format := "%." + strconv.Itoa(decimal) + "f"
	// return fmt.Sprintf(format, num)
}

// GetFirstDateOfMonth 获取本月第一天零点
func GetFirstDateOfMonth(d time.Time) time.Time {
	d = d.AddDate(0, 0, -d.Day()+1)
	return GetZeroTime(d)
}

// GetLastDateOfMonth 获取本月最后一天零点
func GetLastDateOfMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 1, -1)
}

// GetZeroTime 获取某一天的0点时间
func GetZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func Transfer(src interface{}, dir interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if err := json.Unmarshal(b, dir); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func TransferMapStringSlice(data interface{}) ([]map[string]string, error) {
	rt := []map[string]string{}
	b1, err := json.Marshal(data)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(b1, &m); err != nil {
		log.Debug(err.Error())
		err.Error()
		return nil, err
	}
	for k, v := range m {
		temp := make(map[string]string)
		temp[k] = fmt.Sprintf("%v", v)
		rt = append(rt, temp)
	}

	return rt, nil
}

func ParseJsonParam(req *http.Request, rt interface{}) (code int, desc string) {
	code = Success
	desc = ErrMsg[code]
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err.Error())
		code = Fail
		desc = ErrMsg[code]
		return
	}
	log.Debug("【接收到的参数】%v", string(data))
	if err := json.Unmarshal(data, rt); err != nil {
		log.Error(err.Error())
		code = FormatFail
		desc = ErrMsg[code]
		return
	}
	return
}
