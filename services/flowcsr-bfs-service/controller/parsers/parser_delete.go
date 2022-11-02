package parsers

import (
	"bigrule/common/ico"
	"bigrule/common/logger"
	"bigrule/pkg/utils"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"bigrule/services/flowcsr-bfs-service/model/public"
	"bigrule/services/flowcsr-bfs-service/model/response"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

type ParserDelete struct {
	RepoParserId int   `json:"repo_parser_id"   binding:"required"`
	ParserIds    []int `json:"parser_ids"       binding:"required"`
	lineNums     []int
}

func (This ParserDelete) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("解析规则删除")
	// 0.权限判断
	permissionToken, code, err := middleware.GetPermission(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	if !utils.IsContainsInt(permissionToken.ParserIds, This.RepoParserId) {
		return ico.Err(2007, "权限不足")
	}
	token = permissionToken.Token
	// 1.解析规则查询
	for _, parserId := range This.ParserIds {
		parserDataList, code, err := This.GetParserData(token, parserId)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		// 1.1 多条重复id判断
		if len(parserDataList.List) > 1 {
			return ico.Err(2301, "有多条重复id规则")
		}
		This.lineNums = append(This.lineNums, parserDataList.List[0].LineNum)
	}
	// 2.解绑解析规则
	if code, err = This.UnLinkRule(token); err != nil {
		public.Cancel(token)
		return ico.Err(code, err.Error())
	}
	// 3.删除解析规则
	message := fmt.Sprintf("解析规则删除：[%s]", fmt.Sprint(This.ParserIds))
	for i, ParserId := range This.ParserIds {
		if code, err := This.DeleteRule(token, ParserId, This.lineNums[i]); err != nil {
			public.Cancel(token)
			return ico.Err(code, err.Error())
		}
	}
	logger.Info(message)
	public.Commit(token, message)
	return ico.Succ("删除成功")
}

func (This ParserDelete) UnLinkRule(token string) (code int, err error) {
	// 1.查询绑定的识别规则库
	repoDataList, code, err := This.GetRepoData(token)
	if err != nil {
		return
	}
	repoIds := []int{}
	for _, repo := range repoDataList.List {
		if repo.ParserMsg.Id == This.RepoParserId {
			repoIds = append(repoIds, repo.RepoId)
		}
	}
	// 2.查询绑定的识别规则
	for _, repoId := range repoIds {
		ruleDataList, code, err := This.GetRuleData(token, repoId)
		if err != nil {
			return code, err
		}
		ruleIds := []int{}
		for _, rule := range ruleDataList.List {
			if utils.IsContainsInt(This.ParserIds, rule.ParserMessage.ParserId) {
				ruleIds = append(ruleIds, rule.RuleId)
			}
		}
		// 3.解绑解析规则
		for _, ruleId := range ruleIds {
			code, err = This.UnLinkParser(token, repoId, ruleId)
			if err != nil {
				return code, err
			}
		}
	}
	return
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
	RepoId    int            `json:"repo_id"`
	Name      string         `json:"name"`
	Desc      string         `json:"description"`
	ParserMsg RepoParserInfo `json:"parser_msg"`
}

type RepoParserInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (This ParserDelete) GetRepoData(token string) (resData RepoQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"type": 1}
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
	RuleId        int            `json:"rule_id"`
	LineNum       int            `json:"line_num"`
	ParserMessage RuleParserInfo `json:"parser_message"`
}

type RuleParserInfo struct {
	ParserId    int    `json:"parser_id"`
	Description string `json:"description"`
}

func (This ParserDelete) GetRuleData(token string, repoId int) (resData RuleListRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_id": repoId, "type": 1, "data_type": 3}
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

func (This ParserDelete) UnLinkParser(token string, repoId, ruleId int) (code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_id": repoId, "rule_id": ruleId, "parser_repo_id": This.RepoParserId}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/parser/link-unlink", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return 2301, errors.New("数据查询失败")
	}
	res := response.CsrRes{}
	err = json.Unmarshal(respRepo, &res)
	if err != nil {
		logger.Info("数据查询失败", string(respRepo), pars)
		return 2301, errors.New("数据查询失败")
	}
	if res.Code != 200 {
		logger.Info("数据查询失败")
		return res.Code, errors.New(res.Message)
	}
	return
}

// 解析规则
type ParserRes struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    ParserListRes `json:"data"`
}

type ParserListRes struct {
	List []ParserInfo `json:"list"`
}

type ParserInfo struct {
	ParserId int `json:"parser_id"`
	LineNum  int `json:"line_num"`
}

func (This ParserDelete) GetParserData(token string, parserId int) (resData ParserListRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"parser_repo_id": This.RepoParserId, "type": 1, "parser_id": parserId}
	urlRepo := fmt.Sprintf("http://%s/v1/parser-repo/parser/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := ParserRes{}
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

func (This *ParserDelete) DeleteRule(token string, parserId, lineNum int) (code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"parser_repo_id": This.RepoParserId, "parser_id": parserId, "line_num": lineNum}
	urlRepo := fmt.Sprintf("http://%s/v1/parser-repo/parser/delete", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return 2301, errors.New("数据查询失败")
	}
	res := response.CsrRes{}
	err = json.Unmarshal(respRepo, &res)
	if err != nil {
		logger.Info("数据查询失败", string(respRepo), pars)
		return 2301, errors.New("数据查询失败")
	}
	if res.Code != 200 {
		logger.Info("数据查询失败")
		return res.Code, errors.New(res.Message)
	}
	return
}
