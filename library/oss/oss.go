package oss

import (
	"collection-center/internal/ossutil"
	redisConstant "collection-center/library/constant"
	"collection-center/library/redis"
	"github.com/pkg/errors"
	"time"
)

var RoleSessionName, RoleArn string

var config OSSConfig

type OSSConfig struct {
	RegionId        string `json:"regionId"`
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	Endpoint        string `json:"endpoint"`
	CommonBucket    string `json:"commonBucket"` //通用bucket name
}

func SetOSSConfig(regionId, accessKeyId, accessKeySecret, _endpoint, commonBucket string) error {
	config = OSSConfig{
		RegionId:        regionId,
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
		Endpoint:        _endpoint,
		CommonBucket:    commonBucket,
	}
	err := ossutil.InitStsClient(regionId, accessKeyId, accessKeySecret, _endpoint, new(RedisCache))
	return err
}

func GetOSSConfig() OSSConfig {
	return config
}

type RedisCache struct {
}

func (t *RedisCache) Get() (json string, err error) {
	json, _ = redis.Client().Get(redisConstant.STSConfigKey).Result()
	return
}
func (t *RedisCache) Set(json string) error {
	_, err := redis.Client().Set(redisConstant.STSConfigKey, json, time.Second*3600).Result()
	return err
}

func CreateUploadUrl(fileName, filePath string, expireSeconds int64) (string, string, string, error) {
	if RoleArn == "" || RoleSessionName == "" {
		return "", "", "", errors.New("oss配置异常")
	}
	err, _config := ossutil.GetTempInfo(3600, RoleArn, RoleSessionName)
	if err != nil {
		return "", "", "", err
	}
	return ossutil.CreatePutUrlBySts(config.CommonBucket, fileName, filePath, expireSeconds, *_config)
}

func CreateUploadSts(fileName, filePath string, expireSeconds int) (error, *ossutil.StsStruct, string) {
	if RoleArn == "" || RoleSessionName == "" {
		return errors.New("oss配置异常"), nil, ""
	}
	return ossutil.CreateSTSTokenForPost(config.CommonBucket, fileName, filePath, expireSeconds, RoleArn, RoleSessionName)
}
