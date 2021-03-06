package msg

import "ddz/conf"

func init() {
	Processor.Register(&S2C_FirstRechage{})
	Processor.Register(&S2C_UpdateUserCoupon{})
	Processor.Register(&S2C_UpdateUserAfterTaxAward{})
	Processor.Register(&C2S_GetCoupon{})
	Processor.Register(&S2C_GetCoupon{})
	Processor.Register(&C2S_SetNickName{})
	Processor.Register(&S2C_UpdateNickName{})
	Processor.Register(&C2S_UsrnPwdLogin{})
	Processor.Register(&C2S_Register{})
	Processor.Register(&S2C_Register{})
	Processor.Register(&C2S_FindPassword{})
	Processor.Register(&S2C_FindPassword{})
	Processor.Register(&C2S_ChangePassword{})
	Processor.Register(&S2C_ChangePassword{})
	Processor.Register(&C2S_TakenFirstCoupon{})
	Processor.Register(&S2C_TakenFirstCoupon{})
}

type C2S_TokenLogin struct {
	Token string
}

type C2S_AccountLogin struct {
	Account  string //手机号
	Code     string //验证码
	Password string //密码
}

type C2S_UsrnPwdLogin struct {
	Username string
	Password string
}

type C2S_Register struct {
	Account   string //手机号
	Code      string //验证码
	Password  string //密码
	ShareCode string // 邀请码
}

const (
	ErrRegisterSuccess = 0
)

type S2C_Register struct {
	Error  int
	ErrMsg string
}

type C2S_FindPassword struct {
	Account  string //手机号
	Code     string //验证码
	Password string //密码
}

const (
	ErrFindPasswordSuccess = 0
)

type S2C_FindPassword struct {
	Error  int
	ErrMsg string
}

type C2S_ChangePassword struct {
	OldPassword string
	NewPassword string
}

const (
	ErrChangePasswordSuccess = 0
	ErrChangePasswordFail    = 1
	ErrChangePasswordOldNo   = 2
)

type S2C_ChangePassword struct {
	Error int
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
	S2C_Close_Pwd_Error       = 11 // 密码错误
	S2C_Close_Usrn_Nil        = 12 // 用户名不存在
	S2C_Close_Usrn_Exist      = 13 // 用户名不存在
	S2C_Close_Pass_Length     = 14 // 密码长度8到15位
	S2C_Close_ServerRestart   = 15 // 服务器停服更新
)

type S2C_Close struct {
	Error        int
	WeChatNumber string
	Info         interface{}
}

type Customer struct {
	WeChat   string //微信
	Email    string //邮箱
	PhoneNum string //电话号码
	QQ       string
	QQGroup  string
}

type S2C_Login struct {
	AccountID         int
	Nickname          string
	Headimgurl        string
	Sex               int // 1 男、2 女
	Role              int // 1 玩家、2 代理、3 管理员、4 超管
	Token             string
	AnotherLogin      bool     // 其他设备登录
	FirstLogin        bool     // 首次登录
	AfterTaxAward     float64  // 税后奖金
	Coupon            int      // 点劵数量
	SignIcon          bool     //签到标签是否显示
	NewWelfareIcon    bool     //新人福利标签是否显示
	FirstRechargeIcon bool     //首充标签是否显示
	ShareIcon         bool     //分享推广标签是否显示
	Customer          Customer //客服
	RealName          string
	PhoneNum          string
	BankName          string //银行名称
	BankCardNoTail    string //银行卡号后四位
	SetNickName       bool
	IsNewUser         bool
}

type S2C_FirstRechage struct {
	Gifts []conf.CfgFirstRecharge
	Money int64
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
	ErrMsg   string
}

type C2S_TakenFirstCoupon struct {
}

const (
	ErrS2CTakenFirstCouponSuccess = 0
	ErrS2CTakenFirstCouponFail    = 1
)

type S2C_TakenFirstCoupon struct {
	Error int
}
