package middleware

import (
	"bytes"
	"collection-center/internal/ecode"
	"collection-center/internal/logger"
	"collection-center/library/response"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// 对请求的追踪内容进行自定义处理
type RespTraceHandler func(responseBody string) (code ecode.Code, msg string, data interface{}, err error)
type ReqTraceHandler func(req *http.Request) (reqStr string, err error)

// 需要忽略的键值名称
func TraceLog(reqHandler ReqTraceHandler, respHandler RespTraceHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyLogWriter := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyLogWriter
		//处理请求
		c.Next()
		responseBody := bodyLogWriter.body.String()
		var responseCode ecode.Code
		var responseMsg string
		var reqStr string
		var responseData interface{}

		if responseBody != "" {
			if respHandler != nil {
				code, msg, dataStr, err := respHandler(responseBody)
				if err != nil {
					logger.Error(err)
					return
				}
				responseCode = code
				responseMsg = msg
				responseData = dataStr
			} else {
				response := response.Result{}
				err := json.Unmarshal([]byte(responseBody), &response)
				if err == nil {
					responseCode = response.Code
					responseMsg = response.Message
					responseData = response.Data
				}
			}
		}
		if reqHandler != nil {
			req, err := reqHandler(c.Request)
			if err != nil {
				logger.Error(err)
				return
			}
			reqStr = req
		} else {
			if c.Request.Method == "POST" {
				c.Request.ParseForm()
			}
			reqStr = c.Request.PostForm.Encode()
		}
		//日志格式
		accessLogMap := make(map[string]interface{})
		accessLogMap["request_method"] = c.Request.Method
		accessLogMap["request_uri"] = c.Request.RequestURI
		accessLogMap["request_proto"] = c.Request.Proto
		accessLogMap["request_ua"] = c.Request.UserAgent()
		accessLogMap["request_referer"] = c.Request.Referer()
		accessLogMap["request_post_data"] = reqStr
		accessLogMap["request_client_ip"] = c.ClientIP()

		accessLogMap["response_code"] = responseCode
		accessLogMap["response_msg"] = responseMsg
		accessLogMap["response_data"] = responseData
		accessLogJson, _ := json.Marshal(accessLogMap)
		logger.Info(string(accessLogJson))
	}
}
