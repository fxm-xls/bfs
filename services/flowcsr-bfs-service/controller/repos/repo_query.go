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

type RepoQuery struct {
	RepoType int  `json:"repo_type"`
	RepoSum  bool `json:"repo_sum"`
}

type RepoTypeQueryRes struct {
	RepoType int            `json:"repo_type"`
	List     []RepoTypeInfo `json:"list"`
}

type RepoTypeInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"name_en"`
	Sum  int    `json:"sum"`
}

func (This RepoQuery) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("规则库清单查询")
	res := []RepoTypeQueryRes{}
	// 1.识别规则库查询
	if This.RepoType == 1 || This.RepoType == 0 {
		repoTypeQueryRes := RepoTypeQueryRes{RepoType: 1}
		// 1.1 查询规则库
		repoDataList, code, err := This.GetRepoData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		repoIds := []int{}
		for _, repoData := range repoDataList.List {
			repoTypeQueryRes.List = append(repoTypeQueryRes.List, RepoTypeInfo{
				Id: repoData.RepoId, Name: repoData.Name, Desc: repoData.Desc,
			})
			repoIds = append(repoIds, repoData.RepoId)
		}
		// 1.2 查询规则数量
		if This.RepoSum {
			repoCountList, code, err := This.GetRepoCountData(token, repoIds)
			if err != nil {
				return ico.Err(code, err.Error())
			}
			for _, repoCount := range repoCountList {
				for i, repoType := range repoTypeQueryRes.List {
					if repoType.Id == repoCount.RepoId {
						repoTypeQueryRes.List[i].Sum = repoCount.RuleNumber
						break
					}
				}
			}
		}
		res = append(res, repoTypeQueryRes)
	}
	// 2.解析规则库查询
	if This.RepoType == 2 || This.RepoType == 0 {
		repoTypeQueryRes := RepoTypeQueryRes{RepoType: 2}
		// 2.1 查询规则库
		repoDataList, code, err := This.GetParserRepoData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		repoIds := []int{}
		for _, repoData := range repoDataList.List {
			repoTypeQueryRes.List = append(repoTypeQueryRes.List, RepoTypeInfo{
				Id: repoData.ParserRepoId, Name: repoData.Name, Desc: repoData.Desc,
			})
			repoIds = append(repoIds, repoData.ParserRepoId)
		}
		// 2.2 查询规则数量
		if This.RepoSum {
			repoCountList, code, err := This.GetParserRepoCountData(token, repoIds)
			if err != nil {
				return ico.Err(code, err.Error())
			}
			for _, repoCount := range repoCountList {
				for i, repoType := range repoTypeQueryRes.List {
					if repoType.Id == repoCount.RepoId {
						repoTypeQueryRes.List[i].Sum = repoCount.RuleNumber
						break
					}
				}
			}
		}
		res = append(res, repoTypeQueryRes)
	}
	// 3.关联规则库查询
	if This.RepoType == 3 || This.RepoType == 0 {
		repoTypeQueryRes := RepoTypeQueryRes{RepoType: 3}
		// 3.1 查询规则库
		repoDataList, code, err := This.GetMappingData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		for _, repoData := range repoDataList.List {
			repoTypeInfo := RepoTypeInfo{Id: repoData.Id, Name: repoData.Name, Desc: repoData.Desc}
			// 3.2 查询规则数量
			if This.RepoSum {
				repoCountList, code, err := This.GetMappingCountData(token, repoData.Id)
				if err != nil {
					return ico.Err(code, err.Error())
				}
				repoTypeInfo.Sum = repoCountList.Count
			}
			repoTypeQueryRes.List = append(repoTypeQueryRes.List, repoTypeInfo)
		}
		res = append(res, repoTypeQueryRes)
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
	RepoId  int        `json:"repo_id"`
	Name    string     `json:"name"`
	Desc    string     `json:"description"`
	AttrMsg []AttrInfo `json:"attr_msg"`
}

type AttrInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (This RepoQuery) GetRepoData(token string) (resData RepoQueryRes, code int, err error) {
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

// 识别规则库数量
type RepoCountRes struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    []RepoCountQueryRes `json:"data"`
}

type RepoCountQueryRes struct {
	RepoId     int `json:"repo_id"`
	RuleNumber int `json:"rule_number"`
}

func (This RepoQuery) GetRepoCountData(token string, repoIds []int) (resData []RepoCountQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_ids": repoIds}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/count", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := RepoCountRes{}
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

// 解析规则库
type ParserRes struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    ParserQueryRes `json:"data"`
}

type ParserQueryRes struct {
	List []ParserInfo `json:"list"`
}

type ParserInfo struct {
	ParserRepoId int        `json:"parser_repo_id"`
	Name         string     `json:"name"`
	Desc         string     `json:"description"`
	AttrMsg      []AttrInfo `json:"attr_msg"`
}

func (This RepoQuery) GetParserRepoData(token string) (resData ParserQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"type": 1}
	urlRepo := fmt.Sprintf("http://%s/v1/parser-repo/query", middleware.GetAddr("repo-service"))
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

// 解析规则库数量
type ParserRepoCountRes struct {
	Code    int                       `json:"code"`
	Message string                    `json:"message"`
	Data    []ParserRepoCountQueryRes `json:"data"`
}

type ParserRepoCountQueryRes struct {
	RepoId     int `json:"repo_id"`
	RuleNumber int `json:"rule_number"`
}

func (This RepoQuery) GetParserRepoCountData(token string, repoIds []int) (resData []ParserRepoCountQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"parser_repo_ids": repoIds}
	urlRepo := fmt.Sprintf("http://%s/v1/parser-repo/count", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := ParserRepoCountRes{}
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

// 关联规则库
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
	Desc   string  `json:"description"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (This RepoQuery) GetMappingData(token string) (resData MappingQueryRes, code int, err error) {
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

// 关联规则库数量
type MappingCountRes struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    MappingCountQueryRes `json:"data"`
}

type MappingCountQueryRes struct {
	Count int `json:"count"`
}

func (This RepoQuery) GetMappingCountData(token string, dictId int) (resData MappingCountQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"page_size": 1, "page_index": 1, "dict_id": dictId}
	urlRepo := fmt.Sprintf("http://%s/v1/dictionary/record/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := MappingCountRes{}
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
