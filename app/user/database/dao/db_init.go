package dao

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"sendswork/app/user/database/models"
	"sendswork/config"
	"strings"
	"time"
)

var _db *gorm.DB

func InitDB() {
	mConfig := config.Conf.MySQL["user"]
	host := mConfig.Host
	port := mConfig.Port
	database := mConfig.Database
	username := mConfig.UserName
	password := mConfig.Password
	charset := mConfig.Charset
	dsn := strings.Join([]string{username, ":", password, "@tcp(", host, ":", port, ")/", database, "?charset=" + charset + "&parseTime=true&loc=Asia%2FShanghai"}, "")
	err := Database(dsn)
	if err != nil {
		fmt.Println(err)
	}
}

func migration() { //数据库迁移
	// 自动迁移模式
	err := _db.Set("gorm:table_options", "charset=utf8mb4").
		AutoMigrate(&models.User{})
	//Create(&models.User{})
	if err != nil {
		logrus.Info("register table fail")
		os.Exit(0)
	}
	logrus.Info("register table success")
}

func Database(connString string) (err error) {
	db, err := gorm.Open(mysql.Open(connString), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(20)                  // 设置连接池，空闲连接数
	sqlDB.SetMaxOpenConns(100)                 // 设置最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Second * 30) // 设置连接的最大生命周期
	_db = db
	migration() // 调用数据库迁移函数
	return err
}

func NewDBClient(ctx context.Context) *gorm.DB {
	db := _db
	return db.WithContext(ctx)
}
