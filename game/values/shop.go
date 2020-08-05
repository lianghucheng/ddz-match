package values

const (
	TakenTypeRMB = 1
)

const (
	PropTypeCoupon = 1
)

var PropTypeStr = map[int]string{
	PropTypeCoupon:"点券",
}

type Goods struct {
	ID int `bson:"_id"`
	GoodsTypeID int //商品类型唯一标识
	TakenType int//花费类型。1是RMB
	Price int//花费数量（价格，百分制）
	PropType int//道具类型。1是点券
	GetAmount int//获得数量
	GiftAmount int//赠送数量
	Expire int//过期时间，单位秒，-1为永久
	ImgUrl string//商品图标
	Order int//次序
	UpdatedAt int//更新时间戳
	CreatedAt int//创建时间戳
	DeletedAt int//删除时间戳
}

type GoodsType struct {
	ID int `bson:"_id"`//唯一标识
	MerchantID int//商户唯一标识
	TypeName string//商品名称
	ImgUrl string//商品图标
	Order int//次序
	UpdatedAt int//更新时间戳
	CreatedAt int//创建时间戳
	DeletedAt int//删除时间戳
}

type ShopMerchant struct {
	ID int `bson:"_id"`
	MerchantType int//商户类型。1是体总
	MerchantNo string//商户编号
	PayMin int//支付最低值，百分制
	PayMax int//支付最高值，百分制
	PublicKey string//公钥
	PrivateKey string//私钥
	Order int//次序
	UpPayBranchs []int//上架支付类型
	DownPayBranchs []int//下架支付类型
	UpDownStatus int//上下架状态。0是下架，1是上架
	UpdatedAt int//更新时间戳
	CreatedAt int
	DeletedAt int
}

type PayAccount struct {
	ID int `bson:"_id"`
	MerchantID int//商户唯一标识
	PayBranch int//支付渠道标识
	Order int //次序
	Account string //账户
	UpdatedAt int//更新时间戳
	CreatedAt int//更新时间戳
	DeletedAt int//删除时间戳
}