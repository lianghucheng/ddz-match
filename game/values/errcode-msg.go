package values

const (
	SuccessS2C_UseProp           = 10000
	ErrS2C_UsePropFail           = 20001
	ErrS2C_UsePropCouponFragLack = 20002
)

var ErrMsg = map[int]string{
	SuccessS2C_UseProp:           "兑换成功",
	ErrS2C_UsePropFail:           "兑换失败",
	ErrS2C_UsePropCouponFragLack: "点券碎片不足，无法兑换",
}
