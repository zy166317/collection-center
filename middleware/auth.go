package middleware

import (
	"collection-center/internal/ecode"
	"collection-center/library/constant"
	"collection-center/library/response"
	"collection-center/library/utils"
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
	uid, role, _, _, username, nickname, err := service.AuthByToken(ctx.GetHeader("Token"))
	if err != nil {
		return err
	}
	ctx.Set("uid", uid)
	if uid == 0 {
		return ecode.AccessDenied
	}
	//只有超级账号和医保局账号可操作此平台
	//if platFormType != constant.PLAT_FORM_BP && platFormType != constant.PLAT_FORM_ALL {
	//	return ecode.AccessDenied
	//}
	if checkRoles != nil {
		if len(*checkRoles) != 0 {
			if (*checkRoles)[0] != "" && utils.FindInString(role, checkRoles) == -1 {
				return ecode.AccessDenied
			}
		}
	}
	ctx.Set(constant.ROLE, role)
	//ctx.Set(constant.ORG_CODE, organizationCode)
	//ctx.Set(constant.PLAT_FORM_TYPE, platFormType)
	ctx.Set(constant.USERNAME, username)
	ctx.Set(constant.NICKNAME, nickname)
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
