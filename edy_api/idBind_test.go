package edy_api

import (
	"testing"
)

func TestIDBindReq_IDCardBind(t *testing.T) {
	idBind := NewIDBindReq(1,"", "", "")
	if err := idBind.IDCardBind(); err != nil {
		panic(err)
	}
}