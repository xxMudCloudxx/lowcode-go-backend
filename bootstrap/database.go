package bootstrap

import (
	"log"
	"time"

	"lowercode-go-server/domain/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase 创建并配置 PostgreSQL 数据库连接
// dsn 格式: postgres://user:password@localhost:5432/dbname?sslmode=disable
func NewDatabase(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 开发时显示 SQL
	})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// ========== 连接池配置 ==========
	// 获取底层的 *sql.DB 对象
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ 获取数据库实例失败: %v", err)
	}

	// SetMaxIdleConns: 空闲连接池中的最大连接数
	// 保持少量空闲连接，避免每次请求都重新建立连接
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns: 数据库的最大打开连接数
	// 防止高并发时耗尽数据库连接（PostgreSQL 默认 max_connections=100）
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime: 连接可复用的最大时间
	// 定期关闭长时间的连接，防止连接过久导致的问题
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移表结构
	if err := db.AutoMigrate(&entity.Page{}, &entity.User{}); err != nil {
		log.Fatalf("❌ 数据库迁移失败: %v", err)
	}

	log.Println("✅ PostgreSQL 连接成功，连接池已配置，表结构已同步")
	return db
}
