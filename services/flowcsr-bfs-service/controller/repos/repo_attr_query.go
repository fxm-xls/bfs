package repos

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

type RepoAttrQuery struct {
	RepoId int `json:"repo_id"`
}

func (This RepoAttrQuery) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("识别规则库属性清单查询")
	res := []RepoAttr{}
	if This.RepoId == 0 {
		// 1.识别规则库全部属性查询
		repoAttrDataList, code, err := This.GetRepoAttrData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		for _, repoAttr := range repoAttrDataList {
			res = append(res, RepoAttr{Id: repoAttr.Id, Name: repoAttr.Name})
		}
	} else {
		// 2.识别规则库属性查询
		repoDataList, code, err := This.GetRepoData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		for _, repoAttr := range repoDataList.List[0].AttrMsg {
			res = append(res, RepoAttr{Id: repoAttr.Id, Name: repoAttr.Name})
		}
	}
	return ico.Succ(res)
}

// 识别规则库属性
type RepoAttrRes struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    []RepoAttr `json:"data"`
}

type RepoAttr struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (This RepoAttrQuery) GetRepoAttrData(token string) (resData []RepoAttr, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{}
	urlRepo := fmt.Sprintf("http://%s/v1/public/attribute/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.GetUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := RepoAttrRes{}
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

func (This RepoAttrQuery) GetRepoData(token string) (resData RepoQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"type": 1, "repo_id": This.RepoId}
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
