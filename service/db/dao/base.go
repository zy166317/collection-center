package dao

import "github.com/go-xorm/xorm"

type Page struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
}

func (p *Page) GetOffset() int {
	return (p.PageNumber - 1) * p.PageSize
}

func GetPageData(xormSession *xorm.Session, list interface{}, pageSize int, offset int, conditionBean interface{}) (int64, error) {
	total, err := xormSession.Limit(pageSize, offset).FindAndCount(list, conditionBean)
	return total, err
}

//注意 使用了selectString的语句，deleted_at的自动填充式失效的
func GetPageDataBySelect(xormSession *xorm.Session, list interface{}, pageSize int, offset int, conditionBean interface{}, selectString string) (int64, error) {
	total, err := xormSession.Select(selectString).Limit(pageSize, offset).FindAndCount(list, conditionBean)
	return total, err
}
