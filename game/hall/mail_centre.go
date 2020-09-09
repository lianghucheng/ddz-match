package hall

import (
	"ddz/conf"
	"ddz/config"
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/game/values"
	"ddz/msg"
	"errors"
	"fmt"
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
	ID              int64   `bson:"_id"`          //唯一id
	TargetID        int64   `json:"target_id"`    //目标用户， -1表示所有用户
	MailType        int     `json:"mail_type"`    //邮箱邮件类型
	CreatedAt       int64   `json:"created_at"`   //收件时间
	Title           string  `json:"title"`        //主题
	Content         string  `json:"content"`      //内容
	Annexes         []Annex `json:"annexes"`      //附件
	Status          int64   `json:"status"`       //邮箱邮件状态
	ExpireValue     float64 `json:"expire_value"` //有效时长
	MailServiceType int     //邮件服务类型
}

const (
	AnnexTypeCoupon     = 1
	AnnexTypeCouponFrag = 2
)

type Annex struct {
	PropType int     `json:"prop_type"`
	Num      float64 `json:"num"`
	Desc     string  `json:"desc"`
}

const (
	NotReadUserMail = 0
	ReadUserMail    = 1
	TakenUserMail   = 2
	DeleteUserMail  = 3
)

const (
	MailServiceTypeOfficial = 0
	MailServiceTypeMatch    = 1
	MailServiceTypeActivity = 2
)

type UserMail struct {
	ID              int64 `bson:"_id"` //唯一主键
	UserID          int64
	CreatedAt       int64   //收件时间
	Title           string  //主题
	Content         string  //内容
	Annexes         []Annex //附件
	Status          int64   //用户邮件状态
	ExpireValue     float64 //有效时长
	ExpiredAt       int64   //有效期
	MailServiceType int     //邮件服务类型
	MailType        int
}

func ReadMail(mid int64) {
	mail := readUserMailByID(mid)
	if mail.Status != NotReadUserMail {
		return
	}
	mail.Status = ReadUserMail
	mail.ExpiredAt = time.Now().Unix() + int64(mail.ExpireValue*86400)
	log.Debug("邮件过期时间：%v", mail.ExpiredAt)
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
	mails := readMailBox(user, user.GetUserData().LastTakenMail)
	if len(*mails) > 0 {
		user.GetUserData().LastTakenMail = (*mails)[len(*mails)-1].ID
	}

	game.GetSkeleton().Go(func() {
		pullMailBox(user, mails)
	}, func() {
		officialUsermails := readUserMail(user.UID(), MailServiceTypeOfficial)
		matchUsermails := readUserMail(user.UID(), MailServiceTypeMatch)
		activityUsermails := readUserMail(user.UID(), MailServiceTypeActivity)
		user.WriteMsg(&msg.S2C_SendMail{
			Datas:    *transferMsgUserMail(officialUsermails),
			Match:    *transferMsgUserMail(matchUsermails),
			Activity: *transferMsgUserMail(activityUsermails),
		})
	})
}

func readMailBox(user *player.User, lastId int64) *[]MailBox {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)

	rt := new([]MailBox)

	err := se.DB(db.DB).C("mailbox").Find(bson.M{"_id": bson.M{"$gt": lastId}, "$or": []bson.M{{"targetid": user.UID()}, {"targetid": -1}}, "createdat": bson.M{"$gt": user.BaseData.UserData.CreatedAt}}).All(rt)
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
		userMail.MailServiceType = v.MailServiceType
		userMail.MailType = v.MailType
		userMail.save()
	}
}

func GamePushMail(userid int, title, content string) {
	mailBox := new(MailBox)
	mailBox.TargetID = int64(userid)
	mailBox.ExpireValue = float64(conf.GetCfgHall().MailDefaultExpire)
	mailBox.MailType = MailTypeText
	mailBox.MailServiceType = MailServiceTypeMatch
	mailBox.Title = title
	mailBox.Content = content

	mailBox.pushMailBox()
}

