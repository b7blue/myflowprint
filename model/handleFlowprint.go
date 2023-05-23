package model

import "sort"

type flowprints []Flowprint

func (p flowprints) Len() int {
	return len(p)
}
func (p flowprints) Less(i, j int) bool {
	return p[i].Pid < p[j].Pid
}
func (p flowprints) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// 更改指纹的某一对ip和port
func UpdateFingerprint(appname string, pid int, oldip, newip uint32, oldport, newport uint16) error {
	oldprint := Flowprint{
		Appname: appname,
		Pid:     pid,
		Dip:     oldip,
		Dport:   oldport,
	}
	return db.Model(&oldprint).Updates(Flowprint{Dip: newip, Dport: newport}).Error
}

// 删除某app指纹的一个二元组
func DelFingerprint(appname string, pid int, ip uint32, port uint16) error {
	print := Flowprint{
		Appname: appname,
		Pid:     pid,
		Dip:     ip,
		Dport:   port,
	}
	return db.Where("appname = ? AND pid = ? AND dip = ? AND dport = ?", appname, pid, ip, port).Delete(&print).Error
}

// 删除该app的指纹
func DelAppPrints(appname string) error {
	print := Flowprint{
		Appname: appname,
	}
	return db.Delete(&print).Error
}

// 指纹存入数据库
func StoreFigerprint(fingerprint [][]Flowprint) error {
	for i := range fingerprint {
		if err := db.Create(&fingerprint[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// 获取某一app的指纹详细情况,根据pid进行排序
func GetAppPrints(appname string) []Flowprint {
	var appfp []Flowprint
	db.Where("appname = ?", appname).Find(&appfp)
	sort.Sort(flowprints(appfp))

	return appfp
}

// 将指纹库加载进内存 map存储
// 为了后续指纹对比方便，直接把ip：port对化成uint64
func GetFingerprintDB() map[string][][]uint64 {
	var allfp []Flowprint
	db.Find(&allfp)
	fpdb := make(map[string][][]uint64)

	for _, fp := range allfp {
		if fpdb[fp.Appname] == nil {
			fpnum := GetPrintsNum(fp.Appname)
			fpdb[fp.Appname] = make([][]uint64, fpnum)
			for i := range fpdb[fp.Appname] {
				fpdb[fp.Appname][i] = make([]uint64, 0)
			}
		}
		// 注意pid是从1开始的，当时加了1，这里要减回来
		fpdb[fp.Appname][fp.Pid-1] = append(fpdb[fp.Appname][fp.Pid-1], ip_port2dst(fp.Dip, fp.Dport))
	}
	return fpdb
}

// 指纹跟已有指纹库对比成功则将指纹包含的所有ip：port对产生的会话在数据库中标记为属于该app
func GetSessBelonging(tablename, appname string, fp []Flowprint) (err error) {
	for _, f := range fp {
		err = db.Table(tablename).Model(&Session{}).Where("bip = ? AND bport = ?", f.Dip, f.Dport).Update("appname", appname).Error
		if err != nil {
			return
		}

	}
	return
}

func ip_port2dst(ip uint32, port uint16) uint64 {
	return uint64(ip)<<32 | uint64(port)
}
