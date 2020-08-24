package msg

func init() {
	Processor.Register(&C2S_CreateOrderSuccess{})
}

const (
	ErrPaySuccess  = 0 //成功
	ErrPayFail     = 1 //我们的错
	ErrPayBusiness = 2 //他们的错
)

type S2C_PayOK struct {
	Error     int
	AddCoupon int
}

type C2S_CreateEdyOrder struct {
	PriceID int
	//DefPayType string
}

const (
	ErrCreateEdyOrderSuccess     = 0 //成功
	ErrCreateEdyOrderNotRealAuth = 1 //未实名
	ErrCreateEdyOrderOverRange   = 2 //超出范围
)

type S2C_CreateEdyOrder struct {
	Error    int
	ErrMsg   string
	AppID    int
	AppToken string
	Amount   int
	PayType  int
	//DefPayType	string
	Subject          string
	Description      string
	OpenOrderID      string
	OpenNotifyUrl    string
	CreatePaymentUrl string
}

type C2S_CreateOrderSuccess struct {
	OrderID     string
	OpenOrderID string
}
