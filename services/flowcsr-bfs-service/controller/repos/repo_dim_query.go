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

type RepoDimQuery struct {
	RepoType int `json:"repo_type"`
	RepoId   int `json:"repo_id"`
}

type RepoDimQueryRes struct {
	RepoType   int            `json:"repo_type"`
	RepoId     int            `json:"repo_id"`
	Dimensions []DimensionRes `json:"dimensions"`
}

type DimensionRes struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"name_en"`
}

func (This RepoDimQuery) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("规则库维度清单查询")
	res := []RepoDimQueryRes{}
	// 1.识别规则库维度查询
	if This.RepoType == 1 || This.RepoType == 0 {
		// 1.1 查询规则库维度
		repoDimList, code, err := This.GetRepoDimData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		if This.RepoId != 0 {
			// 1.2 单个规则库维度
			repoDimQueryRes := RepoDimQueryRes{RepoType: 1}
			for _, repoDim := range repoDimList.List {
				repoDimQueryRes.RepoId = This.RepoId
				if repoDim.RepoId == This.RepoId {
					repoDimQueryRes.Dimensions = append(repoDimQueryRes.Dimensions,
						DimensionRes{Id: repoDim.DimensionId, Name: repoDim.Name, Desc: repoDim.Description},
					)
				}
				continue
			}
			if len(repoDimQueryRes.Dimensions) > 0 {
				res = append(res, repoDimQueryRes)
			}
		} else {
			// 1.3 全部规则库维度
			for n, repoDim := range repoDimList.List {
				if n == 0 {
					res = append(res, RepoDimQueryRes{RepoId: repoDim.RepoId, RepoType: 1,
						Dimensions: []DimensionRes{{Id: repoDim.DimensionId, Name: repoDim.Name, Desc: repoDim.Description}},
					})
					continue
				}
				temp := false
				for i, repoRes := range res {
					if repoRes.RepoType == 1 && repoRes.RepoId == repoDim.RepoId {
						res[i].Dimensions = append(res[i].Dimensions,
							DimensionRes{Id: repoDim.DimensionId, Name: repoDim.Name, Desc: repoDim.Description},
						)
						temp = true
					}
				}
				if !temp {
					res = append(res, RepoDimQueryRes{RepoId: repoDim.RepoId, RepoType: 1,
						Dimensions: []DimensionRes{{Id: repoDim.DimensionId, Name: repoDim.Name, Desc: repoDim.Description}},
					})
				}
			}
		}
	}
	// 2.识别规则库维度查询
	if This.RepoType == 2 || This.RepoType == 0 {
		// 2.1 查询规则库维度
		repoDimList, code, err := This.GetParserDimData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		if This.RepoId != 0 {
			// 2.2 单个规则库维度
			repoDimQueryRes := RepoDimQueryRes{RepoType: 2}
			for _, repoDim := range repoDimList.List {
				repoDimQueryRes.RepoId = This.RepoId
				for _, repoParser := range repoDim.ParserRepoMsg {
					if repoParser.Id == This.RepoId {
						repoDimQueryRes.Dimensions = append(repoDimQueryRes.Dimensions,
							DimensionRes{Id: repoDim.DimensionId, Name: repoDim.Name, Desc: repoDim.Description},
						)
					}
				}
				continue
			}
			if len(repoDimQueryRes.Dimensions) > 0 {
				res = append(res, repoDimQueryRes)
			}
		} else {
			// 2.3 全部规则库维度
			for _, repoDim := range repoDimList.List {
				// 2.3.1 维度对应多个解析规则库
				for n, repoParser := range repoDim.ParserRepoMsg {
					if n == 0 {
						res = append(res, RepoDimQueryRes{RepoId: repoParser.Id, RepoType: 2,
							Dimensions: []DimensionRes{{Id: repoDim.DimensionId, Name: repoDim.Name, Desc: repoDim.Description}},
						})
						continue
					}
					temp := false
					for i, repoRes := range res {
						if repoRes.RepoType == 2 && repoRes.RepoId == repoParser.Id {
							res[i].Dimensions = append(res[i].Dimensions,
								DimensionRes{Id: repoDim.DimensionId, Name: repoDim.Name, Desc: repoDim.Description},
							)
							temp = true
						}
					}
					if !temp {
						res = append(res, RepoDimQueryRes{RepoId: repoParser.Id, RepoType: 2,
							Dimensions: []DimensionRes{{Id: repoDim.DimensionId, Name: repoDim.Name, Desc: repoDim.Description}},
						})
					}
				}
			}
		}
	}
	// 3.关联规则库维度查询
	if This.RepoType == 3 || This.RepoType == 0 {
		// 3.1 查询规则库维度
		repoDimList, code, err := This.GetMappingData(token)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		for _, repoDim := range repoDimList.List {
			repoDimQueryRes := RepoDimQueryRes{RepoType: 3}
			// 3.2 单个规则库维度
			if This.RepoId != 0 {
				if repoDim.Id == This.RepoId {
					repoDimQueryRes.RepoId = This.RepoId
					for _, repoDimField := range repoDim.Fields {
						repoDimQueryRes.Dimensions = append(repoDimQueryRes.Dimensions,
							DimensionRes{Id: repoDimField.Id, Name: repoDim.Name},
						)
					}
					res = append(res, repoDimQueryRes)
					break
				}
			} else {
				// 3.2 全部规则库维度
				repoDimQueryRes.RepoId = repoDim.Id
				for _, repoDimField := range repoDim.Fields {
					repoDimQueryRes.Dimensions = append(repoDimQueryRes.Dimensions, DimensionRes{Id: repoDimField.Id, Name: repoDim.Name})
				}
				res = append(res, repoDimQueryRes)
			}
		}
	}
	return ico.Succ(res)
}

// 识别维度
type RepoDimRes struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    RepoDimData `json:"data"`
}

type RepoDimData struct {
	List []DimInfo `json:"list"`
}

type DimInfo struct {
	RepoId      int    `json:"repo_id"`
	DimensionId int    `json:"dimension_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (This RepoDimQuery) GetRepoDimData(token string) (resData RepoDimData, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"type": 1}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/dimension/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := RepoDimRes{}
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

// 解析维度
type ParserDimRes struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    ParserDimData `json:"data"`
}

type ParserDimData struct {
	List []ParserDimInfo `json:"list"`
}

type ParserDimInfo struct {
	ParserRepoMsg []ParserRepoMsg `json:"parser_repo_msg"`
	DimensionId   int             `json:"dimension_id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
}

type ParserRepoMsg struct {
	Id int `json:"id"`
}

func (This RepoDimQuery) GetParserDimData(token string) (resData ParserDimData, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"type": 1}
	urlRepo := fmt.Sprintf("http://%s/v1/parser-repo/dimension/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return resData, 2301, errors.New("数据查询失败")
	}
	res := ParserDimRes{}
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

func (This RepoDimQuery) GetMappingData(token string) (resData MappingQueryRes, code int, err error) {
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
