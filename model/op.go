package model

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 数据库名称flowprint
// 会话数据库里有6种不同的表, 分别存储了：
// 1、训练用的app名称与id对应关系 trainlist
// 2、某次检测与id对应 detectlist,
// 3. 训练用色会话 trainsess_id
// 4. 检测用会话 detectsess_id
// 5. 应用程序指纹 print
// 6. 所有待测应用的包名 pkgname (在fingerer那个包里面写)

// *gorm.DB是多线程安全的，您可以在多个例程中使用一个*gorm.DB。您可以将其初始化一次，并在需要时获取它。
var db *gorm.DB = initdb()

// create user 'flowprintmaker'@'localhost' identified by 'flowprint';
// 域问题是host还是啥
var dsn string = "flowprintmaker:flowprint@tcp(127.0.0.1:3306)/flowprint?charset=utf8mb4&parseTime=True"

func initdb() *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("数据库连接成功")

	}

	// 新建所需的表
	err = db.AutoMigrate(&TrainInfo{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("初始化trainlist完成")

	}
	err = db.AutoMigrate(&DetectInfo{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("初始化detectlist完成")

	}
	err = db.AutoMigrate(&Flowprint{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("初始化flowprint完成")

	}
	// PrintsInfo
	err = db.AutoMigrate(&PrintsInfo{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("初始化PrintsInfo完成")

	}
	// users
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("初始化User完成")

	}

	return db
}

func GetConnection() *gorm.DB {
	return db
}
