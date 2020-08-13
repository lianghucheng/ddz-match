package db

import (
	"ddz/config"
	"ddz/game/values"
	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

func save(coll string, data interface{}, selector bson.M) {
	se := BackstageDB.Ref()
	defer BackstageDB.UnRef(se)
	dbName := config.GetCfgDB().BackstageDBName
	_, err := se.DB(dbName).C(coll).Upsert(selector, data)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

const (
	readOne = 1
	readAll = 2
)

func read(coll string, data interface{}, query bson.M, model int) {
	query["deletedat"] = -1
	se := BackstageDB.Ref()
	defer BackstageDB.UnRef(se)
	dbName := config.GetCfgDB().BackstageDBName
	var err error
	if model == readOne {
		err = se.DB(dbName).C(coll).Find(query).One(data)
	} else if model == readAll {
		err = se.DB(dbName).C(coll).Find(query).Sort("order").All(data)
	}
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func GetInterQuery(cond interface{}) bson.M {
	if cond == nil {
		return nil
	}
	bson_arr := []bson.M{}
	cond_map, ok := cond.(map[string]interface{})
	if !ok {
		return nil
	}
	if len(cond_map) == 0 {
		return nil
	}
	for k, v := range cond_map {
		bson_arr = append(bson_arr, bson.M{k: v})
	}
	return bson.M{"$or": bson_arr}
}

func ReadGoodses(query bson.M) *[]values.Goods {
	data := new([]values.Goods)
	read("shopgoods", data, query, readAll)
	return data
}

func ReadGoodsById(id int) *values.Goods {
	data := new(values.Goods)
	read("shopgoods", data, bson.M{"_id": id}, readOne)
	return data
}

func ReadGoodsTypes(merID int) *[]values.GoodsType {
	datas := new([]values.GoodsType)
	read("shopgoodstype", datas, bson.M{"merchantid": merID}, readAll)
	return datas
}

func ReadGoodsTypeFirst() *values.GoodsType {
	data := new(values.GoodsType)
	se := BackstageDB.Ref()
	defer BackstageDB.UnRef(se)
	dbName := config.GetCfgDB().BackstageDBName
	var err error
	err = se.DB(dbName).C("shopgoodstype").Find(bson.M{"deletedat": -1}).Sort("order").One(data)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return data
}

func ReadShopMerchant() *values.ShopMerchant {
	data := new(values.ShopMerchant)
	read("shopmerchant", data, bson.M{"updownstatus": 1}, readOne)
	log.Debug("商家数据： %v", *data)
	return data
}

func ReadPayAccounts(merID, paybranch int) *[]values.PayAccount {
	datas := new([]values.PayAccount)
	read("shoppayaccount", datas, bson.M{"merchantid": merID, "paybranch": paybranch}, readAll)
	log.Debug("支付账号数据： %v", *datas)
	return datas
}
