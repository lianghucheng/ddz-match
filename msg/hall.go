package msg

func init() {
	Processor.Register(&S2C_Knapsack{})
	Processor.Register(&C2S_UseProp{})
	Processor.Register(&S2C_UseProp{})
}

type KnapsackProp struct {
	PropID    int    //道具id
	Name      string //名称
	Num       int    //数量
	IsUse     bool   //是否可使用
	Expiredat int64  //过期时间，-1表示永久
	Desc      string //描述
	Createdat int64  //创建时间
}

type S2C_Knapsack struct {
	Props *[]KnapsackProp //道具数据列表
}

type C2S_UseProp struct {
	PropID int //道具id
	Amount int //要领取的结果数量，比如40个碎片换取2个点券，此时传输2
}

type S2C_UseProp struct {
	Error  int    //错误码
	ErrMsg string //错误信息
}
