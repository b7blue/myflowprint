package fingerer

import (
	"fmt"
	"math/rand"
	"myflowprint/model"
	"myflowprint/monitor"
	"os/exec"
	"time"
)

/**
操作分为以下几大类：
1、返回
2、随机点击页面
3、上下滑动
4、左右滑动
5、点击app内上下导航
**/
const backCMD = "input keyevent BACK"
const swipeDownCMD = "input swipe 540 400 540 1200"
const swipeUpCMD = "input swipe 540 1200 540 400"
const swipeRightCMD = "input swipe 300 1170 800 1170"
const swipeLeftCMD = "input swipe 800 1170 300 1170"

var opnum = 10

// 抓包参数
const τwindow = time.Duration(30) * time.Second
const τbatch = time.Duration(300) * time.Second

// 模拟操作相关时间设置
const testtime = 10
const gap = time.Duration(500) * time.Millisecond
const startwait = time.Duration(7) * time.Second
const fingertime = 100

func Finger(id int, appname, packagename string) {
	rand.Seed(time.Now().UnixNano())

	// searchcmd := exec.Command("adb shell input text baidu")
	catchdone := make(chan struct{})
	go monitor.CatchAll(id, appname, true, catchdone, τbatch)

Loop:
	for i := 0; i < testtime; i++ {
		select {
		case <-catchdone:
			break Loop
		default:
		}
		// 启动app
		exec.Command("adb", "shell", fmt.Sprintf("monkey -p %s -c android.intent.category.LAUNCHER 1", packagename)).Run()
		// 等待app启动的广告结束
		time.Sleep(startwait)

		for i := 0; i < fingertime; i++ {
			select {
			case <-catchdone:
				exec.Command("adb", "shell", fmt.Sprintf("am force-stop %s", packagename)).Run()
				break Loop
			default:
			}

			op := rand.Intn(opnum)
			opnum = 10
			switch op {
			case 9:
				exec.Command("adb", "shell", backCMD).Run()
				opnum = 9 //避免连续两次back，直接退出程序了
				fmt.Println("back")
			case 0, 1:
				exec.Command("adb", "shell", tapRandPos()).Run()
			case 2:
				exec.Command("adb", "shell", swipeDownCMD).Run()
				fmt.Println("swipe down")
			case 3, 8:
				exec.Command("adb", "shell", swipeUpCMD).Run()
				fmt.Println("swipe up")
			case 4:
				exec.Command("adb", "shell", swipeLeftCMD).Run()
				fmt.Println("swipe left")
			case 5:
				exec.Command("adb", "shell", swipeRightCMD).Run()
				fmt.Println("swipe right")
			case 6:
				exec.Command("adb", "shell", randTopNav()).Run()
			case 7:
				exec.Command("adb", "shell", randBottomNav()).Run()
			}
			time.Sleep(gap)

		}
		// adb shell am force-stop结束app
		exec.Command("adb", "shell", fmt.Sprintf("am force-stop %s", packagename)).Run()
	}
	model.CaptureDone(id)

}

// 获得一个随机坐标
// y 200-2150 x 150-950
func tapRandPos() string {
	x, y := rand.Intn(800)+150, rand.Intn(1850)+200
	fmt.Println("random tap", x, y)
	return fmt.Sprintf("input tap %d %d", x, y)
}

// 获得一个随机x坐标
// 80-1000
func randTopNav() string {
	x := rand.Intn(920) + 80
	fmt.Println("random top x", x)
	return fmt.Sprintf("input tap %d 160", x)
}

func randBottomNav() string {
	x := rand.Intn(920) + 80
	fmt.Println("random  bottom x", x)
	return fmt.Sprintf("input tap %d 2100", x)
}

// 改变app联网权限
