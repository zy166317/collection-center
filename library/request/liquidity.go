package request

type OutReq struct {
	Mode                string `json:"mode" form:"mode" binding:"required"`
	Originaltoken       string `json:"originaltoken" form:"originaltoken" binding:"required"`
	Originaltokenamount string `json:"originaltokenamount" form:"originaltokenamount" binding:"required"`
	Targettoken         string `json:"targettoken" form:"targettoken" binding:"required"`
}
