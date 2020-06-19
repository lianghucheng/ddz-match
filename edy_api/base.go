package edy_api

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
)

const (
	modelSandbox = 0
	modelProduct = 1
)

var (
	model = modelSandbox
	sandboxUrl = "https://sandbox-api-cc.cmsa.cn:60001"
	productUrl = ""
	url string
	cp_id = "17101" //厂商id
	secret = "USWH1TDG8K5G5C72N64JP4P6DDC1QDEF"
)

func init() {
	LoadSource()
}

func LoadSource() {
	switch model {
	case modelSandbox:
		url = sandboxUrl
	case modelProduct:
		url = productUrl
	}
}

func newClient() *http.Client {
	c := new(http.Client)
	switch model {
	case modelSandbox:
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c.Transport =tr
	case modelProduct:

	}

	return c
}

func Get(uri, values string) ([]byte, error) {
	client := newClient()
	resp,err := client.Get(url+uri+"?"+values)
	if err != nil {
		return nil, err
	}
	b, err :=ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return b, nil
}

func Post(uri, value string) ([]byte, error ) {
	client := newClient()
	req, err := http.NewRequest("POST", url + uri, bytes.NewBuffer([]byte(value)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
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
