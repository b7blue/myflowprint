package monitor

import (
	"fmt"
	"log"
	"myflowprint/model"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
)

// var HOST net.IP = net.IPv4(192, 168, 137, 135)
// var iface = flag.String("i", "eth0", "Interface to get packets from")
// var fname = flag.String("r", "", "Filename to read from, overrides -i")
// var handle *pcap.Handle
var err error

var host = layers.NewIPEndpoint(net.IPv4(192, 168, 137, 14))
var tcpFlush = time.Duration(2) * time.Minute   //tcp会话超时时间
var udpFlush = time.Duration(2) * time.Minute   //tcp会话超时时间
const τbatch = time.Duration(300) * time.Second //一次捕获的时长
const concurrency = 10

/*
train: 是否用于生成指纹库
id: appid或是detectid
pcapfile: 待分析pcap文件
*/
func CatchSess(train bool, id int, pcapfile string) {

	var tcpSess = NewSyncMap() //暂存会话信息
	var udpSess = NewSyncMap() //暂存会话信息

	var saveSess = make(chan model.Session, 100) //会话判断结束后,通过这个channel发送存储进数据库
	var allDone = make(chan struct{})            //用于主进程等待所有协程执行完毕
	var catchDone = make(chan struct{})          //用于广播抓包的结束，让每个goroutine都停止
	var pkgDone = make(chan struct{})            //用于提醒flushudp结束前最后一次刷新
	var wg_pkg sync.WaitGroup                    //用于等待所有处理数据包的协程执行完毕
	var wg_util sync.WaitGroup                   //用于等待所有工具协程（刷新、入库）执行完毕

	var handle *pcap.Handle
	if pcapfile != "" {
		log.Printf("Reading from pcap dump %s", pcapfile)
		var f *os.File
		if train {
			f, err = os.Open(fmt.Sprintf("traindata/%s", pcapfile))
		} else {
			f, err = os.Open(fmt.Sprintf("detectdata/%s", pcapfile))
		}

		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		handle, err = pcap.OpenOfflineFile(f)
	} else {
		// 电脑热点
		// \Device\NPF_{63523F40-8580-4CD8-9E3F-4DE53B19BEA2}
		// wifi
		// \Device\NPF_{F7D6E13E-DAA8-48FE-9AED-DF428409187D}
		// 指定设备抓包
		log.Printf("Starting capture on interface")
		handle, err = pcap.OpenLive(`\Device\NPF_{63523F40-8580-4CD8-9E3F-4DE53B19BEA2}`, 65535, true, pcap.BlockForever)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// 咱就是说设置一个过滤器，只捕获手机收发的数据包
	// handle.SetBPFFilter("host 192.168.137.14 ")

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	log.Println("packetsource ready")

	// 初始化构建assembly
	streamFactory := &statsStreamFactory{
		tcpSess:  tcpSess,
		saveSess: saveSess,
	}
	streamPools := make([]*tcpassembly.StreamPool, concurrency)
	// assembler.MaxBufferedPagesPerConnection = *bufferedPerConnection
	// assembler.MaxBufferedPagesTotal = *bufferedTotal

	assemblers := make([]*tcpassembly.Assembler, concurrency)
	pkchans := make([]chan gopacket.Packet, concurrency)

	// udp会话超时判断用的goroutine
	wg_util.Add(1)
	go flushUDP(&wg_util, pkgDone, saveSess, tcpSess, udpSess)

	for i := 0; i < concurrency; i++ {
		/**  为啥每一个assembler都配一个streampool：防止全锁了
		If you can guarantee that packets going to a set of Assemblers will contain information on different connections per Assembler (for example, they're already hashed by PF_RING hashing or some other hashing mechanism),
		then we recommend you use a seperate StreamPool per Assembler, thus avoiding all lock contention. Only when different Assemblers could receive packets for the same Stream should a StreamPool be shared between them.
		**/
		pkchans[i] = make(chan gopacket.Packet, 10)
		streamPools[i] = tcpassembly.NewStreamPool(streamFactory)
		assemblers[i] = tcpassembly.NewAssembler(streamPools[i])
		wg_pkg.Add(1)

		go func(id int) {
			defer wg_pkg.Done()
			ticker := time.NewTicker(tcpFlush)
		Loop:
			for {
				select {
				case packet := <-pkchans[id]:
					// 非常重要，因为pkchans关闭后所有接收不被阻塞，会接收nil产生空指针异常
					if packet == nil {
						break
					}
					netFlow := packet.NetworkLayer().NetworkFlow()
					transportLayer := packet.TransportLayer()
					transportFlow := transportLayer.TransportFlow()
					netHash, transportHash := netFlow.FastHash(), transportFlow.FastHash()
					aip := netFlow.Src()
					if aip != host {
						netFlow = netFlow.Reverse()
						transportFlow = transportFlow.Reverse()
					}
					var protocol uint8
					// 首先区分tcp还是udp，然后更新会话记录，假如是tcp就交给assembler
					if transportLayer.LayerType() == layers.LayerTypeTCP {
						protocol = 6

						// 更新会话记录：字节数、包数
						if !tcpSess.UpdateSess(netHash, transportHash, aip == host, packet.Metadata().Length, packet.Metadata().Timestamp) {
							tcpSess.NewSess(netHash, transportHash, protocol, netFlow, transportFlow, packet.Metadata().Timestamp, id)
							tcpSess.UpdateSess(netHash, transportHash, aip == host, packet.Metadata().Length, packet.Metadata().Timestamp)
						}
						// 让assembler处理tcp重组
						assemblers[id].AssembleWithTimestamp(netFlow, transportLayer.(*layers.TCP), packet.Metadata().Timestamp)
					} else {
						protocol = 17

						// 更新会话记录：字节数、包数、end时间
						if !udpSess.UpdateSess(netHash, transportHash, aip == host, packet.Metadata().Length, packet.Metadata().Timestamp) {
							udpSess.NewSess(netHash, transportHash, protocol, netFlow, transportFlow, packet.Metadata().Timestamp, id)
							udpSess.UpdateSess(netHash, transportHash, aip == host, packet.Metadata().Length, packet.Metadata().Timestamp)
						}

					}
				case <-ticker.C:
					// 每隔一分钟按照设定的超时时间刷新停止活动的tcp会话,判断它们超时
					assemblers[id].FlushOlderThan(time.Now().Add(-tcpFlush))

				// 假如收到通知，主协程分发数据包完毕
				case <-catchDone:
					// log.Println("gorountine", id, "开始结束流程")
					// 1 （非常重要）处理完pkchans[id]剩余内容
					for packet := range pkchans[id] {
						netFlow := packet.NetworkLayer().NetworkFlow()
						transportLayer := packet.TransportLayer()
						transportFlow := transportLayer.TransportFlow()
						netHash, transportHash := netFlow.FastHash(), transportFlow.FastHash()
						aip := netFlow.Src()
						if aip != host {
							netFlow = netFlow.Reverse()
							transportFlow = transportFlow.Reverse()
						}
						var protocol uint8
						// 首先区分tcp还是udp，然后更新会话记录，假如是tcp就交给assembler
						if transportLayer.LayerType() == layers.LayerTypeTCP {
							protocol = 6
							// 更新会话记录：字节数、包数
							if !tcpSess.UpdateSess(netHash, transportHash, aip == host, packet.Metadata().Length, packet.Metadata().Timestamp) {
								tcpSess.NewSess(netHash, transportHash, protocol, netFlow, transportFlow, packet.Metadata().Timestamp, id)
								tcpSess.UpdateSess(netHash, transportHash, aip == host, packet.Metadata().Length, packet.Metadata().Timestamp)
							}
							// 让assembler处理tcp重组
							assemblers[id].AssembleWithTimestamp(netFlow, transportLayer.(*layers.TCP), packet.Metadata().Timestamp)
						} else {
							protocol = 17
							// 更新会话记录：字节数、包数、end时间
							if !udpSess.UpdateSess(netHash, transportHash, aip == host, packet.Metadata().Length, packet.Metadata().Timestamp) {
								udpSess.NewSess(netHash, transportHash, protocol, netFlow, transportFlow, packet.Metadata().Timestamp, id)
								udpSess.UpdateSess(netHash, transportHash, aip == host, packet.Metadata().Length, packet.Metadata().Timestamp)
							}

						}
					}
					// 2 停止定时器
					ticker.Stop()
					// 3 跳出循环，结束这个协程
					break Loop
				}
			}
			log.Println("gorountine", id, "结束")
		}(i)
	}

	// 获得这次catch的表名
	var tablename string
	if train {
		tablename = model.NewTrain(id)
	} else {
		// 检查用，正常情况是flowprintservice搞得
		// model.NewDetect(time.Now())
		tablename = fmt.Sprintf("detectsess_%d", id)
	}

	// 将会话存入数据库的goroutine, 有缓存channel
	wg_util.Add(1)
	go model.Session2db(tablename, &wg_util, saveSess, catchDone)

	// 主线程closer
	go func() {
		wg_pkg.Wait()
		close(pkgDone)
		wg_util.Wait()
		close(allDone)
	}()

	// 因为服务端的端口比较单一所以用服务端端口作为endpoint做实验
	// i := 0
	packets := packetSource.Packets()
	tempid := 0
	timetostop := time.After(τbatch) //一次捕获最多τbatch时间

	// 轮询packet找到其中最早的timestamp，表示捕获开始时间
	starttime := time.Now()
	log.Println("开始分发数据包")
Loop:
	for packet := range packets {

		// 到时间跳出循环
		select {
		case <-timetostop:
			break Loop
		default:
		}

		// 首先判断有没得网络层和传输层内，不然可能会空指针啊
		if netLayer, transportLayer := packet.NetworkLayer(), packet.TransportLayer(); netLayer != nil && transportLayer != nil {
			// 查找最早的数据包，也就是捕获的开始时间
			if packet.Metadata().Timestamp.Before(starttime) {
				starttime = packet.Metadata().Timestamp
			}

			netFlow := netLayer.NetworkFlow()
			transportFlow := transportLayer.TransportFlow()
			netHash, transportHash := netFlow.FastHash(), transportFlow.FastHash()
			if netFlow.Src() != host {
				netFlow = netFlow.Reverse()
				transportFlow = transportFlow.Reverse()
			}

			// 该数据包应该给谁处理
			goroutineID := tempid
			var protocol uint8

			// tcp和udp会话分开处理，因为udp会话只用超时来判定，而tcp会话用tcpassembly处理
			if transportLayer.LayerType() == layers.LayerTypeTCP {
				protocol = 6

				// 假如是新会话，初始化
				if tcpSess.IsNew(netHash, transportHash) {
					tcpSess.NewSess(netHash, transportHash, protocol, netFlow, transportFlow, packet.Metadata().Timestamp, tempid)
					tempid = (tempid + 1) % concurrency
				} else {
					if goroutineID = tcpSess.GetID(netHash, transportHash); goroutineID == -1 {
						tcpSess.NewSess(netHash, transportHash, protocol, netFlow, transportFlow, packet.Metadata().Timestamp, tempid)
						goroutineID = tempid
						tempid = (tempid + 1) % concurrency
					}

				}

			} else if transportLayer.LayerType() == layers.LayerTypeUDP {
				protocol = 17
				// 假如是新会话，初始化
				if udpSess.IsNew(netHash, transportHash) {
					udpSess.NewSess(netHash, transportHash, protocol, netFlow, transportFlow, packet.Metadata().Timestamp, tempid)
					tempid = (tempid + 1) % concurrency
				} else {
					if goroutineID = udpSess.GetID(netHash, transportHash); goroutineID == -1 {
						udpSess.NewSess(netHash, transportHash, protocol, netFlow, transportFlow, packet.Metadata().Timestamp, tempid)
						goroutineID = tempid
						tempid = (tempid + 1) % concurrency
					}

				}

			}

			// 多线程处理所有数据包，根据goroutineID来选择协程
			pkchans[goroutineID] <- packet

		}

	}

	// If the underlying PacketDataSource returns an io.EOF error, the channel will be closed.
	// 数据包没了之后，packets这个channel将关闭，此时应该将done关闭，向所有协程广播停止信息。然后每个线程都做好收尾工作。
	close(catchDone)
	for i := 0; i < concurrency; i++ {
		close(pkchans[i])
	}
	log.Println("数据包分发完毕，等待协程处理完毕")

	// 等待所有协程执行完毕，阻塞中
	<-allDone

	// 轮询packet找到其中最早的timestamp，作为捕获的开始时间最精确
	if !train {
		model.StartCapture_Detect(id, starttime)
	} else {
		model.StartCapture_Train(id, starttime)
	}

	// fmt.Println("capture end")
	// for portendpoint := range tcpSession {
	// 	fmt.Printf("%v %v\n", portendpoint.String(), tcpSession[portendpoint])
	// }
	// fmt.Println(len(tcpSession))

}

// map的key为aport+bip+bport组成的uint64好了（sip就是机子的ip，固定的）
// 发现key不存在，则新建一个会话记录结构体
// 三种结束方式：超时（了解重传次数）、rst、fin（三次或四次挥手）
// 一次tcp会话结束，就持久性存储
// 这个过程用生产者消费者吧，存数据库是一个单独的线程
