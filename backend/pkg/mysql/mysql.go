package mysql

import (
	"backend/config"
	"backend/internal/dao"
	"backend/internal/model"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitMySQL 初始化MySQL连接
func InitMySQL() {
	dsn := config.DBConfig.Username + ":" + config.DBConfig.Password +
		"@tcp(" + config.DBConfig.Host + ":" + config.DBConfig.Port + ")/" +
		config.DBConfig.Dbname + "?charset=" + config.DBConfig.Charset + "&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("mysql connect err: %v", err)
	}

	dao.DB = db

	//自动迁移数据库表
	err = db.AutoMigrate(&model.User{}, &model.Monitor{}, &model.MonitorHistory{},
		&model.Alert{}, &model.AlertConfig{}, &model.Incident{}, &model.AlertSendTask{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	err = db.Migrator().CreateTable(&model.AlertConfig{})
	if err != nil {
		log.Printf("AlertConfig表创建/更新失败（非致命）: %v", err)
		// 用Printf而非Fatalf，不终止程序
	}
	log.Println("MySQL连接成功，数据库表迁移完成")
}
