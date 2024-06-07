package constant

/*
地域id
*/
type RegionId string

const (
	SH RegionId = "cn-shanghai"
)

/*
角色arn
*/
type RoleArn string

const (
	RoleArnLocal RoleArn = "acs:ram::1859000650827546:role/ossbackup"
)

/*
角色名称
*/
type RoleSessionName string

const (
	RoleSessionNameLocal RoleSessionName = "ossbackup"
)

/*
代币类型
*/
