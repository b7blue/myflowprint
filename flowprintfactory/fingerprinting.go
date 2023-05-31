package flowprintfactory

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"myflowprint/model"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const τwindow = time.Duration(30) * time.Second
const τbatch = time.Duration(300) * time.Second
const τcorrelation float64 = 0.1
const τsimilarity float64 = 0.7
const slicenum = int(τbatch / τwindow)

// 1 聚类
// 2 集群关联，移除所有弱相关系数的边
// 3 找极大团

func Fingerprint(id int, train bool) string {
	start := time.Now()
	tablename := ""
	appname := ""
	var starttime time.Time
	if train {
		tablename = fmt.Sprintf("trainsess_%d", id)
		appname, starttime = model.GetAppInfo(id)
	} else {
		tablename = fmt.Sprintf("detectsess_%d", id)
		starttime = model.GetCapStart_Detect(id)
	}
	clusters := clustering(tablename, starttime)
	vertex, edge := correlation(clusters)
	flowprints := getMaximalCliques(vertex, edge)

	// 获得这次捕获的得到的指纹后，有两种情况
	// 1 本次捕获用于指纹库生成 - 将所有指纹做相似度分析，最后获得的对应的指纹入库
	// 2 本次捕获用于app识别 - 每个指纹跟已有指纹库做对比，对比成功则将指纹包含的所有ip：port对产生的会话在数据库中标记为属于该app
	if train {
		// appFingerprint :=
		getAppFingerprint(flowprints, appname)
		log.Printf("%d号指纹生成成功\n", id)
		// if err := model.StoreFigerprint(appFingerprint); err != nil {
		// 	log.Println("指纹入库失败")
		// 	log.Fatalln(err)
		// }
		// // 标记该app已经生成过指纹
		// model.PrintsDone(id)
		// log.Printf("%d号指纹入库成功\n", id)

	} else {
		flowprints = unionSameFingerprint(flowprints)
		log.Printf("第%d次app检测完毕\n", id)
		for _, thisfp := range flowprints {
			if appname := Belong2WhichApp(thisfp); appname != "" {
				fmt.Println(appname)
				// 一种情况：（检测app时）要是生成的指纹之间的网络目的地有相互重叠的
				// 由于极大团的端点有可能部分重合
				// 结果这些指纹却命中不同的已知指纹
				// 那么一个网络目的地会被判断属于多个不同的app
				// 那怎么标记该网络目的地对应的会话所属app？
				if err := model.GetSessBelonging(tablename, appname, thisfp); err != nil {
					log.Println(err)
				}
			}
		}
		log.Printf("第%d次检测结果入库成功\n", id)
	}
	return time.Since(start).String()
}

func clustering(tablename string, starttime time.Time) map[uint64]*DstCluster {
	allSess := model.GetAllSess(tablename)
	log.Println("该次捕获共获得会话总数：", len(allSess))
	clusters := make(map[uint64]*DstCluster)
	timezone := make([]time.Time, slicenum)
	for i := range timezone {
		timezone[i] = starttime.Add(time.Duration(i) * τwindow)
	}
	for _, thisSess := range allSess {
		dst := uint64(thisSess.Bip)<<32 | uint64(thisSess.Bport)
		if clusters[dst] == nil {
			clusters[dst] = &DstCluster{
				ip:   thisSess.Bip,
				port: thisSess.Bport,
				c_t:  make([]int, slicenum),
			}
		}
		clusters[dst].sessnum++
		// 判断该会话在哪几个时间片中出现
		// 1 没有end
		// 2 没有start
		// 3 两个都有
		if thisSess.End.IsZero() {
			for i := slicenum - 1; i >= 0; i-- {
				if thisSess.Start.After(timezone[i]) {
					for j := i; j < slicenum; j++ {
						clusters[dst].c_t[j] = 1
					}
					break
				}
			}
		} else if thisSess.Start.IsZero() {
			for i := 0; i < slicenum; i++ {
				if thisSess.End.Before(timezone[i].Add(τwindow)) {
					for j := 0; j < i+1 && j < slicenum; j++ {
						clusters[dst].c_t[j] = 1
					}
					break
				}
			}
		} else {
			for i := slicenum - 1; i >= 0; i-- {
				if thisSess.Start.After(timezone[i]) {
					gap := int(thisSess.End.Sub(thisSess.Start) / τwindow)
					for j := i; j < i+gap+1 && j < slicenum; j++ {
						clusters[dst].c_t[j] = 1
					}
					break
				}
			}
		}
	}
	log.Println("该次捕获的ip:port对总数：", len(clusters))
	// for i := range clusters {
	// 	fmt.Println(clusters[i].ip, clusters[i].port, clusters[i].sessnum)
	// }
	return clusters
}

