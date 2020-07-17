package conf2

type Config2 struct {
	RmbCouponRate int
	NewGiftCoupon int
}

type PriceItem struct {
	PriceID int
	Fee     int64
	Name    string
	Amount  int
}

var dbcfg Config2

//func DBCfgInit() {
//	se := db.MongoDB.Ref()
//	defer db.MongoDB.UnRef(se)
//	if err := se.DB(db.DB).C("conf").Find(nil).One(&dbcfg); err != nil {
//		log.Fatal(err.Error())
//	}
//	log.Debug("【数据库配置文件】%v", dbcfg)
//}

func GetCouponRate() int {
	return dbcfg.RmbCouponRate
}

func GetGiftCoupon() int {
	return dbcfg.RmbCouponRate
}

func GetPriceMenu() *[]PriceItem {
	rt := new([]PriceItem)
	*rt = append(*rt, PriceItem{
		PriceID: 1,
		Fee:     20000,
		Name:    "点券",
		Amount:  200,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 2,
		Fee:     10000,
		Name:    "点券",
		Amount:  100,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 3,
		Fee:     5000,
		Name:    "点券",
		Amount:  50,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 4,
		Fee:     2000,
		Name:    "点券",
		Amount:  20,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 5,
		Fee:     1000,
		Name:    "点券",
		Amount:  10,
	})
	*rt = append(*rt, PriceItem{
		PriceID: 6,
		Fee:     500,
		Name:    "点券",
		Amount:  5,
	})

	return rt
}
