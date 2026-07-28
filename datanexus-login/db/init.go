package db

import (
	"fmt"
	"os"
	"time"

	"wxcloudrun-golang/db/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var dbInstance *gorm.DB

// Init 初始化数据库。
func Init() error {
	sourceTemplate := "%s:%s@tcp(%s)/%s?readTimeout=1500ms&writeTimeout=1500ms&charset=utf8mb4&loc=Local&parseTime=true"
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

	source := fmt.Sprintf(sourceTemplate, user, pwd, addr, dataBase)
	fmt.Printf("start init mysql, address=%s, database=%s, user=%s\n", addr, dataBase, user)

	database, err := gorm.Open(mysql.Open(source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return fmt.Errorf("open mysql failed: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("initialize mysql failed: %w", err)
	}

	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	dbInstance = database

	// 自动确保 DataNexus 登录表存在，后续 Git 推送发布无需手工执行 SQL。
	if err := database.AutoMigrate(&model.DataNexusUserModel{}); err != nil {
		return fmt.Errorf("migrate datanexus users table failed: %w", err)
	}

	fmt.Println("finish init mysql")
	return nil
}

// Get 返回数据库实例。
func Get() *gorm.DB {
	return dbInstance
}
