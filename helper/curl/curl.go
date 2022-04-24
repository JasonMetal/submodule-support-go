package curl

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	Url         string
	Method      string
	Headers     map[string]string
	BodyData    interface{}
	ReqTimeOut  time.Duration
	RespTimeOut time.Duration
}
type Response struct {
	Status     string
	StatusCode int
	Header     map[string]string
	Body       []byte
}

func Get(req Request) (resp Response, err error) {
	req.Method = "GET"
	return Send(req)
}

func Post(req Request) (resp Response, err error) {
	req.Method = "POST"
	return Send(req)
}

func Send(request Request) (resp Response, err error) {
	var client = initClient(request)
	//请求超时时间设置
	var requestTimeout time.Duration = 5
	if request.ReqTimeOut > 0 {
		requestTimeout = 5
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*requestTimeout)
	defer cancel()
	//设置body
	bodyData := setBodyData(request.BodyData)
	// 设置上下文请求
	req, err := http.NewRequestWithContext(ctx, request.Method, request.Url, bodyData)
	if err != nil {
		return resp, err
	}
	//设置请求头
	for k, v := range request.Headers {
		req.Header.Add(k, v)
	}
	response, err := client.Do(req)
	defer response.Body.Close()
	if err != nil {
		return resp, err
	}
	resp.Status = response.Status
	resp.StatusCode = response.StatusCode
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return resp, err
	}
	resp.Body = body
	header := make(map[string][]string)
	for k, v := range response.Header {
		header[k] = v
	}
	return resp, nil

}

// 格式化url
func HttpBuildQuery(params map[string]string) string {
	var uri url.URL
	q := uri.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	queryStr := q.Encode()
	return queryStr
}

/*func CallPostInfo(url string, data interface{}, headers map[string]string, timeout time.Duration) ([]byte, int, error) {
	urlParse, _ := urlParse.Parse(url)
	if timeout == 0 {
		timeout = 5 * time.Second
		postNotSetTimeoutNum.CounterWithLabelValues(urlParse.Host, urlParse.Path).Inc()
	}

	var req *http.Request
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch v := data.(type) {
	case string:
		req, err = http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(v))
	case []byte:
		req, err = http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(v))
	default:
		s, err := json.Marshal(v)
		if err != nil {
			core.ErrorLogDetail(core.LogStruct{
				Err:       err,
				ErrType:   "httpPost",
				NeedStack: false,
				ErrLevel:  "error",
				ErrInfo:   map[string]interface{}{"url": url, "body": data},
				Metrics:   "",
			})

			return nil, 0, err
		}

		req, err = http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(s))
	}

	if err != nil {
		postMakeRequestFailNum.CounterWithLabelValues(urlParse.Host, urlParse.Path).Inc()
		return nil, 0, err
	}
	defer req.Body.Close()

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	postRequestNum.CounterWithLabelValues(urlParse.Host, urlParse.Path).Inc()
	startTime := time.Now().UnixNano()
	resp, err := client.Do(req)
	execTime := float64(time.Now().UnixNano()-startTime) / float64(time.Millisecond)
	postCallTime.CounterWithLabelValues(urlParse.Host, urlParse.Path).Add(execTime)

	if err != nil {
		postCallFailNum.CounterWithLabelValues(urlParse.Host, urlParse.Path).Inc()
		//log.Println("url:", url, "err:", err, "body:", data)
		core.ErrorLogDetail(core.LogStruct{
			Err:       err,
			ErrType:   "httpPost",
			NeedStack: false,
			ErrLevel:  "error",
			ErrInfo:   map[string]interface{}{"url": url, "body": data},
			Metrics:   "",
		})

		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		postReadBodyFailNum.CounterWithLabelValues(urlParse.Host, urlParse.Path).Inc()
		//log.Println("url:", url, "err:", err, "body:", body)
		core.ErrorLogDetail(core.LogStruct{
			Err:       err,
			ErrType:   "httpPost",
			NeedStack: false,
			ErrLevel:  "error",
			ErrInfo:   map[string]interface{}{"url": url, "body": data},
			Metrics:   "",
		})
	}

	return body, resp.StatusCode, err
}
*/

//设置body 肉体
func setBodyData(bodyData interface{}) io.Reader {
	var data io.Reader
	switch v := bodyData.(type) {
	case string:
		data = strings.NewReader(v)
	case []byte:
		data = bytes.NewBuffer(v)
	case map[string]string:
		data = strings.NewReader(HttpBuildQuery(bodyData.(map[string]string)))
	}
	return data
}

func initClient(request Request) (client http.Client) {
	// 忽略 https 证书校验
	var responseheadertimeout time.Duration = 0
	if request.RespTimeOut > 0 {
		responseheadertimeout = time.Second * request.RespTimeOut
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 400,
		IdleConnTimeout:     30 * time.Second,
		//DisableKeepAlives:   true,
		ResponseHeaderTimeout: responseheadertimeout,
	}
	client = http.Client{Transport: transport}
	return client
}
