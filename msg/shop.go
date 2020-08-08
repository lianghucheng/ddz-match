package msg

func init() {
	Processor.Register(&S2C_PayAccount{})
}

type S2C_PayAccount struct {
	Accounts []string
}

type GoodsType struct {
	ID int//唯一标识
	TypeName string//商品名称
	ImgUrl string//商品图标
	PriceItems []PriceItem
}

type PriceItem struct {
	PriceID int
	Fee     int64
	Name    string
	Amount  int
	ImgUrl string//商品图标
	TakenType int//花费类型。1是RMB
	//PropType int//道具类型。1是点券
	GiftAmount int
}
type S2C_PriceMenu struct {
	PriceItems []GoodsType
}
