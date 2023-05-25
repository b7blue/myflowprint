package model

import (
	"fmt"
	"log"
	"time"
)

// 将app全部下载进手机之后，用脚本获得所有app名字与包名，写入trainlist
func NewAppInfo(appname string, packagename string) int {
	var zerotime time.Time
	zerotime, err := time.Parse("2006-01-02 15:04:05", "1970-01-01 01:01:01")
	if err != nil {
		log.Println(err, zerotime)
	}
	thisapp := TrainInfo{
		Appname:     appname,
		Packagename: packagename,
		Captured:    false,
		// Analysed:      false,
		Fingerprinted: false,
		Start:         zerotime,
	}
	db.Create(&thisapp)
	return thisapp.ID
}

// 获得所有还没有自动化访问过的app
func GetAllUncaptured() []TrainInfo {
	var uncap []TrainInfo
	db.Where(map[string]interface{}{"Captured": false}).First(&uncap)
	return uncap
}

// 获得所有（自动化访问后）还没生成指纹的app
func GetAllUnfingerprinted() []TrainInfo {
	var unfp []TrainInfo
	db.Where(map[string]interface{}{"Captured": true, "Fingerprinted": false}).First(&unfp)
	return unfp
}

// 获得所有已生成指纹的app（网页接口用）
func GetFLowprintAppList() []TrainInfo {
	var fplist []TrainInfo
	db.Where(map[string]interface{}{"Captured": true, "Fingerprinted": true}).Find(&fplist)
	return fplist
}

// 用finger自动访问app时，给traininfo填上start，表示捕获开始时间
func StartCapture_Train(id int, start time.Time) {
	thisapp := TrainInfo{
		ID: id,
	}
	db.Model(&thisapp).Update("start", start)
}

// 训练一个新的app指纹，建表存储会话
func NewTrain(id int) string {
	tablename := fmt.Sprintf("trainsess_%d", id)
	if err := db.Table(tablename).AutoMigrate(&Session{}); err != nil {
		fmt.Println(err)
	}
	fmt.Println("新建表存储app会话记录")
	return tablename
}

// 成功自动化访问后，在trainlist中标记为已捕获
func CaptureDone(id int) {
	thisapp := TrainInfo{
		ID: id,
	}
	db.Model(&thisapp).Update("Captured", true)
}

// 成功获得app指纹后，在trainlist中标记为已获得指纹
func PrintsDone(id int) {
	thisapp := TrainInfo{
		ID: id,
	}
	db.Model(&thisapp).Update("Fingerprinted", true)
}

func GetAppInfo(id int) (string, time.Time) {
	var traininfo TrainInfo
	db.First(&traininfo, id)
	return traininfo.Appname, traininfo.Start
}

// 网页接口用：根据id查询该app是否已经生成指纹
func IsAppPrintsExist(id int) string {
	var app TrainInfo
	if db.First(&app, id).Error != nil {
		return ""
	} else {
		if app.Fingerprinted {
			return app.Appname
		}
		return ""
	}
}

// 网页接口用：删除某app的指纹
func PrintsDel(id int) {
	thisapp := TrainInfo{
		ID: id,
	}
	db.Model(&thisapp).Update("Fingerprinted", false)
}
