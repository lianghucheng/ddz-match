package match

import (
	"fmt"
	"strings"
)

// 获取除分数外的所有奖励
func getNoneScoreAward(award []string) string {
	if len(award) == 0 {
		return ""
	}
	ret := map[string]string{}
	retStr := `{`
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
		retStr += `"` + fmt.Sprintf("第%v名", i+1) + `"` + ":" + `"` + str + `"` + ","
	}
	retStr = retStr[:len(retStr)-1]
	retStr += "}"
	// tmp, _ := json.Marshal(ret)
	// return string(tmp)
	return retStr
}
