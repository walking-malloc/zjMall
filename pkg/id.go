package pkg

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// GenerateULID 生成 ULID（用户 ID 等）
// ULID 是 UUID 的替代方案，特点是：
// 1. 可排序（基于时间戳）
// 2. 唯一性（基于随机数）
// 3. 紧凑（26 个字符，比 UUID 的 36 个字符更短）
// 4. URL 安全
func GenerateULID() string {
	entropy := rand.Reader
	ms := ulid.Timestamp(time.Now())
	id, err := ulid.New(ms, entropy)
	if err != nil {
		// 如果生成失败，使用时间戳作为后备方案（不应该发生）
		return ulid.Make().String()
	}
	return id.String()
}
