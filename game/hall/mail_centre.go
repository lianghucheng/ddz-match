package hall

import (
	"ddz/conf"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/msg"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/szxby/tools/log"
	"gopkg.in/mgo.v2/bson"
)

const (
	MailTypeText  = 1
	MailTypeAward = 2
	MailTypeMix   = 3
)

type MailBox struct {
	ID          int64   `bson:"_id"`          //唯一id
	TargetID    int64   `json:"target_id"`    //目标用户， -1表示所有用户
	MailType    int     `json:"mail_type"`    //邮箱邮件类型
	CreatedAt   int64   `json:"created_at"`   //收件时间
	Title       string  `json:"title"`        //主题
	Content     string  `json:"content"`      //内容
	Annexes     []Annex `json:"annexes"`      //附件
	Status      int64   `json:"status"`       //邮箱邮件状态
	ExpireValue int64   `json:"expire_value"` //有效时长
}

type Annex struct {
	Type int    `json:"type"`
	Num  int    `json:"num"`
	Desc string `json:"desc"`
}

const (
	NotReadUserMail = 0
	ReadUserMail    = 1
	TakenUserMail   = 2
	DeleteUserMail  = 3
)

type UserMail struct {
	ID          int64 `bson:"_id"` //唯一主键
	UserID      int64
	CreatedAt   int64   //收件时间
	Title       string  //主题
	Content     string  //内容
	Annexes     []Annex //附件
	Status      int64   //用户邮件状态
	ExpireValue int64   //有效时长
	ExpiredAt   int64   //有效期
}

func ReadMail(mid int64) {
	mail := readUserMailByID(mid)
	mail.Status = ReadUserMail
	mail.ExpiredAt = time.Now().Unix() + mail.ExpireValue*86400
	mail.save()
}

func readUserMailByID(id int64) *UserMail {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)

	rt := new(UserMail)
	err := se.DB(db.DB).C("usermail").Find(bson.M{"_id": id}).One(rt)
	if err != nil {
		log.Error(err.Error())
	}
	return rt
}

func (ctx *UserMail) save() {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)

	if _, err := se.DB(db.DB).C("usermail").Upsert(bson.M{"_id": ctx.ID}, ctx); err != nil {
		log.Error(err.Error())
	}
}

//玩家登录调用一次，系统发邮件时调用一次
func SendMail(user *player.User) {
	mails := readMailBox(user.BaseData.UserData.UserID, user.GetUserData().LastTakenMail)
	if len(*mails) > 0 {
		user.GetUserData().LastTakenMail = (*mails)[len(*mails)-1].ID
	}

	game.GetSkeleton().Go(func() {
		pullMailBox(user, mails)
	}, func() {
		usermails := readUserMail(user.UID())
		user.WriteMsg(&msg.S2C_SendMail{
			Datas: *transferMsgUserMail(usermails),
		})
	})
}

func readMailBox(userid int, lastId int64) *[]MailBox {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)

	rt := new([]MailBox)

	err := se.DB(db.DB).C("mailbox").Find(bson.M{"_id": bson.M{"$gt": lastId}, "$or": []bson.M{{"targetid": userid}, {"targetid": -1}}}).All(rt)
	if err != nil {
		log.Error(err.Error())
	}
	return rt
}

func (ctx *MailBox) save() {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	_, err := se.DB(db.DB).C("mailbox").Upsert(bson.M{"_id": ctx.ID}, ctx)
	if err != nil {
		log.Error(err.Error())
	}
}

func pullMailBox(user *player.User, mails *[]MailBox) {
	for _, v := range *mails {
		userMail := new(UserMail)
		id, err := db.MongoDBNextSeq("usermail")
		if err != nil {
			log.Error(err.Error())
		}
		userMail.ID = int64(id)
		userMail.UserID = int64(user.GetUserData().UserID)
		userMail.CreatedAt = v.CreatedAt
		userMail.Title = v.Title
		userMail.Content = v.Content
		userMail.Annexes = v.Annexes
		userMail.ExpireValue = v.ExpireValue
		userMail.save()
	}
}

func GamePushMail(userid int, title, content string) {
	mailBox := new(MailBox)
	mailBox.TargetID = int64(userid)
	mailBox.ExpireValue = int64(conf.GetCfgHall().MailDefaultExpire)
	mailBox.MailType = MailTypeText
	mailBox.Title = title
	mailBox.Content = content

	mailBox.pushMailBox()
}

func MatchEndPushMail(userid int, matchName string, order int, award string) {
	mailBox := new(MailBox)
	mailBox.TargetID = int64(userid)
	mailBox.ExpireValue = 30
	mailBox.MailType = MailTypeText
	mailBox.Title = "比赛通知"
	mailBox.Content = fmt.Sprintf("恭喜您在【%v】比赛中获得第【%v】名。", matchName, order)
	if len(award) > 0 {
		mailBox.Content += fmt.Sprintf("【%v奖励】已经发放至您的钱包或背包。请核对您的奖励信息。", award)
	}

	mailBox.pushMailBox()
}

