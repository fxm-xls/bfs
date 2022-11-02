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

type RuleAttrQuery struct {
	Id int `json:"id"`
}

func (This RuleAttrQuery) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("规则属性查询")
	repoDataList, code, err := This.GetAttrData(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	return ico.Succ(repoDataList)
}

// 识别规则属性
type RuleAttrRes struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    []RuleAttr `json:"data"`
}

type RuleAttr struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (This RuleAttrQuery) GetAttrData(token string) (resData []RuleAttr, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"id": This.Id}
	urlRepo := fmt.Sprintf("http://%s/v1/public/attribute/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.GetUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := RuleAttrRes{}
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
