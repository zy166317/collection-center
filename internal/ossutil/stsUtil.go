package ossutil

import (
	"collection-center/internal/constant"
	"collection-center/internal/function"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"mime"
	"path/filepath"
	"sync"
)

type StsStruct struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	SecurityToken   string `json:"securityToken"`
}

type STSPolicyStatement struct {
	Action   []string `json:"Action"`
	Resource []string `json:"Resource"`
	Effect   string   `json:"Effect"`
}
type STSPolicy struct {
	Statement []STSPolicyStatement `json:"Statement"`
	Version   string               `json:"Version"`
}

var stsClient *sts.Client
var endpoint string
var cacheFunc function.CacheFunction
var sr sync.Mutex

/**
 * @Author shenfan
 * @Description // stsclient 初始化
 * @Date 2022/1/7
 * @return
 **/
func InitStsClient(regionId, accessKeyId, accessKeySecret, _endpoint string, _cacheFun function.CacheFunction) error {
	sr.Lock()
	defer sr.Unlock()
	if regionId == "" {
		regionId = string(constant.SH)
	}
	endpoint = _endpoint
	cacheFunc = _cacheFun
	stsCli, err := sts.NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret)
	if err != nil {
		return err
	}
	stsClient = stsCli
	return nil
}

/**
 * @Author shenfan
 * @Description //获取临时的sso 授权信息 过期时间 默认1小时
 * @Date 2022/1/7
 * @return
 **/
func GetTempInfo(durationSeconds int, roleArn, roleSessionName string) (err error, config *StsStruct) {
	if stsClient == nil {
		return errors.New("stsClient not init"), nil
	}
	if cacheFunc != nil {
		//定义了缓存方法，则从缓存获取
		json, err1 := cacheFunc.Get()
		if err1 != nil {
			err = err1
			return
		}
		if json != "" {
			//有缓存，返回缓存信息

		}
	}
	//构建请求对象。
	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"

	//设置参数。关于参数含义和设置方法，请参见《API参考》。
	//request.RoleArn = string(constant.RoleArnLocal)
	//request.RoleSessionName = string(constant.RoleSessionNameLocal)
	request.RoleArn = roleArn
	request.RoleSessionName = roleSessionName
	request.DurationSeconds = requests.NewInteger(durationSeconds)
	//发起请求，并得到响应。
	response, err1 := stsClient.AssumeRole(request)
	if err1 != nil {
		fmt.Print(err1.Error())
		return err1, nil
	}
	config = &StsStruct{
		AccessKeyId:     response.Credentials.AccessKeyId,
		AccessKeySecret: response.Credentials.AccessKeySecret,
		SecurityToken:   response.Credentials.SecurityToken,
	}
	if cacheFunc != nil {
		//缓存到缓存中
		json, err1 := json.Marshal(config)
		if err1 != nil {
			fmt.Println(err1)
		} else {
			err1 = cacheFunc.Set(string(json))
			if err1 != nil {
				fmt.Println(err1)
			}
		}
	}
	return
}

func CreateGetUrlBySts(bucketName, fileName, filePath string, expireSeconds int64, config StsStruct) (string, error) {
	//filePath 举例"upload/video/"
	id := config.AccessKeyId
	secret := config.AccessKeySecret
	token := config.SecurityToken
	ossClient, err := oss.New(endpoint,
		id, secret, oss.SecurityToken(token))
	if err != nil {
		return "", err
	}
	// 获取存储空间
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		return "", err
	}
	key := filePath + fileName
	signedGetURL, err := bucket.SignURL(key, oss.HTTPGet, expireSeconds)
	if err != nil {
		return "", err
	}
	return signedGetURL, err
}

func CreatePutUrlBySts(bucketName, fileName, filePath string, expireSeconds int64, config StsStruct) (string, string, string, error) {
	//filePath 举例"upload/video/"
	id := config.AccessKeyId
	secret := config.AccessKeySecret
	token := config.SecurityToken
	ossClient, err := oss.New(endpoint,
		id, secret, oss.SecurityToken(token))
	if err != nil {
		return "", "", "", err
	}
	// 获取存储空间
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		return "", "", "", err
	}
	// 获取扩展名
	ext := filepath.Ext(fileName)
	mimeType := mime.TypeByExtension(ext)
	// 带可选参数的签名直传
	options := []oss.Option{
		oss.ContentType(mimeType),
	}
	key := filePath + fileName
	// 生成签名url, 签名直传
	signedPutURL, err := bucket.SignURL(key, oss.HTTPPut, expireSeconds, options...)
	if err != nil {
		return "", "", "", err
	}
	return signedPutURL, mimeType, key, err
}

func CreateSTSTokenForPost(bucketName, fileName, filePath string, expireSeconds int, roleArn, roleSessionName string) (err error, config *StsStruct, key string) {
	//filePath 举例"upload/video/"
	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"
	request.RoleArn = roleArn
	request.RoleSessionName = roleSessionName
	request.DurationSeconds = requests.NewInteger(expireSeconds)
	key = filePath + fileName
	policy := STSPolicy{
		Statement: []STSPolicyStatement{{
			Action: []string{"oss:PutObject"},
			//Resource: []string{"acs:oss:*:*:" + bucketName + "/" + key},
			Resource: []string{"acs:oss:*:*:*"},
			Effect:   "Allow",
		},
		},
		Version: "1",
	}
	policyBt, err := json.Marshal(policy)
	if err != nil {
		return
	}
	request.Policy = string(policyBt)
	//发起请求，并得到响应。
	response, err1 := stsClient.AssumeRole(request)
	if err1 != nil {
		fmt.Print(err1.Error())
		return err1, nil, ""
	}
	config = &StsStruct{
		AccessKeyId:     response.Credentials.AccessKeyId,
		AccessKeySecret: response.Credentials.AccessKeySecret,
		SecurityToken:   response.Credentials.SecurityToken,
	}
	return
}