func MatchEndPushMail(userid int, matchName string, order int, award string) {
	mailBox := new(MailBox)
	mailBox.TargetID = int64(userid)
	mailBox.ExpireValue = 30
	mailBox.MailType = MailTypeText
	mailBox.MailServiceType = MailServiceTypeMatch
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
	mailBox.ExpireValue = float64(conf.GetCfgHall().MailDefaultExpire)
	mailBox.MailType = MailTypeText
	mailBox.MailServiceType = MailServiceTypeMatch
	mailBox.Title = "退赛通知"
	mailBox.Content = fmt.Sprintf("很抱歉，您报名的【%v】赛事因为人数不足而取消，"+
		"您的报名费【%v点券】已经返回至您的账户，请注意查收！感谢您的支持，祝您下次获得好成绩！",
		matchName, coupon)

	mailBox.pushMailBox()
}

func PrizePresentationPushMail(userid int, bankName string, fee float64) {
	mailBox := new(MailBox)
	mailBox.TargetID = int64(userid)
	mailBox.ExpireValue = float64(conf.GetCfgHall().MailDefaultExpire)
	mailBox.MailType = MailTypeText
	mailBox.MailServiceType = MailServiceTypeOfficial
	mailBox.Title = "提奖通知"
	mailBox.Content = fmt.Sprintf("您的提奖金额【%v】已下发到您的【%v】，如有问题请联系客服微信：wkxjingjipingtai", fee, bankName)

	mailBox.pushMailBox()
}

func RefundPushMail(userid int, fee float64) {
	mailBox := new(MailBox)
	mailBox.TargetID = int64(userid)
	mailBox.ExpireValue = float64(conf.GetCfgHall().MailDefaultExpire)
	mailBox.MailType = MailTypeText
	mailBox.MailServiceType = MailServiceTypeOfficial
	mailBox.Title = "退款通知"
	mailBox.Content = fmt.Sprintf("您的提奖金额【%v】已被官方退回，请联系客服进行处理；客服QQ：wkxjingjipingtai", fee)

	mailBox.pushMailBox()
}

// ShareAwardPushMail 分享有奖邮件
func ShareAwardPushMail(accountID int, itemID int, itemNum int) error {
	mailBox := new(MailBox)
	user := player.ReadUserDataByAid(accountID)
	if user.AccountID <= 0 {
		return errors.New("未知用户")
	}
	mailBox.TargetID = int64(user.AccountID)
	mailBox.ExpireValue = float64(conf.GetCfgHall().MailDefaultExpire)
	mailBox.MailType = MailTypeMix

	itemType := values.PropID2Type[itemID]
	mailBox.Annexes = []Annex{
		Annex{
			PropType: itemType,
			Num:      float64(itemNum),
		},
	}

	mailBox.MailServiceType = MailServiceTypeOfficial
	mailBox.Title = "分享奖励"
	mailBox.Content = fmt.Sprintf("恭喜成功分享一名好友,奖励%v点券。", itemNum)

	mailBox.pushMailBox()
	return nil
}

func (ctx *MailBox) pushMailBox() {
	id, err := db.MongoDBNextSeq("mailbox")
	if err != nil {
		log.Error(err.Error())
	}
	ctx.ID = int64(id)
	ctx.CreatedAt = time.Now().Unix()
	game.GetSkeleton().Go(func() {
		ctx.save()
	}, func() {
		game.GetSkeleton().ChanRPCServer.Go("SendMail", &msg.RPC_SendMail{ID: int(ctx.TargetID)})
	})
}