// 带权重的/带相关系数的edge暂时不用
func correlation(clusters map[uint64]*DstCluster) (vertex []*DstCluster, edge [][]int) {
	vnum := len(clusters)
	vertex = make([]*DstCluster, vnum)
	// edge_weight := make([][]float64, vnum)
	edge = make([][]int, vnum)
	for i := range edge {
		// edge_weight[i] = make([]float64, vnum)
		edge[i] = make([]int, vnum)
	}
	i := 0
	for _, c := range clusters {
		vertex[i] = c
		i++
	}

	for i := 0; i < vnum; i++ {
		for j := i + 1; j < vnum; j++ {
			ci_cj := 0
			ci_cj_max := 0
			for k := 0; k < slicenum; k++ {
				ci_cj += vertex[i].c_t[k] * vertex[j].c_t[k]
				ci_cj_max += max(vertex[i].c_t[k], vertex[j].c_t[k])
			}
			ci_cj_norm := float64(ci_cj) / float64(ci_cj_max)

			if math.Max(ci_cj_norm, τcorrelation) == ci_cj_norm {
				// edge_weight[i][j], edge_weight[j][i] = ci_cj_norm, ci_cj_norm
				edge[i][j], edge[j][i] = 1, 1
			}

		}
	}
	return
}

func getMaximalCliques(vertex []*DstCluster, edge [][]int) [][]model.Flowprint {
	// 调用networkx的find_cliques来找极大团
	// 用exec调用py脚本，边参数化成一个字符串传入
	// 查到的极大团结果通过执行输出获得
	vnum := len(vertex)
	builder := strings.Builder{}

	for i := 0; i < vnum; i++ {
		for j := i + 1; j < vnum; j++ {
			if edge[i][j] == 1 {
				// 参数形式
				// v1,v2
				// v1,v4
				// v2,v3
				builder.WriteString(fmt.Sprintf("%d,%d.", i, j))
			}
		}
	}
	// 扣掉多出来那个\n
	if err := ioutil.WriteFile("extension.txt", []byte(builder.String()[:builder.Len()-1]), 0666); err != nil {
		fmt.Println("Writefile Error =", err)
	}
	//
	cmd := exec.Command("python", "flowprintfactory/getMaximalCliques.py")
	raw, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("py脚本调用出错")
		log.Fatalln(err, string(raw))
	}
	/*
		[3, 1, 2]
		[3, 4]
		[3, 5]
		结果输出形式，一行一个极大团所有的节点
		一个节点 = 一个目的地集团 = 一个ip：port对
		一个指纹 = 一个极大团 = 极大团中ip：port对的集合
	*/
	cliquesStr := strings.Split(string(raw), "]")
	if cliquesStr[len(cliquesStr)-1] == "" {
		cliquesStr = cliquesStr[:len(cliquesStr)-1]
	}
	// fmt.Println(string(raw), len(cliquesStr))

	flowprints := make([][]model.Flowprint, len(cliquesStr)-1)
	// cliques := make([][]int, )
	for i := 0; i < len(cliquesStr)-1; i++ {
		cs := cliquesStr[i]
		c := strings.Split(strings.Trim(strings.TrimSpace(cs), "[]"), ",")
		// fmt.Println(c)
		flowprints[i] = make([]model.Flowprint, len(c))
		for j := range c {
			v, err := strconv.Atoi(strings.TrimSpace(c[j]))
			if err != nil {
				log.Fatalln(err)
			}
			flowprints[i][j] = model.Flowprint{
				Dip:   vertex[v].ip,
				Dport: vertex[v].port,
			}
		}
	}
	log.Println("初步生成的指纹个数：", len(flowprints))

	return flowprints

}
