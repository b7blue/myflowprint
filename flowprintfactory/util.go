package flowprintfactory

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func ip_port2dst(ip uint32, port uint16) uint64 {
	return uint64(ip)<<32 | uint64(port)
}

type DstCluster struct {
	ip      uint32
	port    uint16
	c_t     []int // ci[t], 时间片t内是否有会话
	sessnum int
}

// type graph struct {
// 	vertex    []*DstCluster
// 	edge      [][]float64 //表示目的地集群之间的相关系数
// 	vertexnum int
// }
