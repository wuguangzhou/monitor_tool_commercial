package main

import (
	"backend/config"
	"backend/internal/service"
	"backend/pkg/mysql"
	"backend/pkg/redis"
	"backend/router"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
)

func main() {
	// 初始化配置
	config.InitConfig()

	log.Printf("=== 加载的配置参数 ===")
	log.Printf("MySQL配置：Host=%s, Port=%s, Username=%s, Dbname=%s",
		config.DBConfig.Host, config.DBConfig.Port, config.DBConfig.Username, config.DBConfig.Dbname)
	log.Printf("Redis配置：Host=%s, Port=%s", config.RDConfig.Host, config.RDConfig.Port)
	log.Printf("APP_PORT：%s", os.Getenv("APP_PORT"))
	log.Printf("======================")

	//初始化Mysql数据库连接（关键，关联Dao层DB实例）
	mysql.InitMySQL()

	//初始化Redis连接
	log.Println("开始初始化Redis连接...")
	err := redis.InitRedis()
	if err != nil {
		return
	}
	log.Println("Redis连接成功")

	//启动定时监控任务
	service.StartMonitorCron()

	//启动告警定时任务
	service.StartAlertCron()
	log.Println("告警定时任务初始化成功")

	r := router.InitRouter()

	// 重要：CORS中间件必须在路由初始化之后添加
	// 因为它需要作用于所有路由
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000/", "http://127.0.0.1:3000"},

		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},

		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
		},
		//允许前端携带Cookie/认证信息
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("服务器启动成功，监听端口：%s，访问地址：http://0.0.0.0:%s", port, port)
	log.Fatal(r.Run("0.0.0.0:" + port))

}

// initMySQL 初始化MySQL连接
//func initMySQL() {
//	dsn := config.DBConfig.Username + ":" + config.DBConfig.Password +
//		"@tcp(" + config.DBConfig.Host + ":" + config.DBConfig.Port + ")/" +
//		config.DBConfig.Dbname + "?charset=" + config.DBConfig.Charset + "&parseTime=True&loc=Local"
//	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
//		Logger: logger.Default.LogMode(logger.Info),
//	})
//	if err != nil {
//		log.Fatalf("mysql connect err: %v", err)
//	}
//
//	dao.DB = db
//
//	//自动迁移数据库表
//	err = db.AutoMigrate(&model.User{}, &model.Monitor{}, &model.MonitorHistory{},
//		&model.Alert{}, &model.AlertConfig{}, &model.Incident{}, &model.AlertSendTask{},
//	)
//	if err != nil {
//		log.Fatalf("数据库迁移失败: %v", err)
//	}
//	err = db.Migrator().CreateTable(&model.AlertConfig{})
//	if err != nil {
//		log.Printf("AlertConfig表创建/更新失败（非致命）: %v", err)
//		// 用Printf而非Fatalf，不终止程序
//	}
//	log.Println("MySQL连接成功，数据库表迁移完成")
//}
