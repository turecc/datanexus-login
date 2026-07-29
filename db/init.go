package db

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"wxcloudrun-golang/db/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	dbMu       sync.RWMutex
	dbInstance *gorm.DB
	initMu     sync.Mutex
)

// Init 初始化数据库。
// 该方法允许安全重试；只有连接、Ping 和迁移全部成功后才会发布数据库实例。
func Init() error {
	initMu.Lock()
	defer initMu.Unlock()

	if IsReady() {
		return nil
	}

	sourceTemplate := "%s:%s@tcp(%s)/%s?timeout=2s&readTimeout=2s&writeTimeout=2s&charset=utf8mb4&loc=Local&parseTime=true"
	user := os.Getenv("MYSQL_USERNAME")
	pwd := os.Getenv("MYSQL_PASSWORD")
	addr := os.Getenv("MYSQL_ADDRESS")
	dataBase := os.Getenv("MYSQL_DATABASE")
	if dataBase == "" {
		dataBase = "golang_demo"
	}

	if user == "" || addr == "" {
		return fmt.Errorf("mysql environment variables are incomplete")
	}

	startedAt := time.Now()
	source := fmt.Sprintf(sourceTemplate, user, pwd, addr, dataBase)
	fmt.Printf("start init mysql, address=%s, database=%s, user=%s\n", addr, dataBase, user)

	database, err := gorm.Open(mysql.Open(source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return fmt.Errorf("open mysql failed after %s: %w", time.Since(startedAt), err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("initialize mysql failed after %s: %w", time.Since(startedAt), err)
	}

	// 确保失败连接不会长期占用资源。
	initSucceeded := false
	defer func() {
		if !initSucceeded {
			_ = sqlDB.Close()
		}
	}()

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	err = sqlDB.PingContext(pingCtx)
	pingCancel()
	if err != nil {
		return fmt.Errorf("ping mysql failed after %s: %w", time.Since(startedAt), err)
	}

	// 迁移在后台初始化流程中执行，不再阻塞 HTTP 端口监听。
	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = database.WithContext(migrateCtx).AutoMigrate(&model.DataNexusUserModel{})
	migrateCancel()
	if err != nil {
		return fmt.Errorf("migrate datanexus users table failed after %s: %w", time.Since(startedAt), err)
	}

	dbMu.Lock()
	dbInstance = database
	dbMu.Unlock()
	initSucceeded = true

	fmt.Printf("finish init mysql, cost=%s\n", time.Since(startedAt))
	return nil
}

// Get 返回数据库实例。数据库尚未就绪时返回 nil。
func Get() *gorm.DB {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return dbInstance
}

// IsReady 返回数据库是否已完成初始化。
func IsReady() bool {
	return Get() != nil
}
