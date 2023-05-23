package main

import (
	_ "myflowprint/model"
)

/*
	一次完整的train
	0 手机连电脑热点,电脑抓包(wireshark或自己写都行),保存为pcap文件
	1 自动化操作 fingerer.Figer()
	2 monitor.CatchSess(catchname, true, false, pcapfile)
	3 flowprintfactory.Fingerprint()
*/

func main() {
	// 测试用，正常应该一次性生成app list
	// model.NewAppInfo("哔哩哔哩", "tv.danmaku.bili")

	// // 生成指纹库
	// // 取出trainlist中还没capture的，遍历进行fingerer.Finger
	// fingerer.Finger(2, "哔哩哔哩", "tv.danmaku.bili")

	// // 取出trainlist还没fingerprint的，遍历进行monitor.CatchSess和flowprintfactory.Fingerprint
	// monitor.CatchSess(true, 103, "豆瓣_old.pcap")

	// flowprintfactory.Fingerprint(2, true)

	// s := ""

	// for i := 0; i < 10; i++ {
	// 	s += flowprintfactory.Fingerprint(1, true) + "、"
	// }
	// fmt.Println(s)

	// monitor.Catch4Detect(1)
	// 外部进程根据是否上传有文件，选择两种调用monitor.CatchSess()的方式
	// monitor.CatchSess(false, 1, "1.pcap")
	// flowprintfactory.Fingerprint(1, false)
	// 外部进程会话分析完毕后，将更新数据库，改变redis中detectid的状态

}
