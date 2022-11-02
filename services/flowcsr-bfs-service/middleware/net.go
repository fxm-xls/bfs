package middleware

import (
	"bigrule/common/global"
	"bigrule/common/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"io/ioutil"
	"net/http"
	"time"
)

// HttpClient 会话代理
func HttpClient(method string, url string, params []byte, headparams map[string]string) ([]byte, error) {
	client := http.Client{Timeout: 20 * time.Second}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(params))
	if err != nil {
		return nil, err
	}
	// set head of request
	for k, v := range headparams {
		request.Header.Set(k, v)
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// PostUrl 统一发送请求
func PostUrl(params interface{}, url string, headext ...map[string]string) (resp []byte, err error) {
	headparams := map[string]string{"Content-Type": "application/json"}
	if len(headext) > 0 {
		for _, hm := range headext {
			for k, v := range hm {
				headparams[k] = v
			}
		}
	}
	resqbyte, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	resp, err = HttpClient("POST", url, resqbyte, headparams)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUrl 统一发送请求
func GetUrl(params map[string]interface{}, url string, headext ...map[string]string) (resp []byte, err error) {
	headparams := map[string]string{}
	if len(headext) > 0 {
		for _, hm := range headext {
			for k, v := range hm {
				headparams[k] = v
			}
		}
	}
	url += "?"
	for k, v := range params {
		url += fmt.Sprintf("%s=%s&", k, fmt.Sprint(v))
	}
	resqbyte, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	resp, err = HttpClient("GET", url, resqbyte, headparams)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetAddr 获取ip
func GetAddr(serviceName string) (address string) {
	var retryCount int
	for {
		servers, err := global.EtcdReg.GetService(serviceName)
		if err != nil {
			logger.Error(err.Error())
		}
		var services []*registry.Service
		for _, value := range servers {
			services = append(services, value)
		}
		next := selector.RoundRobin(services)
		if node, err := next(); err == nil {
			address = node.Address
		}
		if len(address) > 0 {
			return
		}
		logger.Errorf("fail times: %d\n", retryCount)
		//重试次数++
		retryCount++
		time.Sleep(time.Second * 1)
		//重试5次为获取返回空
		if retryCount >= 2 {
			return
		}
	}
}
