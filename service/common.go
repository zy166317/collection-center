package service

import (
	"collection-center/internal/ecode"
	"collection-center/library/constant"
	"collection-center/library/oss"
	"collection-center/library/request"
	"collection-center/library/utils"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

type CommonController struct {
	utils.Controller
}

func NewCommonController(ctx *gin.Context) *CommonController {
	c := &CommonController{}
	c.SetContext(ctx)
	return c
}

func (c *CommonController) CreateOssUploadSTS() {
	var req request.CreateOSSUploadReq
	err := c.Ctx.ShouldBind(&req)
	if err != nil {
		c.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}

	username := c.Ctx.GetString(constant.USERNAME)
	username = utils.GenerateMd5(username)[0:10]
	//根据账号的md5生成目录
	err, sts, key := oss.CreateUploadSts(req.FileName, "common/"+username+"/", 900)
	if err != nil {
		c.ResponseErr(err)
		return
	}
	data := gin.H{
		"host":            strings.Replace(oss.GetOSSConfig().Endpoint, "https://", "https://"+oss.GetOSSConfig().CommonBucket+".", -1),
		"accessKeyId":     sts.AccessKeyId,
		"accessKeySecret": sts.AccessKeySecret,
		"securityToken":   sts.SecurityToken,
		"visitUrl":        key,
	}
	c.Response(err, data)
}
func (c *CommonController) CreateOssUploadUrl() {
	var req request.CreateOSSUploadReq
	err := c.Ctx.ShouldBind(&req)
	if err != nil {
		c.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	username := c.Ctx.GetString(constant.USERNAME)
	username = utils.GenerateMd5(username)[0:10]
	//根据账号的md5生成目录
	url, contentType, key, err := oss.CreateUploadUrl(req.FileName, "common/"+username+"/", 300)
	data := gin.H{
		"contentType": contentType,
		"uploadUrl":   url,
		"visitUrl":    key,
	}
	c.Response(err, data)
}

func (c *CommonController) CreatePlatformOssUploadUrl() {
	var req request.CreateOSSUploadReq
	err := c.Ctx.ShouldBind(&req)
	if err != nil {
		c.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	//日期
	monthYear := utils.GetYearMonthStr(time.Now())
	url, contentType, key, err := oss.CreateUploadUrl(req.FileName, "platform/"+monthYear+"/", 900)
	data := gin.H{
		"contentType": contentType,
		"uploadUrl":   url,
		"visitUrl":    key,
	}
	c.Response(err, data)
}
