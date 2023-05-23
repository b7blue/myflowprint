package model

func NewPrintsInfo(appname string, printsnum int) {
	db.Save(&PrintsInfo{Appname: appname, Printnum: printsnum})
}

func GetPrintsNum(appname string) int {
	var pinfo PrintsInfo
	db.Model(&PrintsInfo{}).Where("appname = ?", appname).First(&pinfo)
	return pinfo.Printnum
}

func DelPrintsInfo(appname string) {
	pinfo := PrintsInfo{
		Appname: appname,
	}
	db.Delete(&pinfo)
}
