package values

type EdyOrder struct {
	ID             int `bson:"_id"`
	Accountid      int
	TradeNo        string
	TradeNoReceive string
	Status         bool
	Fee            int64
	Createdat      int64
}
