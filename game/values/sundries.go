package values

const (
	PropTypeCoupon     = 1
	PropTypeAward      = 2
	PropTypeCouponFrag = 3
	PropTypeRedScore   = 4
)

var PropTypes = []int{PropTypeCoupon, PropTypeAward, PropTypeCouponFrag, PropTypeRedScore}

type PropBaseConfig struct {
	ID       int    `bson:"_id"` //唯一标识
	PropType int    //道具类型, 1是点券，2是奖金，3点券碎片
	PropID   int    //道具id
	Name     string //名称
	ImgUrl   string //图片url
	Operator string //操作人

	CreatedAt int //创建时间戳
	UpdatedAt int //更新时间戳
	DeletedAt int //删除时间戳，0表示没有删除
}

var AwardWordToPropType = map[string]int {
	Money: PropTypeAward,
	RedScore: PropTypeRedScore,
	Coupon: PropTypeCoupon,
	Fragment: PropTypeCouponFrag,
}