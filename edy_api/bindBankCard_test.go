package edy_api

import "testing"

func TestBindBankCardReq_BindBankCard(t *testing.T) {
	bindBankCard := NewBindBankCardReq(1, "xxx", "平安银行", "6230580000155032654")
	if err := bindBankCard.BindBankCard(); err != nil {
		panic(err)
	}
}
