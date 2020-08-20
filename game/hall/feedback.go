package hall

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/msg"
	"github.com/name5566/leaf/log"
	"time"
)

type FeedBack struct {
	ID        int `bson:"_id"`
	AccountID int
	Title     string
	Content   string
	PhoneNum  string //联系方式
	Nickname  string //昵称
	CreatedAt int64
	UpdatedAt int64
	DeletedAt int64
}

func Feedback(user *player.User, title, content string) {
	if title == "" || content == "" {
		user.WriteMsg(&msg.S2C_FeedBack{
			Error: msg.S2C_FeedBaock_Fail,
		})
		return
	}

	feedBack := new(FeedBack)
	feedBack.ID, _ = db.MongoBkDBNextSeq("feedback")
	feedBack.PhoneNum = user.GetUserData().Username
	feedBack.Nickname = user.GetUserData().Nickname
	feedBack.CreatedAt = time.Now().Unix()
	feedBack.AccountID = user.AcountID()
	feedBack.Title = title
	feedBack.Content = content

	saveFeedBack(feedBack)
	user.WriteMsg(&msg.S2C_FeedBack{
		Error: msg.S2C_FeedBaock_OK,
	})
	return
}

func saveFeedBack(feedBack *FeedBack) {
	game.GetSkeleton().Go(func() {
		se := db.BackstageDB.Ref()
		defer db.BackstageDB.UnRef(se)

		if err := se.DB(db.BkDBName).C("feedback").Insert(feedBack); err != nil {
			log.Error("用户反馈数据存储错误：error：" + err.Error())
		}
	}, nil)
}
