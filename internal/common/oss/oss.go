package upload

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"zjMall/internal/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// UploadClient 上传客户端接口
type UploadClient interface {
	UploadAvatar(userID string, file io.Reader, filename string) (string, error)
}

// OSSClient OSS 上传客户端
type OSSClient struct {
	client     *oss.Client
	bucket     *oss.Bucket
	bucketName string
	baseURL    string
	avatarPath string
}

// NewOSSClient 创建 OSS 客户端
func NewOSSClient(cfg *config.OSSConfig) (UploadClient, error) {
	// 1. 创建 OSS 客户端
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建 OSS 客户端失败: %w", err)
	}

	// 2. 获取 Bucket
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("获取 Bucket 失败: %w", err)
	}

	// 3. 检查 Bucket 是否存在
	exist, err := client.IsBucketExist(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("检查 Bucket 是否存在失败: %w", err)
	}
	if !exist {
		return nil, fmt.Errorf("Bucket %s 不存在", cfg.BucketName)
	}

	return &OSSClient{
		client:     client,
		bucket:     bucket,
		bucketName: cfg.BucketName,
		baseURL:    cfg.BaseURL,
		avatarPath: cfg.AvatarPath,
	}, nil
}

// UploadAvatar 上传头像到 OSS
func (o *OSSClient) UploadAvatar(userID string, file io.Reader, filename string) (string, error) {
	// 1. 生成 OSS 对象名（路径）
	// 格式：avatars/2025/01/user_01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg
	now := time.Now()
	yearMonth := now.Format("2006/01")
	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("%s/%s/user_%s%s", o.avatarPath, yearMonth, userID, ext)

	// 2. 上传到 OSS，并设置 Content-Type 和 Content-Disposition
	// 根据文件扩展名设置 Content-Type，这样浏览器才能正确显示图片而不是下载
	contentType := getContentType(ext)
	options := []oss.Option{
		oss.ContentType(contentType),     // 设置 Content-Type，让浏览器直接显示图片
		oss.ContentDisposition("inline"), // 设置为 inline，让浏览器预览而不是下载
	}
	err := o.bucket.PutObject(objectName, file, options...)
	if err != nil {
		return "", fmt.Errorf("上传到 OSS 失败: %w", err)
	}

	// 3. 上传后单独设置 ACL 为公共读（如果阻止公共访问功能关闭，这一步会成功）
	// 如果失败，说明阻止公共访问功能已开启，需要先在控制台关闭
	err = o.bucket.SetObjectACL(objectName, oss.ACLPublicRead)
	if err != nil {
		// 如果设置 ACL 失败，记录日志但不阻止上传（文件已上传成功，只是无法公共访问）
		// 可以通过预签名 URL 访问，或者需要在控制台关闭"阻止公共访问"功能
		return "", fmt.Errorf("设置文件 ACL 为公共读失败: %w (请在 OSS 控制台关闭'阻止公共访问'功能)", err)
	}

	// 3. 生成访问 URL（因为设置了公共读，可以直接访问）
	url := fmt.Sprintf("%s/%s", o.baseURL, objectName)
	return url, nil
}

// getContentType 根据文件扩展名返回对应的 Content-Type
func getContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "image/jpeg" // 默认使用 jpeg
	}
}
