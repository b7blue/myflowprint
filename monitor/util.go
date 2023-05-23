package monitor

import (
	"log"
	"myflowprint/model"
	"sync"
	"time"
)

// 将ipv4地址转换为uint32
func ip2uint(ip []byte) uint32 {
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func port2uint(port []byte) uint16 {
	return uint16(port[0])<<8 | uint16(port[1])
}

// udp会话超时
func flushUDP(wg_util *sync.WaitGroup, pkgDone chan struct{}, saveSess chan model.Session, tcpSess, udpSess *SyncMap) {
	defer wg_util.Done()
	ticker := time.NewTicker(udpFlush)
Loop:
	for {
		select {
		case <-pkgDone:
			ticker.Stop()
			// 将map中临时存储的全部入库
			UDPSess2DB(udpSess, saveSess)
			TCPSess2DB(tcpSess, saveSess) // 重点：有可能最后tcp会话在map中有残留，所以在结束前全部入库
			close(saveSess)
			break Loop
		case <-ticker.C:
			UDPSess2DB(udpSess, saveSess)
		}
	}
	log.Println("udpflush结束")

}

func TCPSess2DB(tcpSess *SyncMap, saveSess chan model.Session) {
	allkeys := tcpSess.AllKey()
	for netHash := range allkeys {
		for _, transportHash := range allkeys[netHash] {
			sess := *tcpSess.GetSess(netHash, transportHash)
			if sess.End.Before(time.Now().Add(-tcpFlush)) {
				saveSess <- sess
				tcpSess.DelKey(netHash, transportHash)

			}
		}
	}
}

func UDPSess2DB(udpSess *SyncMap, saveSess chan model.Session) {
	allkeys := udpSess.AllKey()
	for netHash := range allkeys {
		for _, transportHash := range allkeys[netHash] {
			sess := *udpSess.GetSess(netHash, transportHash)
			if sess.End.Before(time.Now().Add(-udpFlush)) {
				// 存入数据库
				saveSess <- sess
				// 方案1:直接删除map中这一项,但是非常不巧删除之前该会话又有新的数据包到达,就会引起空指针异常
				// 也许可以在catch中多判断几次?
				udpSess.DelKey(netHash, transportHash)

			}
		}
	}
}
