package values

const (
	PayStatusAction = iota
	PayStatusSuccess
	PayStatusFail
)

const (
	MerchantSportCentralAthketicAssociation = 1
)

var MerchantPay = []int{MerchantSportCentralAthketicAssociation}

const (
	GoodsTypeCoupon = iota
	GoodsTypeCouponFrag
)

type EdyOrder struct {
	ID             int    `bson:"_id"` //唯一标识
	Accountid      int    //用户id
	TradeNo        string //订单号
	TradeNoReceive string //商户订单号
	Status         bool   //订单状态
	Fee            int64  //支付金额
	Createdat      int64  //订单的创建时间和完成时间 todo:一个问题，支付失败和没有支付也需要完成时间？
	PayStatus      int    //0表示支付中， 1表示支付成功， 2表示支付失败
	GoodsType      int    //商品类型。0表示点券，1表示碎片
	Merchant   int    //商户
	Amount         int    //商品数量
}
