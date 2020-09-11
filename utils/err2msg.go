package utils

const (
	Success               = 1000
	Fail                  = 1001
	FormatFail            = 1002
	TaxFeeLack            = 1003
	UserNotExist          = 1004
	MatchNotExist         = 10005
	MatchRobotConfExist   = 10006
	RobotNotBan           = 10007
	MongoDBCreFail        = 10008
	ModelTransferFail     = 10009
	PropBaseConfCacheFail = 10010
	MongoReadFail         = 10011
	PropIDNotExist        = 10012
	MailcontrolFail       = 10013
	PayLimitRangeError    = 10014
	SendAllMailFail       = 10015
)

var ErrMsg = map[int]string{
	Success:               "成功",
	Fail:                  "失败",
	FormatFail:            "格式错误",
	TaxFeeLack:            "所剩税后奖金不足",
	UserNotExist:          "该用户不存在",
	MatchNotExist:         "该赛事不存在",
	MatchRobotConfExist:   "该赛事机器人配置已存在",
	RobotNotBan:           "该赛事机器人没有金禁用",
	MongoDBCreFail:        "MongoDB自增错误",
	ModelTransferFail:     "模型格式转换失败",
	PropBaseConfCacheFail: "道具配置失败，请重试",
	MongoReadFail:         "数据库读取数据失败",
	PropIDNotExist:        "道具id不存在",
	MailcontrolFail:       "操作失败",
	PayLimitRangeError:    "支付限额范围非法",
	SendAllMailFail:       "一键发送邮件失败",
}
