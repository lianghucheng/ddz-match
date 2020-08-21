package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/szxby/tools/log"
)

type Config struct {
	CfgLeafSvr       CfgLeafSvr
	CfgTimeout       CfgTimeout
	CfgDDZ           CfgDDZ
	CfgDailySign     []CfgDailySign
	CfgRedis         CfgRedis
	CfgJuHeSms       CfgJuHeSms
	CfgFirstRecharge []CfgFirstRecharge
	CfgNotice        []CfgNotice
	CfgHall          CfgHall
}
type CfgLeafSvr struct {
	LogLevel       string
	LogPath        string
	WSAddr         string
	CertFile       string
	KeyFile        string
	TCPAddr        string
	MaxConnNum     int
	DBUrl          string
	DBMaxConnNum   int
	ConsolePort    int
	ProfilePath    string
	HTTPAddr       string
	DBName         string
	Model          bool //false :表示测试环境  true:表示正式环境
	FirstRecharge  int64
	AgentServer    string // 推广服务器
	ActivityServer string // 活动服务器
}
type CfgDDZ struct {
	DefaultAndroidDownloadUrl string
	DefaultIOSDownloadUrl     string
	DefaultSougouDownloadUrl  string
	Gamename                  string
	AndroidVersion            int
	IOSVersion                int
	SougouVersion             int
	AndroidGuestLogin         bool
	IOSGuestLogin             bool
	SougouGuestLogin          bool
	Notice                    string
	Radio                     string
	WeChatNumber              string
	EnterAddress              bool
	CardCodeDesc              string
}

type CfgTimeout struct {
	ConnectTimeout         int
	HeartTimeout           int
	LandlordBid            int
	LandlordSystemHost     int
	LandlordDouble         int
	LandlordDiscard        int
	LandlordDiscardNothing int
	LandloadMatchPrepare   int
	LandlordEndPrepare     int
	LandlordNextStart      int
}
type CfgDailySign struct {
	Chips int64
}

const (
	TicketCounting = 1
)

type CfgFirstRecharge struct {
	Type int
	Num  int
}

type CfgNotice struct {
	Title   string
	Content string
}
type CfgRedis struct {
	Address  string //数据库连接地址。
	Password string //数据库连接地址。
	Db       int    //db序号
}

type CfgJuHeSms struct {
	AppKey           string
	FindTemplate     string
	RegisterTemplate string
}

type CfgHall struct {
	SignIcon          bool     //签到标签是否显示
	NewWelfareIcon    bool     //新人福利标签是否显示
	FirstRechargeIcon bool     //首充标签是否显示
	ShareIcon         bool     //分享推广标签是否显示
	UserMailLimit     int      //玩家邮件列表限制数量
	MailDefaultExpire int      //默认过期时间多少天
	RankingTitle      []string //排行榜标题序列
	RankTypeJoinNum   string   //参赛
	RankTypeWinNum    string   //胜局
	RankTypeFailNum   string   //衰神
	RankTypeAward     string   //奖励
	WithDrawMin       float64  //最低提奖
}

var ServerConfig Config

func init() {
	ReadConfigure()
}
func ReadConfigure() {
	_, err := toml.DecodeFile("conf/ddz-server.toml", &ServerConfig)
	if err != nil {
		log.Error("读取server.toml失败,error:%v", err)
	}
	log.Release("*****************:%v", ServerConfig.CfgTimeout)

}

func GetCfgLeafSrv() *CfgLeafSvr {
	return &ServerConfig.CfgLeafSvr
}
func GetCfgTimeout() *CfgTimeout {
	return &ServerConfig.CfgTimeout
}
func GetCfgDDZ() *CfgDDZ {
	return &ServerConfig.CfgDDZ
}

func GetCfgDailySign() []CfgDailySign {
	return ServerConfig.CfgDailySign
}

func GetCfgRedis() *CfgRedis {
	return &ServerConfig.CfgRedis
}

func GetCfgJuHeSms() *CfgJuHeSms {
	return &ServerConfig.CfgJuHeSms
}

func GetCfgFirstRechage() []CfgFirstRecharge {
	return ServerConfig.CfgFirstRecharge
}

func GetCfgNotice() []CfgNotice {
	return ServerConfig.CfgNotice
}

func GetCfgHall() *CfgHall {
	return &ServerConfig.CfgHall
}
