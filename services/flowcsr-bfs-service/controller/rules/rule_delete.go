package rules

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

type RuleDelete struct {
	RepoId   int   `json:"repo_id"   binding:"required"`
	RuleIds  []int `json:"rule_ids"  binding:"required"`
	TagOp    int   `json:"tag_op"    binding:"required"`
	lineNums []int
}

func (This RuleDelete) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("规则删除")
	// 0.权限判断
	permissionToken, code, err := middleware.GetPermission(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	if !utils.IsContainsInt(permissionToken.RepoIds, This.RepoId) {
		return ico.Err(2007, "权限不足")
	}
	// 1.规则查询
	ruleDataList, code, err := This.GetRuleData(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	message := fmt.Sprint("规则删除： ")
	// 2.查询标签
	deleteTagRes := []DeleteTag{}
	if This.TagOp == 2 {
		// 2.1 多条重复id判断
		if len(ruleDataList.List) != len(This.RuleIds) {
			public.Cancel(token)
			return ico.Err(2301, "有多条重复id规则")
		}
		// 2.2 获取标签
		deleteTagRes, code, err = This.GetTag(token, ruleDataList)
		if err != nil {
			public.Cancel(token)
			return ico.Err(code, err.Error())
		}
	}
	// 3.删除规则
	messageRule := " 删除规则：["
	for _, ruleData := range ruleDataList.List {
		if code, err := This.DeleteRule(token, ruleData.RuleId, ruleData.LineNum); err != nil {
			public.Cancel(token)
			return ico.Err(code, err.Error())
		}
		messageRule += fmt.Sprint(ruleData.RuleId)
	}
	message += messageRule + "]"
	// 4.删除标签
	if len(deleteTagRes) != 0 {
		messageTag := " 删除标签：["
		for _, deleteTag := range deleteTagRes {
			logger.Info(deleteTag)
			if code, err := This.DeleteTag(token, deleteTag); err != nil {
				public.Cancel(token)
				return ico.Err(code, err.Error())
			}
			messageTag += fmt.Sprint(deleteTag.TagvalIds)
		}
		message += messageTag + "]"
	}
	logger.Info(message)
	public.Commit(token, message)
	return ico.Succ("删除成功")
}

type DeleteTag struct {
	TagValTblId int   `json:"tagval_tbl_id"`
	TagvalIds   []int `json:"tagval_ids"`
	LineNums    []int `json:"line_nums"`
}

func (This RuleDelete) GetTag(token string, ruleDataList RuleListRes) (deleteTagRes []DeleteTag, code int, err error) {
	// 1.获取标签表，id
	// 多条规则
	for _, ruleData := range ruleDataList.List {
		// 单条规则的多个维度
		for _, dim := range ruleData.Dimensions {
			// 增加标签
			temp := true
			for i, deleteTag := range deleteTagRes {
				if dim.TagValTblId == deleteTag.TagValTblId {
					temp = false
					deleteTagRes[i].TagvalIds = append(deleteTagRes[i].TagvalIds, dim.TagId)
				}
			}
			// 增加表
			if temp {
				deleteTagRes = append(deleteTagRes, DeleteTag{TagValTblId: dim.TagValTblId, TagvalIds: []int{dim.TagId}})
			}
		}
	}
	// 2.获取标签行号
	for i, deleteTag := range deleteTagRes {
		tagDataList, code, err := This.GetTagData(token, deleteTag.TagValTblId, deleteTag.TagvalIds)
		if err != nil {
			logger.Info("数据查询失败")
			return deleteTagRes, code, errors.New("数据查询失败")
		}
		for _, tagData := range tagDataList.List {
			deleteTagRes[i].LineNums = append(deleteTagRes[i].LineNums, tagData.LineNum)
		}
	}
	return
}

func (This RuleDelete) GetRuleData(token string) (resData RuleListRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"data_type": 2, "repo_id": This.RepoId, "type": 1, "rule_ids": This.RuleIds}
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

// 标签
type TagRes struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    TagData `json:"data"`
}

type TagData struct {
	List []TagInfo `json:"list"`
}

type TagInfo struct {
	Id      int    `json:"id"`
	LineNum int    `json:"line_num"`
	Name    string `json:"name"`
}

func (This *RuleDelete) GetTagData(token string, tagValTblId int, tagIds []int) (resData TagData, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"tagval_tbl_id": tagValTblId, "type": 1, "tagval_ids": tagIds}
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

func (This *RuleDelete) DeleteTag(token string, deleteTag DeleteTag) (code int, err error) {
	// 1.整理入参
	pars := deleteTag
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

func (This *RuleDelete) DeleteRule(token string, ruleId, lineNum int) (code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_id": This.RepoId, "rule_id": ruleId, "line_num": lineNum}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/rule/delete", middleware.GetAddr("repo-service"))
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
