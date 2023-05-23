package model

import (
	"log"
	"sync"
)

// 一个线程运行此方法,将结束的会话插入数据库中
// 调用之前记得wg.Add()
// 注意：sync.WaitGroup是值传递，这里我们必须传入地址
func Session2db(tablename string, wg *sync.WaitGroup, saveSess chan Session, catchDone chan struct{}) {

	/* 从channel中接收会话记录,批量插入 */
	i := 0
	sessions := make([]Session, 0, 100)
Loop:
	for {
		select {
		case <-catchDone:
			// 记得将剩下的所有会话入库
			i := len(sessions)
			for thisSess := range saveSess {
				i++
				sessions = append(sessions, thisSess)
				if i == 100 {
					db.Table(tablename).Create(&sessions)
					i = 0
					sessions = sessions[0:0]
				}

			}
			if i > 0 {
				db.Table(tablename).Create(sessions)
			}

			break Loop

		case thisSess := <-saveSess:
			i++
			sessions = append(sessions, thisSess)
			if i == 100 {
				db.Table(tablename).Create(sessions)
				i = 0
				sessions = sessions[0:0]
			}
		}
	}
	log.Println("入库完毕")
	wg.Done()
}

// 取出指定表的所有会话记录
func GetAllSess(tablename string) []*Session {
	allSess := make([]*Session, 0)
	db.Table(tablename).Find(&allSess)
	return allSess
}

// 根据给定的app名称查找会话
func GetSessByApp(tablename, appname string) []*Session {
	allSess := make([]*Session, 0)
	db.Table(tablename).Where("appname = ?", appname).Find(&allSess)
	return allSess
}

// 根据给定的ip查找会话
func GetSessByIP(tablename string, ip uint32) []*Session {
	allSess := make([]*Session, 0)
	db.Table(tablename).Where("bip = ?", ip).Find(&allSess)
	return allSess
}

// 根据给定的时间范围查找会话
// func GetSessByTimeZone(tablename string, ip uint32) []*Session {
// 	allSess := make([]*Session, 0)
// 	db.Table(tablename).Where("bip = ?", ip).Find(&allSess)
// 	return allSess
// }
