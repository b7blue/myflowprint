package main

import (
	"myflowprint/flowprintfactory"
	"myflowprint/model"
	_ "myflowprint/model"
	"myflowprint/monitor"
	fingerer "myflowprint/p30fingerer"
	"myflowprint/utils"
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
	totrainlist := utils.Excel_2_train_info("5.24.xlsx")
	for _, info := range totrainlist {
		// 因为公司和笔记本上有的指纹不一样，序号也不一样，以本机traininfo为准
		id := model.NewAppInfo(info.Appname, info.Packagename) //新建条目
		fingerer.Finger(id, info.Appname, info.Packagename)    //模拟点击
		monitor.CatchSess(true, id, info.Appname+".pcap")      //提取会话记录
		flowprintfactory.Fingerprint(id, true)                 //生成指纹
	}
	// model.NewAppInfo("京东", "com.jingdong.app.mall")

	// 生成指纹库
	// 取出trainlist中还没capture的，遍历进行fingerer.Finger
	// fingerer.Finger(5, "京东", "com.jingdong.app.mall")

	// // 取出trainlist还没fingerprint的，遍历进行monitor.CatchSess和flowprintfactory.Fingerprint
	// monitor.CatchSess(true, 5, "京东.pcap")
	// monitor.CatchSess(true, 3, "腾讯新闻.pcap")

	// flowprintfactory.Fingerprint(1, true)

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
