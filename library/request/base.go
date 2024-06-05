package request

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

type Page struct {
	PageNumber int `json:"pageNumber" binding:"required"`
	PageSize   int `json:"pageSize" binding:"required"`
}

func (p *Page) GetOffset() int {
	return (p.PageNumber - 1) * p.PageSize
}

func GetPageQuery(c *gin.Context) (*Page, error) {
	pageNumber, _ := strconv.Atoi(c.DefaultQuery("pageNumber", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	return &Page{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}, nil
}

func GetPagePost(c *gin.Context) (*Page, error) {
	pageNumber, err := strconv.Atoi(c.DefaultPostForm("pageNumber", "1"))
	if err != nil {
		return nil, err
	}
	pageSize, err := strconv.Atoi(c.DefaultPostForm("pageSize", "20"))
	if err != nil {
		return nil, err
	}
	return &Page{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}, nil
}
