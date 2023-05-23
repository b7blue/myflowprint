package model

import "time"

type Id_ses struct {
	GoroutineID int
	Session     *Session
}

type Session struct {
	ID int `gorm:"primarykey;auto_increment"`
	// a - client
	// b - server
	Aip      uint32    `gorm:"type:int unsigned;NOT NULL"`
	Bip      uint32    `gorm:"type:int unsigned;NOT NULL"`
	Aport    uint16    `gorm:"type:smallint unsigned;NOT NULL"`
	Bport    uint16    `gorm:"type:smallint unsigned;NOT NULL"`
	Protocol uint8     `gorm:"type:tinyint unsigned;NOT NULL"`
	Start    time.Time `gorm:"type:datetime(6)"` //有可能为空
	End      time.Time `gorm:"type:datetime(6)"`
	Uflow    int
	Dflow    int
	Upacket  int
	Dpacket  int
	Appname  string `gorm:"type:varchar(30)"`
}

type TrainInfo struct {
	ID          int       `gorm:"primarykey;auto_increment"`
	Appname     string    `gorm:"type:varchar(30)"` // 假如是抓训练数据，则为app名称，否则由用户在网页端口指定
	Packagename string    `gorm:"type:varchar(50)"` //
	Start       time.Time `gorm:"type:datetime(6)"` //捕获开始时间，在finger中填上
	Captured    bool      `gorm:"type:boolean"`     //是否自动化访问获得通信行为信息
	// Analysed      bool      `gorm:"type:boolean"`     //是否分析会话
	Fingerprinted bool `gorm:"type:boolean"` //是否生成指纹
}

type DetectInfo struct {
	ID int `gorm:"primarykey;auto_increment"`
	// 假如是抓训练数据，则为app名称，否则由用户在网页端口指定
	// Name  string    `gorm:"type:varchar(30)"`
	Start time.Time `gorm:"type:datetime(6);NOT NULL"`
}

type Flowprint struct {
	Appname string `gorm:"type:varchar(30)"`
	Pid     int
	Dip     uint32 `gorm:"type:int unsigned"`
	Dport   uint16 `gorm:"type:smallint unsigned"`
}

type PrintsInfo struct {
	Appname  string `gorm:"type:varchar(30);primarykey"`
	Printnum int
}

type User struct {
	Uid string `gorm:"type:varchar(15);primarykey"`
	Pw  string `gorm:"type:varchar(100)"`
}
