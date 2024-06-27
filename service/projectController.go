package service

import (
	"collection-center/internal/ecode"
	"collection-center/library/request"
	"collection-center/library/utils"
	"github.com/gin-gonic/gin"
)

type ProjectController struct {
	utils.Controller
}

func NewProjectController(ctx *gin.Context) *ProjectController {
	c := &ProjectController{}
	c.SetContext(ctx)
	return c
}

func (p *ProjectController) CreateProject() {
	req := &request.CreateProjectReq{}
	err := p.Ctx.ShouldBind(req)
	if err != nil {
		p.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	//校验收款信息
	err = CheckCollectInfo(req)
	if err != nil {
		p.ResponseErr(err)
		return
	}
	value, _ := p.Ctx.Get("uid")
	_, err = CreateProject(req, value.(int64))
	if err != nil {
		p.ResponseErr(err)
		return
	}
	p.ResponseOk(map[string]interface{}{})
}

func (p *ProjectController) AddTokenInfo() {
	req := &request.AddTokenInfoReq{}
	err := p.Ctx.ShouldBind(req)
	if err != nil {
		p.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	err = AddTokenInfo(req)
	if err != nil {
		p.ResponseErr(err)
		return
	}
	p.ResponseOk(map[string]interface{}{})
}

func (p *ProjectController) UpdateProjectInfo() {
	req := &request.UpdateProjectInfo{}
	err := p.Ctx.ShouldBind(req)
	if err != nil {
		p.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	value, _ := p.Ctx.Get("uid")
	err = UpdateProjectInfo(req, value.(int64))
	if err != nil {
		p.ResponseErr(err)
		return
	}
	p.ResponseOk(map[string]interface{}{})
}

func (p *ProjectController) UpdateCollectRate() {
	req := &request.UpdateCollectRate{}
	err := p.Ctx.ShouldBind(req)
	if err != nil {
		p.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	value, _ := p.Ctx.Get("uid")
	err = UpdateCollectRate(req, value.(int64))
	if err != nil {
		p.ResponseErr(err)
		return
	}
	p.ResponseOk(map[string]interface{}{})
}

func (p *ProjectController) UpdateCollectAddress() {
	req := &request.UpdateCollectAddress{}
	err := p.Ctx.ShouldBind(req)
	if err != nil {
		p.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	value, _ := p.Ctx.Get("uid")
	err = UpdateCollectAddress(req, value.(int64))
	if err != nil {
		p.ResponseErr(err)
		return
	}
	p.ResponseOk(map[string]interface{}{})
}

func (p *ProjectController) FreezeProjectReq() {
	req := &request.FreezeProjectReq{}
	err := p.Ctx.ShouldBind(req)
	if err != nil {
		p.ResponseWrapErr(ecode.IllegalParam, err)
		return
	}
	value, _ := p.Ctx.Get("uid")
	err = FreezeProject(req, value.(int64))
	if err != nil {
		p.ResponseErr(err)
		return
	}
	p.ResponseOk(map[string]interface{}{})
}
