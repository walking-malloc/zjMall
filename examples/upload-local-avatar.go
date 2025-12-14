package main

import (
	"fmt"
	"log"
	"path/filepath"

	upload "zjMall/internal/common/oss"
	"zjMall/internal/config"
)

// 示例：如何上传本地图片文件到 OSS
func main() {
	// 1. 加载配置
	configPath := filepath.Join("./configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 创建 OSS 客户端
	ossConfig := cfg.GetOSSConfig()
	ossClient, err := upload.NewOSSClient(ossConfig)
	if err != nil {
		log.Fatalf("创建 OSS 客户端失败: %v", err)
	}

	// 3. 上传本地文件
	userID := "01ARZ3NDEKTSV4RRFFQ69G5FAV" // 用户ID
	localFilePath := "C:\\Users\\Administrator\\Pictures\\avatar.jpg" // 本地文件路径（Windows）
	// localFilePath := "/home/user/avatar.jpg" // Linux/Mac 路径

	avatarURL, err := ossClient.UploadAvatarFromFile(userID, localFilePath)
	if err != nil {
		log.Fatalf("上传失败: %v", err)
	}

	fmt.Printf("✅ 上传成功！\n")
	fmt.Printf("头像URL: %s\n", avatarURL)
}

