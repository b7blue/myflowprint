package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func Install() {
	// re, err := exec.Command("adb", "install", "apk/1_5e8d4c33f23dfa47f370977d0bb26676.apk").CombinedOutput()
	// if err != nil {
	// 	log.Println(err, string(re))
	// }
	apklist, err := ioutil.ReadDir("apk")
	if err != nil {
		log.Println(err)
	}
	for _, apk := range apklist {
		apkpath := fmt.Sprintf("apk/%s", apk.Name())
		log.Println("installing", apkpath)
		re, err := exec.Command("adb", "install", apkpath).CombinedOutput()
		if err != nil {
			log.Println("fail:", string(re))
		} else {
			log.Println("secceed!")
			// 删掉文件
			os.Remove(apkpath)
		}
	}
}
