package utils

import (
	"fmt"
	"log"
	"myflowprint/model"

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
	num := 0
	for i, row := range rows {
		if row[0] != "" {
			num++
			info[i] = model.TrainInfo{
				Appname:     row[1],
				Packagename: row[0],
			}
		} else {
			break
		}

	}
	log.Println("get app infos from xlsx file succeed")
	return info[:num]

}
