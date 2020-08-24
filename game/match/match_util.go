package match

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 获取除分数外的所有奖励
func getNoneScoreAward(award []string) string {
	ret := map[string]string{}
	for i, v := range award {
		s := strings.Split(v, ",")
		for _, one := range s {
			if strings.Index(one, "分") == -1 {
				str := ret[fmt.Sprintf("第%v名", i+1)]
				str += one + ","
				ret[fmt.Sprintf("第%v名", i+1)] = str
			}
		}
		str := ret[fmt.Sprintf("第%v名", i+1)]
		str = str[:len(str)-1]
		ret[fmt.Sprintf("第%v名", i+1)] = str
	}
	tmp, _ := json.Marshal(ret)
	return string(tmp)
}
