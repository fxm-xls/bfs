package rules

import (
	"bigrule/common/ico"
	"bigrule/common/logger"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

type RuleRegexQuery struct {
	Type      int         `json:"type"`       // 是否分页（0：分页；1：不分页）
	DataType  int         `json:"data_type"`  // 1：attributes+pattern
	PageSize  int         `json:"page_size"`  // 分页选填
	PageIndex int         `json:"page_index"` // 分页选填
	Limit     int         `json:"limit"`      // 不分页选填
	RepoId    int         `json:"repo_id"`
	RuleIds   []int       `json:"rule_ids"`
	RegexLike interface{} `json:"regex_like"`
}

func (This RuleRegexQuery) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("规则体查询")
	repoDataList, code, err := This.GetRuleRegexData(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	return ico.Succ(repoDataList)
}

// 识别规则体
type RuleRegexRes struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (This RuleRegexQuery) GetRuleRegexData(token string) (resData interface{}, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{
		"type": This.Type, "data_type": This.DataType, "page_size": This.PageSize, "page_index": This.PageIndex,
		"limit": This.Limit, "repo_id": This.RepoId, "rule_ids": This.RuleIds, "regex_like": This.RegexLike,
	}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/rule/regex/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := RuleRegexRes{}
	err = json.Unmarshal(respRepo, &res)
	if err != nil {
		logger.Info("数据查询失败", string(respRepo), pars)
		return resData, 2301, errors.New("数据查询失败")
	}
	if res.Code != 200 {
		logger.Info("数据查询失败")
		return resData, res.Code, errors.New(res.Message)
	}
	resData = res.Data
	return
}
