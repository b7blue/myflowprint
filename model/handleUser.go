package model

import (
	"crypto/sha256"
	"fmt"
)

// 数据库存储hash值

func NewUser(uid, pw string) bool {
	// 不存在
	if db.Where("uid = ?", uid).First(&User{}).Error != nil {
		nu := User{
			Uid: uid,
			Pw:  fmt.Sprintf("%x", sha256.Sum256([]byte(pw))),
		}
		if db.Save(nu).Error == nil {
			return true

		}
	}
	return false
}

func CheckPw(uid, pw string) bool {
	trueuser := User{}
	if db.Where("uid = ?", uid).First(&trueuser).Error != nil {
		return false
	} else {
		return trueuser.Pw == fmt.Sprintf("%x", sha256.Sum256([]byte(pw)))
	}
}
