package values

type HorseRaceLampControl struct {
	ID           int    `bson:"_id"` //唯一标识
	Name         string //通告名称
	Level        int    //等级排序，1：A，2：B，3：C，4：D
	ExpiredAt    int    //过期时间戳
	TakeEffectAt int    //发布时间戳
	Duration     int    //间隔时长，单位s
	LinkMatchID  string //关联赛事id
	Content      string //内容
	Operator     string //操作人
	Status       int    //0表示发布，1表示暂停，2表示过期

	CreatedAt int //创建时间戳
	UpdatedAt int //更新时间戳，0表示未更新，对应着操作时间
	DeletedAt int //删除时间戳，0表示未删除
}
