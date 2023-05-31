package main

import (
	"fmt"
	"myflowprint/model"
)

func main() {
	// cmd := exec.Command("go", "run", "test2.go")
	// cmd.Start()
	// fmt.Println("test1结束了")
	// fmt.Println(len("88b1cca59060320e5e5662a7da636884eb7580f4dc7e22cfb6f88b8f99045a71"))
	// args := os.Args
	// for i := range args {
	// 	fmt.Println(i, args[i])
	// }
	// fmt.Println(os.IsExist(os.Mkdir("../flowprintservice/detectdata", 0666)))

	fpdb := model.GetFingerprintDB()
	for app, appfp := range fpdb {
		fmt.Println(app, "的指纹如下：")
		for i := range appfp {
			fmt.Printf("%d号指纹包含%d个网络目的地：\n%v\n", i, len(appfp[i]), appfp[i])
		}
	}
}
