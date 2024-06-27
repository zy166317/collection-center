package ossutil

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"mime"
	"path/filepath"
	"sync"
)

var ossClient *oss.Client

var r sync.Mutex

/**
* @Author shenfan
* @Description // 创建OSSClient实例 项目启动创建
* @Date 2022/1/5
* @return
**/
func OSSClientInit(endpoint string, accessKeyId string, accessKeySecret string) error {
	r.Lock()
	defer r.Unlock()
	if ossClient == nil {
		c, err := oss.New(endpoint, accessKeyId, accessKeySecret)
		if err != nil {
			return err
		}
		ossClient = c
		fmt.Println("oss client created")
	}
	return nil
}

/**
* @Author shenfan
* @Description //文件上传 传文件数组
* @Date 2022/1/5
* @return
**/
func UploadFileBytes(bucketName string, fileBytes []byte, objectName string) error {
	if ossClient == nil {
		fmt.Println("ossClient not init")
		return errors.New("ossClient not init")
	}

	// 获取存储空间。
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error1:", err)
		return err
	}

	// 上传Byte数组。
	err = bucket.PutObject(objectName, bytes.NewReader(fileBytes))
	if err != nil {
		fmt.Println("Error2:", err)
		return err
	}
	return nil
}

/**
* @Author shenfan
* @Description //文件下载  返回[]byte
* @Date 2022/1/5
* @return
**/
func DownloadFile(bucketName string, objectName string) ([]byte, error) {
	if ossClient == nil {
		fmt.Println("ossClient not init")
		return nil, errors.New("ossClient not init")
	}
	// 获取存储空间。
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error3:", err)
		return nil, err
	}

	// 下载文件到流。
	body, err := bucket.GetObject(objectName)
	if err != nil {
		fmt.Println("Error4:", err)
		return nil, err
	}
	// 数据读取完成后，获取的流必须关闭，否则会造成连接泄漏，导致请求无连接可用，程序无法正常工作。
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println("Error5:", err)
		return nil, err
	}
	return data, nil
}

/**
* @Author shenfan
* @Description // 图片删除
* @Date 2022/1/5
* @return
**/
func DeleteFile(bucketName string, objectName string) error {
	if ossClient == nil {
		fmt.Println("ossClient not init")
		return errors.New("ossClient not init")
	}
	// 获取存储空间。
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error6:", err)
		return err
	}

	// 删除单个文件。objectName表示删除OSS文件时需要指定包含文件后缀在内的完整路径，例如abc/efg/123.jpg。
	// 如需删除文件夹，请将objectName设置为对应的文件夹名称。如果文件夹非空，则需要将文件夹下的所有object删除后才能删除该文件夹。
	err = bucket.DeleteObject(objectName)
	if err != nil {
		fmt.Println("Error7:", err)
		return err
	}
	return nil
}

func CreateGetUrl(bucketName, fileName, filePath string, expireSeconds int64) (string, error) {
	//filePath 举例"upload/video/"
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

func CreatePutUrl(bucketName, fileName, filePath string, expireSeconds int64) (string, string, error) {
	//filePath 举例"upload/video/"
	// 获取存储空间
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		return "", "", err
	}
	// 获取扩展名
	ext := filepath.Ext(fileName)
	// 带可选参数的签名直传
	options := []oss.Option{
		oss.ContentType(mime.TypeByExtension(ext)),
	}
	key := filePath + fileName
	// 生成签名url, 签名直传
	signedPutURL, err := bucket.SignURL(key, oss.HTTPPut, expireSeconds, options...)
	if err != nil {
		return "", "", err
	}
	return signedPutURL, key, err
}
