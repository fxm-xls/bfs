package middleware

import (
	"bigrule/common/logger"
	"bigrule/pkg/utils"
	"bigrule/services/flowcsr-bfs-service/config"
	"encoding/json"
	"errors"
	"fmt"
)

type Permission struct {
	RepoIds    []int
	TagValIds  []int
	ParserIds  []int
	MappingIds []int
	Token      string
}

func GetPermission(token string) (permissionToken Permission, code int, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("permission read error, please ensure the type of data")
		}
	}()
	// 1.获取规则库权限
	permissionRes, code, err := getPermissionData(token)
	if err != nil {
		return
	}
	repoIds := []int{}
	for _, permission := range permissionRes {
		if permission.MenuType == "1" && permission.MenuStatus == "3" {
			repoIds = append(repoIds, permission.MenuId)
		}
	}
	permissionToken.RepoIds = repoIds
	// 2.获取新token
	//newToken, code, err := getUser()
	//if err != nil {
	//	return
	//}
	newToken := token
	permissionToken.Token = newToken
	// 3.获取标签表权限
	dimRes, code, err := getDimData(newToken, repoIds)
	if err != nil {
		return
	}
	for _, dim := range dimRes.List {
		if utils.IsContainsInt(permissionToken.TagValIds, dim.TagValMsg.Id) {
			continue
		}
		permissionToken.TagValIds = append(permissionToken.TagValIds, dim.TagValMsg.Id)
	}
	// 4.获取解析规则库权限
	repoRes, code, err := getRepoData(newToken, repoIds)
	if err != nil {
		return
	}
	for _, repo := range repoRes.List {
		if utils.IsContainsInt(permissionToken.ParserIds, repo.ParserMsg.Id) {
			continue
		}
		permissionToken.ParserIds = append(permissionToken.ParserIds, repo.ParserMsg.Id)
	}
	return
}

// 权限数据
type CheckRes struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    []PermissionRes `json:"data"`
}

type PermissionRes struct {
	MenuId     int    `json:"data_id"`
	MenuType   string `json:"data_type"`
	MenuStatus string `json:"operation"`
}

func getPermissionData(token string) (permissionRes []PermissionRes, code int, err error) {
	pars := map[string]interface{}{"service_name": "flowcsr-service"}
	url := fmt.Sprintf("http://%s/v2/permission/data/get", GetAddr("authentication-service"))
	headers := map[string]string{"X-Access-Token": token}
	resp, err := PostUrl(pars, url, headers)
	if err != nil {
		logger.Error("请求异常")
		return
	}
	checkRes := CheckRes{}
	if err = json.Unmarshal(resp, &checkRes); err != nil {
		logger.Error(string(resp), " json解析异常")
		return
	}
	if checkRes.Code != 200 {
		logger.Error(checkRes.Message)
		return permissionRes, checkRes.Code, errors.New(checkRes.Message)
	}
	permissionRes = checkRes.Data
	return
}

// 用户token
type UserTokenRes struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    UserRes `json:"data"`
}

type UserRes struct {
	UserId int    `json:"user_id"`
	Token  string `json:"token"`
}

func getUser() (token string, code int, err error) {
	pars := map[string]interface{}{"account": config.UserConfig.Name, "password": config.UserConfig.Pwd}
	url := fmt.Sprintf("http://%s/v2/users/login", GetAddr("authentication-service"))
	headers := map[string]string{"X-Access-Token": token}
	resp, err := PostUrl(pars, url, headers)
	if err != nil {
		logger.Error("请求异常")
		return
	}
	userTokenRes := UserTokenRes{}
	if err = json.Unmarshal(resp, &userTokenRes); err != nil {
		logger.Error(string(resp), " json解析异常")
		return
	}
	if userTokenRes.Code != 200 {
		logger.Error(userTokenRes.Message)
		return token, userTokenRes.Code, errors.New(userTokenRes.Message)
	}
	token = userTokenRes.Data.Token
	return
}

// 识别维度
type DimRes struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    DimQueryRes `json:"data"`
}

type DimQueryRes struct {
	List []DimInfo `json:"list"`
}

type DimInfo struct {
	RepoId       int        `json:"repo_id"`
	DictionaryId int        `json:"dictionary_id"`
	TagValMsg    TagValInfo `json:"tagval_msg"`
}

type TagValInfo struct {
	Id int `json:"id"`
}

func getDimData(token string, repo_ids []int) (resData DimQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_ids": repo_ids, "type": 1}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/dimension/query", GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := PostUrl(pars, urlRepo, headers)
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
	ParserMsg ParserRepoInfo `json:"parser_msg"`
}

type ParserRepoInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func getRepoData(token string, repo_ids []int) (resData RepoQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_ids": repo_ids, "type": 1}
	urlRepo := fmt.Sprintf("http://%s/v1/rule-repo/query", GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := PostUrl(pars, urlRepo, headers)
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
