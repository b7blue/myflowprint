package monitor

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

// 捕获用于app检测的数据包
// 手机连接电脑热点
func Catch4Detect(id int) {
	catchDone := make(chan struct{})
	go CatchAll(id, "", false, catchDone, time.Duration(300)*time.Second)
	<-catchDone

}

// finger模拟触控时或者抓用户包时调用，抓包做成pcap文件，根据τbatch调整抓包时长
func CatchAll(id int, appname string, train bool, catchdone chan struct{}, τbatch time.Duration) {
	// Write a new file:
	var f *os.File
	if train {
		f, err = os.Create(fmt.Sprintf("traindata/%s.pcap", appname))
	} else {
		f, err = os.Create(fmt.Sprintf("detectdata/%d.pcap", id))
	}
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.Println("创建pcap文件成功")

	w := pcapgo.NewWriter(f)
	if err := w.WriteFileHeader(65536, layers.LinkTypeEthernet); err != nil {
		log.Fatalf("WriteFileHeader: %v", err)
	} // new file, must do this.

	// 打开网卡抓包
	// \Device\NPF_{63523F40-8580-4CD8-9E3F-4DE53B19BEA2} 笔记本
	// \Device\NPF_{62373F76-6A01-47D7-922E-648AE11AC519} 公司
	handle, err := pcap.OpenLive(`\Device\NPF_{62373F76-6A01-47D7-922E-648AE11AC519}`, 65535, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	tc := time.After(τbatch) //定时器
	log.Println("抓包开始，倒计时五分钟")

	//开始计时(发现在后来的catchsess都要再走一次，在那里遍历packet获得最早时间，作为捕获开始时间更加精确)
	// if train {
	// 	model.StartCapture_Train(id, time.Now())
	// } else {
	// 	model.StartCapture_Detect(id, time.Now())
	// }
Loop:
	for {
		select {
		case <-tc:
			break Loop
		default:
			packet := <-packets
			if packet.NetworkLayer() != nil && packet.TransportLayer() != nil {
				if err := w.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
					log.Fatalf("pcap.WritePacket(): %v", err)
				}
			}
		}

	}
	log.Println("抓包结束")
	// 向主goroutine发送结束信号
	catchdone <- struct{}{}

}
