package fingerer

import (
	"fmt"
	"os/exec"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// *gorm.DB是多线程安全的，您可以在多个例程中使用一个*gorm.DB。您可以将其初始化一次，并在需要时获取它。
var db *gorm.DB = initdb()

// create user 'flowprintmaker'@'localhost' identified by 'flowprint';
// 域问题是host还是啥
var dsn string = "flowprintmaker:flowprint@tcp(127.0.0.1:3306)/flowprint?charset=utf8mb4&parseTime=True"

func initdb() *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("数据库连接成功")

	}

	err = db.AutoMigrate(&pkgname{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("初始化pkgname完成")

	}

	return db
}

// 3.18数据库里有测试存的29个
// truncate table pkgnames; 重置表
func GetAllPkgs() {

	cmd := exec.Command("adb", "shell", " pm list package -3")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	pkgnames := strings.Split(string(stdoutStderr), "package:")
	pkgs := make([]pkgname, len(pkgnames)-1)
	pkgnames = pkgnames[1:]
	for i, pn := range pkgnames {
		pkgs[i].Name = strings.Trim(pn, "\n")

	}
	if err = db.Create(pkgs).Error; err != nil {
		fmt.Println(err)
	}

	fmt.Println(len(pkgs))
	for _, p := range pkgs {
		fmt.Println(p.Name)
	}

}

type pkgname struct {
	ID   int    `gorm:"primarykey;auto_increment"`
	Name string `gorm:"type:varchar(50)"`
}