func MatchInterruptPushMail(userid int, matchName string, coupon int64) {
	mailBox := new(MailBox)
	mailBox.TargetID = int64(userid)
	mailBox.ExpireValue = int64(conf.GetCfgHall().MailDefaultExpire)
	mailBox.MailType = MailTypeText
	mailBox.Title = "退赛通知"
	mailBox.Content = fmt.Sprintf("很抱歉，您报名的【%v】赛事因为人数不足而取消，"+
		"您的报名费【%v点券】已经返回至您的账户，请注意查收！感谢您的支持，祝您下次获得好成绩！",
		matchName, coupon)

	mailBox.pushMailBox()
}

//todo：卡住原因，产品还没出发邮件的时机规则
func (ctx *MailBox) pushMailBox() {
	id, err := db.MongoDBNextSeq("mailbox")
	if err != nil {
		log.Error(err.Error())
	}
	ctx.ID = int64(id)
	ctx.CreatedAt = time.Now().Unix()
	ctx.ExpireValue = 30
	game.GetSkeleton().Go(func() {
		ctx.save()
	}, func() {
		game.GetSkeleton().ChanRPCServer.Go("SendMail", &msg.RPC_SendMail{ID: int(ctx.TargetID)})
	})
}

func readUserMail(uid int) *[]UserMail {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	limit := conf.GetCfgHall().UserMailLimit
	rt := new([]UserMail)

	notRead := new([]UserMail)
	err := se.DB(db.DB).C("usermail").
		Find(bson.M{"status": NotReadUserMail, "userid": uid}).
		Sort("-createdat").Limit(limit).All(notRead)
	if err != nil {
		log.Error(err.Error())
	}
	*rt = append(*rt, *notRead...)
	if len(*notRead) < conf.GetCfgHall().UserMailLimit {
		limit -= len(*notRead)
		readed := new([]UserMail)
		err = se.DB(db.DB).C("usermail").
			Find(bson.M{"status": ReadUserMail, "expiredat": bson.M{"$gt": time.Now().Unix()}, "userid": uid}).
			Sort("-createdat").Limit(limit).All(readed)
		if err != nil {
			log.Error(err.Error())
		}
		*rt = append(*rt, *readed...)
	}

	return rt
}

func transferMsgUserMail(usermails *[]UserMail) *[]msg.UserMail {
	rt := new([]msg.UserMail)
	for _, v := range *usermails {
		temp := new(msg.UserMail)
		temp.Annexes = *transferMsgAnnexes(v.Annexes)
		temp.Content = v.Content
		temp.Title = v.Title
		temp.CreatedAt = v.CreatedAt
		temp.ID = v.ID
		temp.Status = v.Status

		*rt = append(*rt, *temp)
	}

	return rt
}

func transferMsgAnnexes(annex []Annex) *[]msg.Annex {
	rt := new([]msg.Annex)
	for _, v := range annex {
		temp := new(msg.Annex)
		temp.Desc = v.Desc
		temp.Num = v.Num
		temp.Type = v.Type
		*rt = append(*rt, *temp)
	}
	return rt
}

func DeleteMail(mid int64) {
	mail := readUserMailByID(mid)
	mail.Status = DeleteUserMail
	mail.save()
}

type MailBoxParam struct {
	TargetID    int64   `json:"target_id"`    //目标用户， -1表示所有用户
	MailType    int     `json:"mail_type"`    //邮箱邮件类型
	Title       string  `json:"title"`        //主题
	Content     string  `json:"content"`      //内容
	Annexes     []Annex `json:"annexes"`      //附件
	ExpireValue int64   `json:"expire_value"` //有效时长
}

func HandlePushMail(w http.ResponseWriter, r *http.Request) {
	data := r.FormValue("data")
	fmt.Println(data)
	param := new(MailBoxParam)
	if err := json.Unmarshal([]byte(data), param); err != nil {
		log.Error(err.Error())
		return
	}

	param.pushMailBox()
}

func (ctx *MailBoxParam) pushMailBox() {
	mailBox := new(MailBox)
	mailBox.TargetID = ctx.TargetID
	mailBox.ExpireValue = ctx.ExpireValue
	mailBox.MailType = ctx.MailType
	mailBox.Title = ctx.Title
	mailBox.Content = ctx.Content
	mailBox.Annexes = ctx.Annexes

	mailBox.pushMailBox()
}

func TakenMailAnnex(mid int64) {
	mail := readUserMailByID(mid)
	mail.Status = TakenUserMail
	mail.save()
}
