package values

// RestartConfig 服务器重启配置
type RestartConfig struct {
	RestartTime int64
	Restartting bool
}

// WhiteListConfig 白名单配置
type WhiteListConfig struct {
	WhiteSwitch bool  `bson:"whiteswitch"`
	WhiteList   []int `bson:"whitelist"`
}

// 默认配置
var (
	DefaultRestartConfig   = RestartConfig{}
	DefaultWhiteListConfig = WhiteListConfig{}
)
