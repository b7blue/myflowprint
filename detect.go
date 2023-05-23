package main

import (
	"fmt"
	"log"
	"myflowprint/flowprintfactory"
	"myflowprint/monitor"

	"os"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

func main() {
	// 用cmd异步调起外部进程，传入detectid、是否有pcap文件参数
	// 参数1：detectid 参数2：offline
	args := os.Args
	detectid, _ := strconv.Atoi(args[1])
	offline := args[2] == "y"
	pcapfile := ""
	if offline {
		fmt.Sprintf("%d.pcap", detectid)
	}

	// 外部进程根据是否上传有文件，选择两种调用monitor.CatchSess()的方式
	monitor.CatchSess(false, detectid, pcapfile)
	flowprintfactory.Fingerprint(detectid, false)
	// 外部进程会话分析完毕后，将更新数据库，改变redis中detectid的状态
	StopDetect(detectid)

}

func StopDetect(detectid int) {
	// 创建连接
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Printf("redis.Dial() error:%v", err)
		return
	}
	// 关闭连接
	defer c.Close()
	c.Do("DEL", detectid)
}
