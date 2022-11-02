package rules

import (
	"bigrule/common/ico"
	"bigrule/common/logger"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"bigrule/services/flowcsr-bfs-service/model/response"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

type RuleQuery struct {
	RepoIds []int `json:"repo_ids"`
	RuleIds []int `json:"rule_ids"    binding:"required"`
}

type RuleQueryRes struct {
	RepoId     int                `json:"repo_id"`
	RuleId     int                `json:"rule_id"`
	Attributes response.Attribute `json:"attributes"`
	Dimensions []Dimension        `json:"dimensions"`
}

func (This RuleQuery) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("规则查询")
	// 1.获取规则库
	repoIds := []int{}
	if len(This.RepoIds) == 0 {
		repoDataList, code, err := This.GetRepoData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		for _, repoData := range repoDataList.List {
			repoIds = append(repoIds, repoData.RepoId)
		}
	} else {
		repoIds = This.RepoIds
	}
	// 2.获取规则信息
	res := []RuleQueryRes{}
	for _, repoId := range repoIds {
		ruleDataList, code, err := This.GetRuleData(token, repoId)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		for _, ruleData := range ruleDataList.List {
			statusInt, _ := strconv.Atoi(ruleData.Attributes.Status)
			priorityInt, _ := strconv.Atoi(ruleData.Attributes.Priority)
			res = append(res, RuleQueryRes{
				RepoId: repoId, RuleId: ruleData.RuleId, Dimensions: ruleData.Dimensions,
				Attributes: response.Attribute{Status: statusInt, Priority: priorityInt, Desc: ruleData.Attributes.Desc, Sample: ruleData.Attributes.Sample},
			})
		}
	}
	return ico.Succ(res)
}

// 识别规则库
type RepoRes struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    RepoQueryRes `json:"data"`
}

type RepoQueryRes struct {
	List []RepoInfo `json:"list"`
}

type RepoInfo struct {
	RepoId       int             `json:"repo_id"`
	Name         string          `json:"name"`
	ParserMsg    ParserRepoInfo  `json:"parser_msg"`
	DimensionMsg []DimensionInfo `json:"dimension_msg"`
}

type ParserRepoInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type DimensionInfo struct {
	DimensionId int    `json:"dimension_id"`
	Name        string `json:"name"`
}

func (This RuleQuery) GetRepoData(token string) (resData RepoQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_ids": This.RepoIds, "type": 1}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := RepoRes{}
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

// 识别规则
type RuleRes struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    RuleListRes `json:"data"`
}

type RuleListRes struct {
	List []RuleInfo `json:"list"`
}

type RuleInfo struct {
	RuleId        int                   `json:"rule_id"`
	LineNum       int                   `json:"line_num"`
	ParserMessage ParserInfo            `json:"parser_message"`
	Attributes    response.AttributeRes `json:"attributes"`
	Dimensions    []Dimension           `json:"dimensions"`
}

type Dimension struct {
	TagId         int    `json:"tag_id"`
	TagName       string `json:"tag_name"`
	DimensionId   int    `json:"dimension_id"`
	DimensionName string `json:"dimension_name"`
	TagValTblId   int    `json:"tagval_tbl_id"`
}

type ParserInfo struct {
	ParserId    int    `json:"parser_id"`
	Description string `json:"description"`
}

func (This RuleQuery) GetRuleData(token string, repoId int) (resData RuleListRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_id": repoId, "type": 1, "rule_ids": This.RuleIds}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/rule/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := RuleRes{}
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
