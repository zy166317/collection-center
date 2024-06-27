package utils

import (
	"github.com/bwmarrin/snowflake"
)

func GenerateUid() int64 {
	node, err := snowflake.NewNode(1)
	if err != nil {

	}
	//定义开始时间，毫秒级时间戳
	snowflake.Epoch = 1719365902000
	snowflakeId := node.Generate()
	return snowflakeId.Int64()
}
