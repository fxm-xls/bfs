package public

import (
	"bigrule/common/logger"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"bigrule/services/flowcsr-bfs-service/model/response"
	"encoding/json"
	"fmt"
)

func Commit(token, message string) {
	pars := map[string]interface{}{"message": message}
	url := fmt.Sprintf("http://%s/v1/public/public", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	resp, err := middleware.GetUrl(pars, url, headers)
	if err != nil {
		logger.Error("请求异常")
		return
	}
	res := response.CsrRes{}
	if err = json.Unmarshal(resp, &res); err != nil {
		logger.Error(string(resp), " json解析异常")
		return
	}
}
