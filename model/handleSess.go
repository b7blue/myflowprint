package model

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

const limitlen int = 30

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

// 获得条件查询（ip, appname...）的语句
func getquerystr(ip uint32, appname string) string {
	querystr := strings.Builder{}
	if ip != 0 {
		querystr.WriteString(fmt.Sprintf("ip = '%d'", ip))
	}
	if appname != "" {
		if querystr.Len() != 0 {
			querystr.WriteString(" AND ")
		}
		querystr.WriteString(fmt.Sprintf("appname = '%s'", appname))
	}

	// if !tt_start.IsZero() && !tt_end.IsZero() {
	// 	if querystr.Len() != 0 {
	// 		querystr.WriteString(" AND ")
	// 	}
	// 	querystr.WriteString(fmt.Sprintf("test_time BETWEEN '%v' AND '%v'", tt_start, tt_end))
	// }
	// if !ct_start.IsZero() && !ct_end.IsZero() {
	// 	if querystr.Len() != 0 {
	// 		querystr.WriteString(" AND ")
	// 	}
	// 	querystr.WriteString(fmt.Sprintf("create_time BETWEEN '%v' AND '%v'", ct_start, ct_end))
	// }
	return querystr.String()
}

// 条件查询session获得总数量
func Count_sess_by_term(detectid int, ip uint32, appname string) int {
	querystr := getquerystr(ip, appname)
	var num int64
	if err := db.Table(fmt.Sprintf("detectsess_%d", detectid)).Where(querystr).Count(&num).Error; err != nil {
		log.Println("db find error:", err)
	}
	return int(num)
}

// 分页条件查询session
func Find_sess_by_term_page(detectid int, ip uint32, appname string, page int) []Session {
	querystr := getquerystr(ip, appname)
	result := make([]Session, 0)
	db.Table(fmt.Sprintf("detectsess_%d", detectid)).Where(querystr).Limit(limitlen).Offset((page - 1) * limitlen).Find(&result)
	return result
}

// 根据给定的时间范围查找会话
// func GetSessByTimeZone(tablename string, ip uint32) []*Session {
// 	allSess := make([]*Session, 0)
// 	db.Table(tablename).Where("bip = ?", ip).Find(&allSess)
// 	return allSess
// }
