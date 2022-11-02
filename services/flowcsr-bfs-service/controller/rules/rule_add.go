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

type RuleAdd struct {
	RepoId   int    `json:"repo_id"      binding:"required"`
	RuleList []Rule `json:"rule_list"    binding:"required"`
	TagVal   TagVal `json:"tag"          binding:"required"`
}

type Rule struct {
	Id         int                `json:"id"          binding:"required"`
	Attr       response.Attribute `json:"attr"        binding:"required"`
	Pattern    map[string]string  `json:"pattern"     binding:"required"`
	Dimensions []DimensionAddRule `json:"dimensions"`
}

type DimensionAddRule struct {
	DimensionId int `json:"dimension_id"`
	ValueId     int `json:"value_id"`
}

type TagVal struct {
	TagValTblId int   `json:"tagval_tbl_id"       binding:"required"`
	Tag         []Tag `json:"data"                binding:"required"`
}

type Tag struct {
	Id    int    `json:"id"          binding:"required"`
	Value string `json:"name"        binding:"required"`
}

func (This RuleAdd) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("规则新增")
	// 0.权限判断
	permissionToken, code, err := middleware.GetPermission(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	if !utils.IsContainsInt(permissionToken.RepoIds, This.RepoId) {
		return ico.Err(2007, "权限不足")
	}
	if !utils.IsContainsInt(permissionToken.TagValIds, This.TagVal.TagValTblId) {
		return ico.Err(2007, "权限不足")
	}
	message := fmt.Sprint("批量规则增加： ")
	// 1.新增标签
	if code, err := This.AddTag(token); err != nil {
		public.Cancel(token)
		return ico.Err(code, err.Error())
	}
	messageTag := " 增加标签：["
	for _, tag := range This.TagVal.Tag {
		messageTag += fmt.Sprint(tag.Id) + " "
	}
	message += messageTag + "]"
	// 2.新增规则
	if code, err := This.AddRule(token); err != nil {
		public.Cancel(token)
		return ico.Err(code, err.Error())
	}
	messageRule := " 增加规则：["
	for _, rule := range This.RuleList {
		messageRule += fmt.Sprint(rule.Id) + " "
	}
	message += messageRule + "]"
	logger.Info(message)
	public.Commit(token, message)
	return ico.Succ("新增成功")
}

func (This *RuleAdd) AddTag(token string) (code int, err error) {
	// 1.整理入参
	pars := This.TagVal
	urlRepo := fmt.Sprintf("http://%s/v1/tag/add", middleware.GetAddr("repo-service"))
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

func (This *RuleAdd) AddRule(token string) (code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_id": This.RepoId, "messages": This.RuleList}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/rule/batchadd", middleware.GetAddr("repo-service"))
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
		logger.Info("数据查询失败", pars)
		return res.Code, errors.New(res.Message)
	}
	return
}
