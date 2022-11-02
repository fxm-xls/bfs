package tags

import (
	"bigrule/common/ico"
	"bigrule/common/logger"
	"bigrule/services/flowcsr-bfs-service/middleware"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

type TagExport struct {
	RepoType int `json:"repo_type"     binding:"required"`
	RepoId   int `json:"repo_id"       binding:"required"`
}

type TagExportRes struct {
	RepoId   int             `json:"repo_id"`
	RepoName string          `json:"repo_name"`
	Rules    []ExportRule    `json:"rules"`
	RepoAttr []RepoAttribute `json:"repo_attr"`
}

type ExportRule struct {
	RuleId int         `json:"rule_id"`
	Tags   []ExportTag `json:"tags"`
	Attr   Attribute   `json:"attr"`
}

type ExportTag struct {
	TagKeyId int    `json:"tagkey_id"`
	TagId    int    `json:"tag_id"`
	TagName  string `json:"tag_name"`
}

type RepoAttribute struct {
	TagKeyId   int    `json:"tagkey_id"`
	TagKeyName string `json:"tagkey_name"`
}

func (This TagExport) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("标签导出")
	res := TagExportRes{RepoId: This.RepoId}
	// 识别规则库
	if This.RepoType == 1 {
		// 1.通过规则库获取维度和规则库名称
		repoDataList, code, err := This.GetRepoData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		for _, dim := range repoDataList.List[0].DimensionMsg {
			res.RepoAttr = append(res.RepoAttr, RepoAttribute{TagKeyId: dim.DimensionId, TagKeyName: dim.Name})
		}
		res.RepoName = repoDataList.List[0].Name
		// 2.通过规则库获取规则、标签、attr信息
		ruleData, code, err := This.GetRuleData(token, This.RepoId)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		for _, rule := range ruleData.List {
			exportRule := ExportRule{RuleId: rule.RuleId}
			for _, dim := range rule.Dimensions {
				exportRule.Tags = append(exportRule.Tags, ExportTag{TagKeyId: dim.DimensionId, TagId: dim.TagId, TagName: dim.TagName})
			}
			if len(exportRule.Tags) == 0 {
				exportRule.Tags = []ExportTag{}
			}
			statusInt, _ := strconv.Atoi(rule.Attributes.Status)
			priorityInt, _ := strconv.Atoi(rule.Attributes.Priority)
			exportRule.Attr = Attribute{Desc: rule.Attributes.Desc, Sample: rule.Attributes.Sample, Status: statusInt, Priority: priorityInt}
			res.Rules = append(res.Rules, exportRule)
		}
		return ico.Succ(res)
	}
	// 关联规则库
	if This.RepoType == 2 {
		// 1.通过规则库名称获取维度和规则库名称
		mappingDataList, code, err := This.GetMappingData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		for _, mapping := range mappingDataList.List {
			if mapping.Id == This.RepoId {
				res.RepoName = mapping.Name
				for _, field := range mapping.Fields {
					res.RepoAttr = append(res.RepoAttr, RepoAttribute{TagKeyId: field.Id, TagKeyName: field.Name})
				}
				break
			}
		}
		// 2.通过规则库获取规则、标签、attr信息
		return ico.Succ(res)
	}
	return ico.Err(2301, "类型异常")
}

func (This TagExport) GetRepoData(token string) (resData RepoQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_ids": []int{This.RepoId}, "type": 1}
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

func (This TagExport) GetRuleData(token string, repoId int) (resData RuleQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_id": repoId, "type": 1, "data_type": 2}
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

// 字典
type MappingRes struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    MappingQueryRes `json:"data"`
}

type MappingQueryRes struct {
	List []MappingInfo `json:"list"`
}

type MappingInfo struct {
	Id     int     `json:"id"`
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (This TagExport) GetMappingData(token string) (resData MappingQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"page_size": 10000, "page_index": 1}
	urlRepo := fmt.Sprintf("http://%s/v1/dictionary/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := MappingRes{}
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

func (This TagExport) GetMappingRuleData(token string, dictId int) (resData RuleQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"page_size": 10, "page_index": 1, "dict_id": dictId}
	urlRepo := fmt.Sprintf("http://%s/v1/dictionary/record/query", middleware.GetAddr("repo-service"))
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
