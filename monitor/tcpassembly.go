package monitor

import (
	"log"
	"myflowprint/model"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
)

// tcpassembly要用的构造的结构体与方法
// simpleStreamFactory implements tcpassembly.StreamFactory
// 这字段值都是为了把参数传进来，怒。。。
type statsStreamFactory struct {
	tcpSess  *SyncMap
	saveSess chan model.Session
}

// statsStream处理tcp会话信息.
type statsStream struct {
	net, transport   gopacket.Flow
	start, end       time.Time
	sawStart, sawEnd bool
	tcpSess          *SyncMap
	saveSess         chan model.Session
}

// New creates a new stream.  It's called whenever the assembler sees a stream
// it isn't currently following.
func (factory *statsStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	log.Printf("new stream %v:%v started", net, transport)
	s := &statsStream{
		net:       net,
		transport: transport,
		start:     time.Now(),
		tcpSess:   factory.tcpSess,
		saveSess:  factory.saveSess,
	}
	s.end = s.start
	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return s
}

// Reassembled is called whenever new packet data is available for reading.
// Reassembly objects contain stream data IN ORDER.
func (s *statsStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	for _, reassembly := range reassemblies {
		s.end = reassembly.Seen
		s.sawStart = s.sawStart || reassembly.Start
		s.sawEnd = s.sawEnd || reassembly.End
	}
}

// ReassemblyComplete is called when the TCP assembler believes a stream has finished.
// 判断一个tcp会话结束后，存入数据库
// 不完成的tcp会话超时后会被flush结束，假如有syn就填开始时间，有fin或rst就填结束时间
func (s *statsStream) ReassemblyComplete() {
	netHash := s.net.FastHash()
	transportHash := s.transport.FastHash()

	s.tcpSess.EndSess(netHash, transportHash, s.sawStart, s.sawEnd)
	thisSess := *(s.tcpSess).GetSess(netHash, transportHash)

	// 存入数据库
	s.saveSess <- thisSess
	// 删除map中的key
	s.tcpSess.DelKey(netHash, transportHash)

}
