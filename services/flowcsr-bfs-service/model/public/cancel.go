package public

import (
	"bigrule/common/logger"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"bigrule/services/flowcsr-bfs-service/model/response"
	"encoding/json"
	"fmt"
)

func Cancel(token string) {
	pars := map[string]interface{}{}
	url := fmt.Sprintf("http://%s/v1/public/cancel", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	resp, err := middleware.PostUrl(pars, url, headers)
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