func readUserMail(uid, mailServiceType int) *[]UserMail {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	limit := conf.GetCfgHall().UserMailLimit
	rt := new([]UserMail)

	notRead := new([]UserMail)
	err := se.DB(db.DB).C("usermail").
		Find(bson.M{"status": NotReadUserMail, "userid": uid, "mailservicetype": mailServiceType}).
		Sort("-createdat").Limit(limit).All(notRead)
	if err != nil {
		log.Error(err.Error())
	}
	*rt = append(*rt, *notRead...)
	if len(*notRead) < conf.GetCfgHall().UserMailLimit {
		limit -= len(*notRead)
		readed := new([]UserMail)
		err = se.DB(db.DB).C("usermail").
			Find(bson.M{"$or": []bson.M{{"status": ReadUserMail}, {"status": TakenUserMail}}, "expiredat": bson.M{"$gt": time.Now().Unix()}, "userid": uid, "mailservicetype": mailServiceType}).
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
		temp.MailType = v.MailType
		temp.MailServiceType = v.MailServiceType
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
		temp.PropType = v.PropType
		temp.Imgurl = config.GetPropBaseConfig(v.PropType).ImgUrl
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
	TargetID        int64   `json:"target_id"`         //目标用户， -1表示所有用户
	MailType        int     `json:"mail_type"`         //邮箱邮件类型
	MailServiceType int     `json:"mail_service_type"` //邮件服务类型
	Title           string  `json:"title"`             //主题
	Content         string  `json:"content"`           //内容
	Annexes         []Annex `json:"annexes"`           //附件
	ExpireValue     float64 `json:"expire_value"`      //有效时长
}

func (ctx *MailBoxParam) PushMailBox() {
	mailBox := new(MailBox)
	mailBox.TargetID = ctx.TargetID
	mailBox.ExpireValue = ctx.ExpireValue
	mailBox.MailType = ctx.MailType
	mailBox.MailServiceType = ctx.MailServiceType
	mailBox.Title = ctx.Title
	mailBox.Content = ctx.Content
	mailBox.Annexes = ctx.Annexes

	mailBox.pushMailBox()
}

func TakenMailAnnex(mid int64) {
	mail := readUserMailByID(mid)
	if mail.Status == TakenUserMail {
		return
	}
	mail.Status = TakenUserMail
	mail.save()

	var ud *player.UserData
	user, ok := player.UserIDUsers[int(mail.UserID)]
	if !ok {
		ud = player.ReadUserDataByID(int(mail.UserID))
	} else {
		ud = user.GetUserData()
	}
	for _, v := range mail.Annexes {
		AddSundries(v.PropType, ud, v.Num, db.MailOpt, db.Mail, "")
	}
}

func TakenAllMail(user *player.User) {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	usermails := new([]UserMail)
	err := se.DB(db.DB).C("usermail").Find(bson.M{"status": bson.M{"$ne": TakenUserMail}, "userid": user.UID()}).All(usermails)
	if err != nil {
		log.Error(err.Error())
	}
	for _, usermail := range *usermails {
		switch usermail.MailType {
		case MailTypeText:
			//ReadMail(usermail.ID)
		case MailTypeAward:
			TakenMailAnnex(usermail.ID)
		case MailTypeMix:
			TakenMailAnnex(usermail.ID)
		}
	}
	SendMail(user)
	user.WriteMsg(&msg.S2C_TakenAllMail{})
}

func DeleteAllMail(user *player.User) {
	se := db.MongoDB.Ref()
	defer db.MongoDB.UnRef(se)
	usermails := new([]UserMail)
	err := se.DB(db.DB).C("usermail").Find(bson.M{"userid": user.UID(), "$or": []bson.M{
		{"mailtype": MailTypeText, "status": ReadUserMail},
		{"mailtype": MailTypeAward, "status": TakenUserMail},
		{"mailtype": MailTypeMix, "status": TakenUserMail},
	}}).All(usermails)
	if err != nil {
		log.Error(err.Error())
	}
	for _, usermail := range *usermails {
		DeleteMail(usermail.ID)
	}

	SendMail(user)
	user.WriteMsg(&msg.S2C_DeleteAllMail{})
}
