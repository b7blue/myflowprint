package main

import (
	"myflowprint/flowprintfactory"
	_ "myflowprint/model"
	"myflowprint/monitor"
)

/*
	一次完整的train
	0 手机连电脑热点,电脑抓包(wireshark或自己写都行),保存为pcap文件
	1 自动化操作 fingerer.Figer()
	2 monitor.CatchSess(catchname, true, false, pcapfile)
	3 flowprintfactory.Fingerprint()
*/

func main() {

	// totrainlist := utils.Excel_2_train_info("apptotrain.xlsx")
	// for _, info := range totrainlist {
	// 	// 因为公司和笔记本上有的指纹不一样，序号也不一样，以本机traininfo为准
	// 	id := model.NewAppInfo(info.Appname, info.Packagename) //新建条目
	// 	log.Println("收集流量：", id, info.Appname)
	// 	fingerer.Finger(id, info.Appname, info.Packagename) //模拟点击
	// 	// monitor.CatchSess(true, id, info.Appname+".pcap")      //提取会话记录
	// 	// flowprintfactory.Fingerprint(id, true)                 //生成指纹
	// }
	monitor.CatchSess(true, 7, "抖音(世界杯高清直播)"+".pcap")
	flowprintfactory.Fingerprint(7, true)

}
