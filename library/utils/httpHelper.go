package utils

import (
	"collection-center/library/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct {
	Ctx      *gin.Context
	Language string
}

func (c *Controller) SetContext(ctx *gin.Context) {
	c.Ctx = ctx
	c.Language = ctx.GetHeader("Language")
}

func (c *Controller) Response(err error, data interface{}) {
	if err != nil {
		c.ResponseErr(err)
	} else {
		c.ResponseOk(data)
	}
}

func (c *Controller) ResponseOk(data interface{}) {
	res := response.ResOk(data)
	if res.Data == nil {
		res.Data = map[string]string{}
	}
	c.Ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ResponseErr(err error) {
	res := response.ResErr(err, c.Language)
	if res.Data == nil {
		res.Data = map[string]string{}
	}
	c.Ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ResponseWrapErr(err error, err1 error) {
	res := response.ResWrapErr(err, err1, c.Language)
	if res.Data == nil {
		res.Data = map[string]string{}
	}
	c.Ctx.JSON(http.StatusOK, res)
}
