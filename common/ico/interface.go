package ico

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type IController interface {
	DoHandle(c *gin.Context) *Result
}

func Handler(controller IController) gin.HandlerFunc {
	return func(c *gin.Context) {

		rst := controller.DoHandle(c)

		switch strings.ToLower(rst.Type) {
		case "json":
			c.JSON(http.StatusOK, rst)

		case "string":
			c.String(http.StatusOK, rst.Message)

		case "file":
		}
	}
}

type Result struct {
	Type    string        `json:"-"`
	Status  int           `json:"status"`
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    interface{}   `json:"data"`
	Logs    []interface{} `json:"-"`
}

func Succ(data interface{}, logs ...interface{}) *Result {
	if data == nil {
		data = map[string]string{}
	}
	return &Result{
		Type:    "json",
		Status:  1,
		Code:    200,
		Message: "ok",
		Data:    data,
		Logs:    logs,
	}
}

func Err(code int, message string, logs ...interface{}) *Result {
	return &Result{
		Type:    "json",
		Status:  0,
		Code:    code,
		Message: message,
		Data:    map[string]string{},
		Logs:    logs,
	}
}

func String(message string) *Result {
	return &Result{
		Type:    "string",
		Status:  1,
		Message: message,
	}
}

func File(message string) *Result {
	return &Result{
		Type:    "file",
		Status:  1,
		Message: message,
	}
}
