package monitor

import (
	"myflowprint/model"
	"sync"
	"time"

	"github.com/google/gopacket"
)

// map加锁
type SyncMap struct {
	Sess map[uint64]map[uint64]*model.Id_ses
	Lock *sync.RWMutex
}

func NewSyncMap() *SyncMap {
	return &SyncMap{
		Sess: make(map[uint64]map[uint64]*model.Id_ses),
		Lock: &sync.RWMutex{},
	}
}

func (d SyncMap) GetID(netHash, transportHash uint64) int {
	d.Lock.RLock()
	defer d.Lock.RUnlock()
	if d.Sess[netHash] == nil || d.Sess[netHash][transportHash] == nil {
		return -1
	}
	return d.Sess[netHash][transportHash].GoroutineID
}

func (d SyncMap) GetSess(netHash, transportHash uint64) *model.Session {
	d.Lock.RLock()
	defer d.Lock.RUnlock()
	if d.Sess[netHash] == nil || d.Sess[netHash][transportHash] == nil {
		return nil
	}
	return d.Sess[netHash][transportHash].Session
}

func (d SyncMap) DelKey(netHash, transportHash uint64) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	delete(d.Sess[netHash], transportHash)
	if len(d.Sess[netHash]) == 0 {
		delete(d.Sess, netHash)
	}
}

func (d SyncMap) AllKey() map[uint64][]uint64 {
	d.Lock.RLock()
	defer d.Lock.RUnlock()
	keys := make(map[uint64][]uint64, len(d.Sess))
	for nh := range d.Sess {
		keys[nh] = make([]uint64, len(d.Sess[nh]))
		i := 0
		for th := range d.Sess[nh] {
			keys[nh][i] = th
			i++
		}
	}
	return keys
}

func (d SyncMap) IsNew(netHash, transportHash uint64) bool {
	d.Lock.RLock()
	defer d.Lock.RUnlock()
	return d.Sess[netHash] == nil || d.Sess[netHash][transportHash] == nil
}

func (d SyncMap) NewSess(netHash, transportHash uint64, protocol uint8, netFlow, transportFlow gopacket.Flow, starttime time.Time, tempid int) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	if d.Sess[netHash] == nil {
		d.Sess[netHash] = make(map[uint64]*model.Id_ses)
	}
	d.Sess[netHash][transportHash] = &model.Id_ses{}
	d.Sess[netHash][transportHash].Session = &model.Session{
		Aip:      ip2uint(netFlow.Src().Raw()),
		Bip:      ip2uint(netFlow.Dst().Raw()),
		Aport:    port2uint(transportFlow.Src().Raw()),
		Bport:    port2uint(transportFlow.Dst().Raw()),
		Protocol: protocol,
		Start:    starttime,
	}
	// 轮询-负载均衡
	d.Sess[netHash][transportHash].GoroutineID = tempid
}

func (d SyncMap) UpdateSess(netHash, transportHash uint64, upflow bool, length int, end time.Time) bool {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	if d.Sess[netHash] == nil || d.Sess[netHash][transportHash] == nil {
		return false
	}
	thissess := d.Sess[netHash][transportHash].Session
	if upflow {
		thissess.Uflow += length
		thissess.Upacket++
	} else {
		thissess.Dflow += length
		thissess.Dpacket++
	}
	thissess.End = end
	return true
}

func (d SyncMap) EndSess(netHash, transportHash uint64, sawStart, sawEnd bool) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	if !sawStart {
		d.Sess[netHash][transportHash].Session.Start = time.Time{} //赋予零值
	}
	if !sawEnd {
		d.Sess[netHash][transportHash].Session.End = time.Time{}
	}
}
