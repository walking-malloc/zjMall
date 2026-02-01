package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	host := "127.0.0.1"
	port := 3306
	username := "root"
	
	// 尝试几个常见的密码
	passwords := []string{
		"root123456",
		"root",
		"123456",
		"",
		"root@123",
		"Root@123",
	}

	for _, password := range passwords {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql?charset=utf8mb4&parseTime=true", username, password, host, port)
		
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("尝试密码 '%s' 时打开连接失败: %v", maskPassword(password), err)
			continue
		}
		defer db.Close()

		// 设置连接超时
		db.SetMaxOpenConns(1)
		
		err = db.Ping()
		if err == nil {
			fmt.Printf("✅ 连接成功！密码是: %s\n", maskPassword(password))
			fmt.Printf("DSN: %s:%s@tcp(%s:%d)/\n", username, maskPassword(password), host, port)
			os.Exit(0)
		} else {
			log.Printf("❌ 密码 '%s' 不正确: %v", maskPassword(password), err)
		}
	}

	fmt.Println("❌ 所有常见密码都尝试失败，请手动检查 MySQL 密码")
	fmt.Println("\n💡 解决方案：")
	fmt.Println("1. 如果你知道正确的密码，请修改 configs/config.yaml 中的 password")
	fmt.Println("2. 如果是 Docker 容器，运行: docker-compose restart mysql")
	fmt.Println("3. 如果是本地 MySQL，需要重置密码或使用正确的密码")
}

func maskPassword(pwd string) string {
	if pwd == "" {
		return "(空密码)"
	}
	if len(pwd) <= 2 {
		return "***"
	}
	return pwd[:1] + "***" + pwd[len(pwd)-1:]
}




