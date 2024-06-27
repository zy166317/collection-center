package middleware

import (
	"collection-center/internal/ecode"
	"collection-center/library/response"
	"collection-center/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ForbiddenRule struct {
	ForbiddenPaths []string
	WhiteList      []int64
	Message        string
	Active         bool
}

func checkAuth(ctx *gin.Context, checkRoles *[]string) error {
	//uid, role, organizationCode, platFormType, username, nickname, err := service.AuthByToken(ctx.GetHeader("Token"))
	uid, err := service.AuthByToken(ctx.GetHeader("Token"))
	if err != nil {
		return err
	}
	ctx.Set("uid", uid)
	if uid == 0 {
		return ecode.AccessDenied
	}
	return nil
}

func checkPermission(ctx *gin.Context, permission ...string) error {
	if len(permission) == 0 {
		return nil
	}
	role, exists := ctx.Get("role")
	if !exists {
		return ecode.AccessDenied
	}
	has := service.CheckPermission(role.(string), permission...)
	if !has {
		return ecode.AccessDenied
	}
	return nil
}

func CheckAuth(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		check := checkAuth(c, &roles)
		if check != nil {
			//logger.Info("checkAuth failure for URI: ", c.Request.RequestURI)
			// c.Json will return two JSON Objs to caller(the request will not be stopped at here.),
			// change it to c.AbortWithStatusJSON, which will certainly end the request.
			language := c.GetHeader("Language")
			c.AbortWithStatusJSON(http.StatusOK, response.ResErr(check, language))
			return
		}
		c.Next()
	}
}

func CheckPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		check := checkPermission(c, permissions...)
		if check != nil {
			language := c.GetHeader("Language")
			c.AbortWithStatusJSON(http.StatusOK, response.ResErr(check, language))
			return
		}
		c.Next()
	}
}
