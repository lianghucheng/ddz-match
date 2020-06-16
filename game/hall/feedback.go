package hall

import (
	"ddz/game"
	"ddz/game/db"
	"ddz/game/player"
	"ddz/msg"
	"time"
)

type FeedBack struct {
	UserID    int
	Title     string
	Content   string
	CreatedAt int64
}

func Feedback(user *player.User, title, content string) {
	if title == "" || content == "" {
		user.WriteMsg(&msg.S2C_FeedBack{
			Error: msg.S2C_FeedBaock_Fail,
		})
		return
	}

	feedBack := new(FeedBack)
	feedBack.CreatedAt = time.Now().Unix()
	feedBack.UserID = user.BaseData.UserData.UserID
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
		se := db.MongoDB.Ref()
		defer db.MongoDB.UnRef(se)

		se.DB(db.DB).C("feedback").Insert(feedBack)
	}, nil)
}
