package flowprintfactory

import (
	"fmt"
	"log"
	"math"
	"myflowprint/model"
)

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func unionSameFingerprint(flowprints [][]model.Flowprint) [][]model.Flowprint {
	// 将所有指纹做相似度分析
	num := len(flowprints)
	orinum := num
	// 记录没有被合并的剩下的
	exist := make([]bool, num)
	for i := range exist {
		exist[i] = true
	}
	for i := 0; i < num; i++ {
		// flowprints[i]没有被合并
		if exist[i] {
			for j := 0; j < num; j++ {
				if exist[j] && i != j {
					issame := IsSameByJaccard_One(flowprints, i, j)
					if issame {
						exist[j] = false
						num--
					}
				}
			}
		}
	}

	re := make([][]model.Flowprint, 0, num)
	for i := range exist {
		if exist[i] {
			re = append(re, flowprints[i])
		}
	}
	if orinum == num {
		return re
	}
	re = unionSameFingerprint(re)
	log.Println("根据相似度", τsimilarity, "合并后的指纹个数：", num)
	return re
}

func getAppFingerprint(flowprints [][]model.Flowprint, appname string) [][]model.Flowprint {
	re := unionSameFingerprint(flowprints)
	for i := range re {
		for j := range re[i] {
			re[i][j].Appname = appname
			re[i][j].Pid = i + 1
		}
	}
	// fmt.Println(re)

	model.NewPrintsInfo(appname, len(re))
	return re
}

// 计算两个指纹之间的雅卡德系数, 用于合并app的多个指纹
func IsSameByJaccard_One(allfp [][]model.Flowprint, i1, i2 int) bool {
	// intersection, union
	f1, f2 := allfp[i1], allfp[i2]
	m := make(map[uint64]int)
	for _, f := range f1 {
		dst := ip_port2dst(f.Dip, f.Dport)
		m[dst]++
	}
	for _, f := range f2 {
		dst := ip_port2dst(f.Dip, f.Dport)
		m[dst]++
	}
	union := len(m)
	intersection := 0
	// 出现两次说明是交集
	for _, time := range m {
		if time == 2 {
			intersection++
		}
	}
	jaccard := float64(intersection) / float64(union)
	// 将并集放在f1中

	if almostEqual(math.Max(jaccard, τsimilarity), jaccard) {
		for _, f := range f2 {
			dst := ip_port2dst(f.Dip, f.Dport)
			if m[dst] == 1 {
				f1 = append(f1, f)
			}
		}
		allfp[i1] = f1
		// log.Printf("两个指纹的相似度为%.3f, 判断为相似，合并后的指纹含有%d个网络目的地，指纹内容如下：\n%v", jaccard, len(f1), f1)
		return true
	}
	return false
}

// 计算两个指纹之间的雅卡德系数, 得出是否属于同一app
func IsSameByJaccard_All(appfp []uint64, fp []model.Flowprint) (bool, float64) {
	// intersection, union
	m := make(map[uint64]int)
	for _, dst := range appfp {
		m[dst]++
	}
	for _, f := range fp {
		dst := ip_port2dst(f.Dip, f.Dport)
		m[dst]++
	}
	union := len(m)
	intersection := 0
	// 出现两次说明是交集
	for _, time := range m {
		if time == 2 {
			intersection++
		}
	}
	jaccard := float64(intersection) / float64(union)
	return math.Max(jaccard, τsimilarity) == jaccard, jaccard
}

// （app识别中）得出该指纹与哪个已知指纹最相似
func Belong2WhichApp(thisfp []model.Flowprint) (appname string) {
	fpdb := model.GetFingerprintDB()

	var maxjaccard float64
	for app, appfp := range fpdb {
		for i := range appfp {
			issame, jaccard := IsSameByJaccard_All(appfp[i], thisfp)
			fmt.Println(app, jaccard)
			if issame {
				if math.Max(jaccard, maxjaccard) == jaccard {
					appname = app
					maxjaccard = jaccard
				}
			}
		}
	}
	return appname
}
