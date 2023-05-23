package model

func Count(tablename string) int64 {
	var re int64
	db.Table(tablename).Count(&re)
	return re
}

// func PageData(tablename string) interface{} {
// 	if tablename == "flowprint"
// }
