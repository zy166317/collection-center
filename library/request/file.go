package request

type CreateOSSUploadReq struct {
	FileName string `json:"FileName" binding:"required"` // 动作（n选1）
}
