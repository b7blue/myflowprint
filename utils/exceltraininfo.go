package utils

import (
	"fmt"
	"log"
	"myflowprint/model"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// excel表格导入，id, 包名， app名
func Excel_2_train_info(filepath string) []model.TrainInfo {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		log.Println("open xlsx file fail: ", err)
		return nil
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	activesheet := f.GetSheetName(f.GetActiveSheetIndex())
	if activesheet == "" {
		log.Println("no active sheet in xlsx file: ", filepath)
		return nil
	}
	rows, err := f.GetRows(activesheet)
	if err != nil {
		log.Println("read xlsx file fail:", err)
	}

	info := make([]model.TrainInfo, len(rows))

	for i, row := range rows {
		id, _ := strconv.Atoi(row[0])
		info[i] = model.TrainInfo{
			ID:          id,
			Appname:     row[2],
			Packagename: row[1],
		}

	}
	log.Println("get web infos from xlsx file succeed")
	return info

}
