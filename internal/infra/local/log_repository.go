package local

import (
	"fmt"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

type ExcelLogRepo struct {
	filePath string
}

func NewExcelLogRepo(filePath string) *ExcelLogRepo {
	// Make file if file not Exsist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		f := excelize.NewFile()
		sheet := "Sheet1"
		f.SetSheetName(f.GetSheetName(0), sheet)
		// Write header
		headers := []string{
			"Id", "ChatId", "ReqMessage", "ResMessage", "Prompt",
			"ReqToken", "ResToken", "SendTime", "ReceiveTime",
		}
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h)
		}
		_ = f.SaveAs(filePath)
	}
	return &ExcelLogRepo{filePath: filePath}
}

func (r *ExcelLogRepo) Insert(sessionId, reqMsg, resMsg string, reqToken, resToken int, sendTime, receiveTime time.Time) error {
	f, err := excelize.OpenFile(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %w", err)
	}

	sheet := "Sheet1"
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("failed to read rows: %w", err)
	}
	rowIndex := len(rows) + 1

	record := Record{
		Id:          time.Now().Format("20060102150405"),
		ChatId:      sessionId,
		ReqMessage:  reqMsg,
		ResMessage:  resMsg,
		Prompt:      "",
		ReqToken:    reqToken,
		ResToken:    resToken,
		SendTime:    sendTime,
		ReceiveTime: receiveTime,
	}

	values := []interface{}{
		record.Id, record.ChatId, record.ReqMessage, record.ResMessage, record.Prompt,
		record.ReqToken, record.ResToken, record.SendTime.Format(time.RFC3339), record.ReceiveTime.Format(time.RFC3339),
	}

	for i, val := range values {
		cell, _ := excelize.CoordinatesToCellName(i+1, rowIndex)
		f.SetCellValue(sheet, cell, val)
	}

	err = f.Save()
	if err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}
	return nil
}

type Record struct {
	Id          string
	ChatId      string
	ReqMessage  string
	ResMessage  string
	Prompt      string
	ReqToken    int
	ResToken    int
	SendTime    time.Time
	ReceiveTime time.Time
}
