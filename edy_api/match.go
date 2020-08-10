package edy_api

import (
	"ddz/edy_api/internal/base"
	"ddz/game/values"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/szxby/tools/log"
)

func checkCode(data []byte) error {
	log.Debug("data:%v", string(data))
	tmp := map[string]interface{}{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		log.Error("err:%v", err)
		return err
	}
	if tmp["resp_code"] == nil {
		log.Error("err ret:%+v", tmp)
		return errors.New("unknow err")
	}
	s, ok := tmp["resp_code"].(string)
	if !ok {
		log.Error("err ret:%+v", tmp)
		return errors.New("unknow err")
	}
	if s != "000000" {
		log.Error("err ret:%+v", tmp)
		return fmt.Errorf("err:%v", s)
	}
	return nil
}

// CheckMatch 验证比赛有效性
func CheckMatch(matchID string) ([]byte, error) {
	c := base.NewClient("/edy/match/check", fmt.Sprintf("cp_id=%v&match_id=%v", base.CpID, matchID), base.ReqGet)
	c.GenerateSign(base.ReqGet)
	ret, err := c.DoGet()
	if err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	if err := checkCode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// SignMatch 报名比赛
func SignMatch(matchID, name, uid string) ([]byte, error) {
	data := struct {
		Cp_id       string `json:"cp_id"`
		Match_id    string `json:"match_id"`
		Player_name string `json:"player_name"`
		Player_id   string `json:"player_id"`
	}{
		Cp_id:       base.CpID,
		Match_id:    matchID,
		Player_name: name,
		Player_id:   uid,
	}
	str, _ := json.Marshal(data)
	log.Debug("str:%v", string(str))
	c := base.NewClient("/edy/match/reg", string(str), base.ReqPost)
	c.GenerateSign(base.ReqPost)
	ret, err := c.DoPost()
	if err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	if err := checkCode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// SendMatchResultWithRobot 人机对局结果上报
func SendMatchResultWithRobot(data values.SportsCenterReportRobot) ([]byte, error) {
	data.Cp_id = base.CpID
	str, _ := json.Marshal(data)
	log.Debug("str:%v", string(str))
	c := base.NewClient("/edy/match/eview/round/send", string(str), base.ReqPost)
	c.GenerateSign(base.ReqPost)
	ret, err := c.DoPost()
	if err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	if err := checkCode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// SendMatchResultWithPerson 人人对局结果上报
func SendMatchResultWithPerson(data values.SportsCenterReportPersonal) ([]byte, error) {
	data.Cp_id = base.CpID
	str, _ := json.Marshal(data)
	log.Debug("str:%v", string(str))
	c := base.NewClient("/edy/match/renn/round/send", string(str), base.ReqPost)
	c.GenerateSign(base.ReqPost)
	ret, err := c.DoPost()
	if err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	if err := checkCode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// RoundRankReport 轮次排名上报
func RoundRankReport(data values.SportsCenterRankResult) ([]byte, error) {
	data.Cp_id = base.CpID
	str, _ := json.Marshal(data)
	log.Debug("str:%v", string(str))
	c := base.NewClient("/edy/match/round/rank/send", string(str), base.ReqPost)
	c.GenerateSign(base.ReqPost)
	ret, err := c.DoPost()
	if err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	if err := checkCode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// FinalRankReport  最终排名上报
func FinalRankReport(data values.SportsCenterFinalRankResult) ([]byte, error) {
	data.Cp_id = base.CpID
	str, _ := json.Marshal(data)
	log.Debug("str:%v", string(str))
	c := base.NewClient("/edy/match/ranks/send", string(str), base.ReqPost)
	c.GenerateSign(base.ReqPost)
	ret, err := c.DoPost()
	if err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	if err := checkCode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// RankReportFinish  排名上报完毕
func RankReportFinish(matchID string) ([]byte, error) {
	data := struct {
		Cp_id    string `json:"cp_id"`
		Match_id string `json:"match_id"`
	}{
		Cp_id:    base.CpID,
		Match_id: matchID,
	}
	str, _ := json.Marshal(data)
	log.Debug("str:%v", string(str))
	c := base.NewClient("/edy/match/ranks/sent", string(str), base.ReqPost)
	c.GenerateSign(base.ReqPost)
	ret, err := c.DoPost()
	if err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	if err := checkCode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// HXMatchPromotionReport  海选赛晋级名单上报
func HXMatchPromotionReport(matchID, promotionMatchID string, players []int) ([]byte, error) {
	type PlayerID struct {
		Player_id string `json:"player_id"`
	}
	tmpPlayerID := []PlayerID{}
	for _, p := range players {
		tmpPlayerID = append(tmpPlayerID, PlayerID{Player_id: strconv.Itoa(p)})
	}
	data := struct {
		Cp_id                 string     `json:"cp_id"`
		Match_id              string     `json:"match_id"`
		Promotion_match_id    string     `json:"promotion_match_id"`
		Promotion_player_list []PlayerID `json:"promotion_player_list"`
	}{
		Cp_id:                 base.CpID,
		Match_id:              matchID,
		Promotion_match_id:    promotionMatchID,
		Promotion_player_list: tmpPlayerID,
	}
	str, _ := json.Marshal(data)
	log.Debug("str:%v", string(str))
	c := base.NewClient("/edy/match/promotion/ranks/send", string(str), base.ReqPost)
	c.GenerateSign(base.ReqPost)
	ret, err := c.DoPost()
	if err != nil {
		log.Error("err:%v", err)
		return nil, err
	}
	if err := checkCode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}
