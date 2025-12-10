package main

import (
	"log"
	"os"

	"lowercode-go-server/bootstrap"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ 未找到 .env 文件，使用系统环境变量")
	}

	// 获取数据库连接字符串
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ DATABASE_URL 环境变量未设置")
	}

	// 测试数据库连接
	db := bootstrap.NewDatabase(dsn)

	// 验证连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ 获取数据库实例失败: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ 数据库 Ping 失败: %v", err)
	}

	log.Println("🎉 数据库连接验证成功！")
}
