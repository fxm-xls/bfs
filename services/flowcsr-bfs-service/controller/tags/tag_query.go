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

type TagQuery struct {
	TagValTblId int `json:"tagval_tbl_id"     binding:"required"`
	TagId       int `json:"tag_id"            binding:"required"`
	TagName     string
	RepoIds     []int
}

type TagQueryRes struct {
	RepoId        int       `json:"repo_id"`
	RuleId        int       `json:"rule_id"`
	RepoName      string    `json:"repo_name"`
	Attributes    Attribute `json:"attributes"`
	ParserMessage Parser    `json:"parser_message"`
}

type Attribute struct {
	Priority int    `json:"priority"`
	Status   int    `json:"status"`
	Sample   string `json:"sample"`
	Desc     string `json:"desc"`
}

type Parser struct {
	RepoParserId   int    `json:"repo_parser_id"`
	RepoParserName string `json:"repo_parser_name"`
	ParserId       int    `json:"parser_id"`
	Description    string `json:"description"`
}

func (This TagQuery) DoHandle(c *gin.Context) *ico.Result {
	if err := c.ShouldBindJSON(&This); err != nil {
		return ico.Err(2099, "", err.Error())
	}
	token := c.GetHeader("X-Access-Token")
	token = strings.Split(token, ";")[0]
	logger.Info("标签查询")
	// 1.通过标签表获取维度和规则库
	code, err := This.GetTagName(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	// 2.通过标签表获取维度和规则库
	dimDataList, code, err := This.GetDimData(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	// 3.通过规则库获取解析规则库
	repoDataList, code, err := This.GetRepoData(token)
	if err != nil {
		return ico.Err(code, err.Error())
	}
	// 4.通过规则库、维度、标签获取规则信息
	res := []TagQueryRes{}
	for _, dimData := range dimDataList.List {
		ruleData, code, err := This.GetRuleData(token, dimData.RepoId, dimData.DimensionId)
		if err != nil {
			return ico.Err(code, err.Error())
		}
		// 5.整理
		repoName := ""
		repoParserId := 0
		repoParserName := ""
		for _, repo := range repoDataList.List {
			if repo.RepoId == dimData.RepoId {
				repoName = repo.Name
				repoParserId = repo.ParserMsg.Id
				repoParserName = repo.ParserMsg.Name
				break
			}
		}
		for _, rule := range ruleData.List {
			statusInt, _ := strconv.Atoi(rule.Attributes.Status)
			priorityInt, _ := strconv.Atoi(rule.Attributes.Priority)
			res = append(res, TagQueryRes{
				RepoId: dimData.RepoId, RepoName: repoName, RuleId: rule.RuleId,
				Attributes: Attribute{Desc: rule.Attributes.Desc, Sample: rule.Attributes.Sample, Status: statusInt, Priority: priorityInt},
				ParserMessage: Parser{RepoParserId: repoParserId, RepoParserName: repoParserName,
					ParserId: rule.ParserMessage.ParserId, Description: rule.ParserMessage.Description},
			})
		}
	}
	return ico.Succ(res)
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
	Name    string `json:"name"`
	LineNum int    `json:"line_num"`
}

func (This *TagQuery) GetTagName(token string) (code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"tagval_tbl_id": This.TagValTblId, "type": 1, "tagval_ids": []int{This.TagId}}
	urlRepo := fmt.Sprintf("http://%s/v1/tag/query", middleware.GetAddr("repo-service"))
	headers := map[string]string{"X-Access-Token": token}
	// 2.获取数据
	respRepo, err := middleware.PostUrl(pars, urlRepo, headers)
	if err != nil {
		logger.Info("数据查询失败")
		return 2301, errors.New("数据查询失败")
	}
	res := TagRes{}
	err = json.Unmarshal(respRepo, &res)
	if err != nil {
		logger.Info("数据查询失败", string(respRepo), pars)
		return 2301, errors.New("数据查询失败")
	}
	if res.Code != 200 || len(res.Data.List) == 0 {
		logger.Info("数据查询失败")
		return res.Code, errors.New(res.Message)
	}
	This.TagName = res.Data.List[0].Name
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
	RepoId      int `json:"repo_id"`
	DimensionId int `json:"dimension_id"`
}

func (This *TagQuery) GetDimData(token string) (resData DimQueryRes, code int, err error) {
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
	for _, dim := range res.Data.List {
		This.RepoIds = append(This.RepoIds, dim.RepoId)
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
	RepoId       int             `json:"repo_id"`
	Name         string          `json:"name"`
	ParserMsg    ParserRepoInfo  `json:"parser_msg"`
	DimensionMsg []DimensionInfo `json:"dimension_msg"`
}

type ParserRepoInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type DimensionInfo struct {
	DimensionId int    `json:"dimension_id"`
	Name        string `json:"name"`
}

func (This TagQuery) GetRepoData(token string) (resData RepoQueryRes, code int, err error) {
	// 1.整理入参
	pars := map[string]interface{}{"repo_ids": This.RepoIds, "type": 1}
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
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    RuleQueryRes `json:"data"`
}

type RuleQueryRes struct {
	List []RuleInfo `json:"list"`
}

type RuleInfo struct {
	RuleId        int          `json:"rule_id"`
	LineNum       int          `json:"line_num"`
	ParserMessage ParserInfo   `json:"parser_message"`
	Attributes    AttributeStr `json:"attributes"`
	Dimensions    []Dimension  `json:"dimensions"`
}

type AttributeStr struct {
	Priority string `json:"priority"`
	Status   string `json:"status"`
	Sample   string `json:"sample"`
	Desc     string `json:"desc"`
}

type Dimension struct {
	DimensionId int    `json:"dimension_id"`
	TagId       int    `json:"tag_id"`
	TagName     string `json:"tag_name"`
}

type ParserInfo struct {
	ParserId    int    `json:"parser_id"`
	Description string `json:"description"`
}

func (This TagQuery) GetRuleData(token string, repoId, dimId int) (resData RuleQueryRes, code int, err error) {
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
