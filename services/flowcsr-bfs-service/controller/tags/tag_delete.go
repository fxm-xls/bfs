package tags

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

type TagDelete struct {
	TagValTblId int    `json:"tagval_tbl_id"     binding:"required"`
	TagId       int    `json:"tag_id"            binding:"required"`
	RuleList    []Rule `json:"rule_list"`
}

type Rule struct {
	RepoId     int          `json:"repo_id"`
	RuleId     int          `json:"rule_id"`
	ParserList []ParserRepo `json:"parser_list"`
}

type ParserRepo struct {
	ParserId     int `json:"parser_id"`
	RepoParserId int `json:"repo_parser_id"`
}

func (This TagDelete) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("标签删除")
	// 0.权限判断
	permissionToken, code, err := middleware.GetPermission(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	if !utils.IsContainsInt(permissionToken.TagValIds, This.TagValTblId) {
		return ico.Err(2007, "权限不足")
	}
	for _, repo := range This.RuleList {
		if !utils.IsContainsInt(permissionToken.RepoIds, repo.RepoId) {
			return ico.Err(2007, "权限不足")
		}
		for _, parser := range repo.ParserList {
			if !utils.IsContainsInt(permissionToken.ParserIds, parser.RepoParserId) {
				return ico.Err(2007, "权限不足")
			}
		}
	}
	// 1.通过标签表获取维度和规则库
	dimDataList, code, err := This.GetDimData(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	// 2.通过规则库、维度、标签获取规则信息
	for _, dimData := range dimDataList.List {
		ruleData, code, err := This.GetRuleData(token, dimData.RepoId, dimData.DimensionId)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		// 3.解绑标签
		for _, rule := range ruleData.List {
			code, err = This.DeleteRuleTag(token, dimData.RepoId, rule.RuleId, dimData.DimensionId)
			if err != nil {
				return ico.Err(code, err.Error())
			}
		}
	}
	// 4.删除规则
	message := fmt.Sprint("标签删除： ")
	messageRule := " 删除规则：["
	messageParser := " 删除解析规则：["
	for _, rule := range This.RuleList {
		// 4.1 删除识别规则，只能有一条
		ruleData, code, err := This.QueryRuleData(token, rule.RepoId, rule.RuleId)
		if err != nil {
			public.Cancel(token)
			return ico.Err(code, err.Error())
		}
		if len(ruleData.List) != 1 {
			logger.Info(ruleData.List)
			public.Cancel(token)
			return ico.Err(2301, "该规则有多条重复id")
		}
		if code, err = This.DeleteRule(token, rule.RepoId, rule.RuleId, ruleData.List[0].LineNum); err != nil {
			public.Cancel(token)
			return ico.Err(code, err.Error())
		}
		messageRule += fmt.Sprint(rule.RuleId)
		// 4.2 删除解析规则，只能有一条
		for _, parser := range rule.ParserList {
			parserData, code, err := This.GetParserData(token, parser.ParserId, parser.RepoParserId)
			if err != nil {
				public.Cancel(token)
				return ico.Err(code, err.Error())
			}
			if len(parserData.List) != 1 {
				public.Cancel(token)
				return ico.Err(2301, "该解析规则有多条重复id")
			}
			if code, err = This.DeleteParser(token, parser.ParserId, parser.RepoParserId, parserData.List[0].LineNum); err != nil {
				public.Cancel(token)
				return ico.Err(code, err.Error())
			}
			messageParser += fmt.Sprint(parser.ParserId)
		}
	}
	message += messageRule + "]"
	message += messageParser + "]"
	// 5.删除标签
	tagRes, code, err := This.GetTagData(token)
	if err != nil {
		public.Cancel(token)
		return ico.Err(code, err.Error())
	}
	if len(tagRes.List) > 1 {
		public.Cancel(token)
		return ico.Err(2301, "有多条重复id标签")
	}
	if code, err := This.DeleteTag(token, tagRes.List[0].LineNum); err != nil {
		public.Cancel(token)
		return ico.Err(code, err.Error())
	}
	message += fmt.Sprintf(" 删除标签：[%d]", This.TagId)
	logger.Info(message)
	public.Commit(token, message)
	return ico.Succ("删除成功")
}

func (This *TagDelete) GetDimData(token string) (resData DimQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"tagval_tbl_ids": []int{This.TagValTblId}, "type": 1}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/dimension/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := DimRes{}
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

func (This TagDelete) GetRuleData(token string, repoId, dimId int) (resData RuleQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{
		"repo_id": repoId, "type": 1, "data_type": 3,
		"dimensions": []map[string]interface{}{{"dimension_id": dimId, "tagval_ids": []int{This.TagId}}},
	}
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

func (This TagDelete) QueryRuleData(token string, repoId, ruleId int) (resData RuleQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_id": repoId, "type": 1, "data_type": 3, "rule_ids": []int{ruleId}}
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

// 解析规则
type ParserRes struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    ParserQueryRes `json:"data"`
}

type ParserQueryRes struct {
	List []ParserQueryInfo `json:"list"`
}

type ParserQueryInfo struct {
	ParserId     int `json:"parser_id"`
	LineNum      int `json:"line_num"`
	ParserRepoId int `json:"parser_repo_id"`
}

func (This TagDelete) GetParserData(token string, repoId, parserId int) (resData ParserQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"parser_repo_id": repoId, "type": 1, "parser_id": parserId}
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

// 规则标签
type RuleTagRes struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    TagData `json:"data"`
}

func (This TagDelete) DeleteRuleTag(token string, repoId, ruleId, dimId int) (code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{
		"repo_id": repoId, "rule_id": ruleId, "dimensions": []map[string]interface{}{{"dimension_id": dimId}},
	}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/dimension/link-unlink", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return 2301, errors.New("数据查询失败")
	}
	res := RuleTagRes{}
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

func (This TagDelete) DeleteRule(token string, repoId, ruleId, lineNum int) (code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_id": repoId, "rule_id": ruleId, "line_num": lineNum}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/rule/delete", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return 2301, errors.New("数据查询失败")
	}
	res := RuleTagRes{}
	err = json.Unmarshal(respRepo, &res)
	if err != nil {
		logger.Info("数据查询失败", string(respRepo), pars)
		return 2301, errors.New("数据查询失败")
	}
	if res.Code != 200 {
		logger.Info("数据查询失败")
		logger.Info(pars)
		return res.Code, errors.New(res.Message)
	}
	return
}

