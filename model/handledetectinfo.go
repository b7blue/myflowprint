package model

import (
	"fmt"
	"time"
)

// 为这次捕获创建一个新表
func NewDetect(start time.Time) int {
	thiscatch := DetectInfo{Start: start}
	db.Create(&thiscatch)

	tablename := fmt.Sprintf("detectsess_%d", thiscatch.ID)
	if err := db.Table(tablename).AutoMigrate(&Session{}); err != nil {
		fmt.Println(err)
	}
	fmt.Println("新建表存储app会话记录")
	return thiscatch.ID
}

// 有两种情况：实时和离线
// 1、实时。在catchsess轮询packet之前，给detectinfo填上start，表示捕获开始时间
// 2、离线。轮询packet找到其中最早的timestamp，结束后再给detectinfo填上start
func StartCapture_Detect(id int, start time.Time) {
	thisdetect := DetectInfo{
		ID: id,
	}
	db.Model(&thisdetect).Update("start", start)
}

func GetCapStart_Detect(id int) time.Time {
	var detectinfo DetectInfo
	db.First(&detectinfo, id)
	return detectinfo.Start
}
