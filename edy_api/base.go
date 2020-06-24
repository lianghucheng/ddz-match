package edy_api

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/szxby/tools/log"
)

const (
	modelSandbox = 0
	modelProduct = 1
)

var (
	model      = modelSandbox
	sandboxUrl = "https://sandbox-api-cc.cmsa.cn:60001"
	productUrl = ""
	url        string
	cp_id      = "17101" //厂商id
	secret     = "USWH1TDG8K5G5C72N64JP4P6DDC1QDEF"
)

func init() {
	LoadSource()
}

type MyClient struct {
	http.Client
	Uri       string
	Param     string
	TimeStamp int64 //13位时间戳
	SignCode  string
}

func LoadSource() {
	switch model {
	case modelSandbox:
		url = sandboxUrl
	case modelProduct:
		url = productUrl
	}
}

func NewClient(uri, param string, reqType int) *MyClient {
	c := new(MyClient)
	switch model {
	case modelSandbox:
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c.Transport = tr
	case modelProduct:

	}
	c.Uri = uri
	c.TimeStamp = int64(time.Now().UnixNano() / 1e6)
	if reqType == reqPost {
		c.Param = param
	} else if reqType == reqGet {
		c.Param = param + "&timestamp=" + fmt.Sprintf("%v", c.TimeStamp)
	}

	return c
}

func (ctx *MyClient) DoGet() ([]byte, error) {
	if ctx.SignCode == "" {
		log.Error("no generate sign code. ")
		return nil, errors.New("no generate sign code. ")
	}
	resp, err := ctx.Get(url + ctx.Uri + "?" + ctx.Param + "&sign=" + ctx.SignCode)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return b, nil
}

func (ctx *MyClient) DoPost() ([]byte, error) {
	if ctx.SignCode == "" {
		log.Error("no generate sign code. ")
		return nil, errors.New("no generate sign code. ")
	}

	req, err := http.NewRequest("POST", url+ctx.Uri, bytes.NewBuffer([]byte(ctx.Param)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("timestamp", fmt.Sprintf("%v", ctx.TimeStamp))
	req.Header.Set("sign", ctx.SignCode)

	resp, err := ctx.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return b, nil
}

const (
	reqGet  = 1
	reqPost = 2
)

func (ctx *MyClient) GenerateSign(signType int) {
	str := ""
	if signType == reqPost {
		str = ctx.Uri + ctx.Param + fmt.Sprintf("%v", ctx.TimeStamp) + secret
	} else if signType == reqGet {
		temp := strings.Split(ctx.Param, "&")
		sort.Strings(temp)
		for _, v := range temp {
			t := strings.Split(v, "=")
			for _, v2 := range t {
				str += v2
			}
		}
		str += fmt.Sprintf("%v", ctx.TimeStamp) + secret
	}
	log.Debug("生成签名之前的字符串：%v", str)

	m := md5.New()
	m.Write([]byte(str))
	ctx.SignCode = strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
}