func (This TagDelete) DeleteParser(token string, repoId, parserId, lineNum int) (code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"parser_repo_id": repoId, "parser_id": parserId, "line_num": lineNum}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/parser/delete", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return 2301, errors.New("数据查询失败")
	}
	res := RuleTagRes{}
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

func (This TagDelete) GetTagData(token string) (resData TagData, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"tagval_tbl_id": This.TagValTblId, "type": 1, "tagval_ids": []int{This.TagId}}
	urlRepo := fmt.Sprintf("http://%s/v1/tag/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := TagRes{}
	err = json.Unmarshal(respRepo, &res)
	if err != nil {
		logger.Info("数据查询失败", string(respRepo), pars)
		return resData, 2301, errors.New("数据查询失败")
	}
	if res.Code != 200 || len(res.Data.List) == 0 {
		logger.Info("数据查询失败")
		return resData, res.Code, errors.New(res.Message)
	}
	resData = res.Data
	return
}

func (This TagDelete) DeleteTag(token string, lineNum int) (code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"tagval_tbl_id": This.TagValTblId, "line_nums": []int{lineNum}, "tagval_ids": []int{This.TagId}}
	urlRepo := fmt.Sprintf("http://%s/v1/tag/delete", middleware.GetAddr("repo-service"))
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
