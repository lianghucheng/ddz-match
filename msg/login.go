package msg

import "ddz/conf"

func init() {
	Processor.Register(&S2C_FirstRechage{})
	Processor.Register(&S2C_Notice{})
	Processor.Register(&S2C_UpdateUserCoupon{})
	Processor.Register(&S2C_UpdateUserAfterTaxAward{})
	Processor.Register(&C2S_GetCoupon{})
	Processor.Register(&S2C_GetCoupon{})
	Processor.Register(&C2S_SetNickName{})
	Processor.Register(&S2C_UpdateNickName{})
}

type C2S_TokenLogin struct {
	Token string
}

type C2S_AccountLogin struct {
	Account string //手机号
	Code    string //验证码
}

// Close
const (
	S2C_Close_LoginRepeated   = 1  // 您的账号在其他设备上线，非本人操作请注意修改密码
	S2C_Close_InnerError      = 2  // 登录出错，请重新登录
	S2C_Close_TokenInvalid    = 3  // 登录状态失效，请重新登录
	S2C_Close_UsernameInvalid = 5  // 登录出错，用户名无效
	S2C_Close_SystemOff       = 6  // 系统升级维护中，请稍后重试
	S2C_Close_RoleBlack       = 7  // 账号已冻结，请联系客服微信 S2C_Close.WeChatNumber
	S2C_Close_IPChanged       = 8  // 登录IP发生变化，非本人操作请注意修改密码
	S2C_Close_Code_Valid      = 9  // 验证码错误
	S2C_Close_Code_Error      = 10 // 验证码过期了
)

type S2C_Close struct {
	Error        int
	WeChatNumber string
}

type S2C_Login struct {
	AccountID         int
	Nickname          string
	Headimgurl        string
	Sex               int // 1 男、2 女
	Role              int // 1 玩家、2 代理、3 管理员、4 超管
	Token             string
	AnotherLogin      bool    // 其他设备登录
	FirstLogin        bool    // 首次登录
	AfterTaxAward     float64 // 税后奖金
	Coupon            int     // 点劵数量
	SignIcon          bool    //签到标签是否显示
	NewWelfareIcon    bool    //新人福利标签是否显示
	FirstRechargeIcon bool    //首充标签是否显示
	ShareIcon         bool    //分享推广标签是否显示
	Customer          string  //客服
	RealName          string
	PhoneNum          string
	BankName       string //银行名称
	BankCardNoTail string //银行卡号后四位
}

type S2C_FirstRechage struct {
	Gifts []conf.CfgFirstRecharge
	Money int64
}

type S2C_Notice struct {
	Notices []conf.CfgNotice
}

type S2C_UpdateUserCoupon struct {
	Coupon int64
}

type S2C_UpdateUserAfterTaxAward struct {
	AfterTaxAward float64 // 税后奖金
}

type C2S_GetCoupon struct {
	Count int64 //购买点券的数量
}

const (
	S2C_GetCouponSuccess = 0
	S2C_GetCouponFailed  = 1
)

type S2C_GetCoupon struct {
	Error int
}

type C2S_SetNickName struct {
	NickName string //字符长度3-18
}

const (
	S2C_SetNickName_Length = 1 //长度不合法
	S2C_SetNickName_More   = 2 //超过修改次数
)

type S2C_UpdateNickName struct {
	Error    int    // 0:表示成功
	NickName string //
}
