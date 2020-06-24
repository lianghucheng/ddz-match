package edy_api

var ErrMsg = make(map[string]string)

func init() {
	ErrMsg["000001"] = "此玩家没有绑定身份证号"
	ErrMsg[`000002`]=`重复的对局结果上报失败`
	ErrMsg[`000003`]=`绑定身份信息时出错,已经绑定身份证号码`
	ErrMsg[`000004`]=`绑定身份信息时出错,身份证号码已被其他用户绑定`
	ErrMsg[`000005`]=`没有绑定银行卡`
	ErrMsg[`000006`]=`钱包余额不足`
	ErrMsg[`000007`]=`身份证号不匹配`
	ErrMsg[`000008`]=`排名必须大于等于1`
	ErrMsg[`000009`]=`钱包不存在`
	ErrMsg[`000010`]=`未达到报名要求`
	ErrMsg[`000011`]=`超出取现次数限制`
	ErrMsg[`000012`]=`绑定银行卡信息时出错,已经绑定银行卡`
	ErrMsg[`000013`]=`绑定银行卡信息时出错,银行卡已被其他用户绑定`
	ErrMsg[`000014`]=`提现失败`
	ErrMsg[`000015`]=`身份证格式错误`
	ErrMsg[`000016`]=`银行编号与银行名称不匹配`
	ErrMsg[`000017`]=`此赛事不存在`
	ErrMsg[`000018`]=`未绑定奖金`
	ErrMsg[`000019`]=`此赛事已结束`
	ErrMsg[`000020`]=`此赛事未审核通过`
	ErrMsg[`000021`]=`绑定身份信息时出错,真实姓名和其它平台注册时的姓名不一致`
	ErrMsg[`000022`]=`重复上报或晋级人员在海选赛排名中不存在`
	ErrMsg[`000023`]=`该海选赛未正常结束`
	ErrMsg[`000024`]=`该海选赛与晋级赛事未绑定`
	ErrMsg[`000025`]=`该赛事不是海选赛`
	ErrMsg[`000026`]=`本场比赛已达到报名人数上限`
	ErrMsg[`000093`]=`接口已关闭`
	ErrMsg[`000094`]=`重复上报`
	ErrMsg[`000095`]=`厂商ID错误`
	ErrMsg[`000096`]=`时间戳错误`
	ErrMsg[`000097`]=`签名错误`
	ErrMsg[`000098`]=`参数不合法`
	ErrMsg[`000099`]=`参数不能为空`
	ErrMsg[`000100`]=`内部错误`
}
