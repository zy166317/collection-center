package utils

import (
	"archive/zip"
	"bytes"
	"collection-center/internal/logger"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/pkg/errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
)

// excel上传
func ParseExcelToArray(req *http.Request, filename string) (arrays [][]string, fileBytes []byte, err error) {
	file, header, err := req.FormFile(filename)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	if path.Ext(header.Filename) != ".xlsx" {
		err = errors.New("文件格式不正确:" + header.Filename)
		logger.Info("文件格式不正确:" + header.Filename)
		return
	}
	defer file.Close()
	f, err := excelize.OpenReader(file)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	buffer, err := f.WriteToBuffer()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	fileBytes = buffer.Bytes()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, errors.New("表内没有 sheet！")
	}
	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		logger.Error(err.Error())
		return
	}
	if len(rows) == 0 {
		return nil, nil, errors.New("未读取到内容：空文件！")
	}
	arrays = rows
	return
}

func ParseExcel(req *http.Request, filename string) (arrays [][]string, fileBytes []byte, fileName string, err error) {
	file, header, err := req.FormFile(filename)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	if path.Ext(header.Filename) != ".xlsx" {
		err = errors.New("文件格式不正确:" + header.Filename)
		logger.Info("文件格式不正确:" + header.Filename)
		return
	}
	fileName = header.Filename
	defer file.Close()
	f, err := excelize.OpenReader(file)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	buffer, err := f.WriteToBuffer()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	fileBytes = buffer.Bytes()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, "", errors.New("表内没有 sheet！")
	}
	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		logger.Error(err.Error())
		return
	}
	if len(rows) == 0 {
		return nil, nil, "", errors.New("未读取到内容：空文件！")
	}
	arrays = rows
	return
}

// 导出 excel 字节
func WriteOutBytes(w http.ResponseWriter, bytes *[]byte, fileName string) (err error) {
	w.Header().Set("Content-Disposition", "attachment;filename="+url.QueryEscape(fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(*bytes)
	return
}

// 生成excel字节  上链
func WriteArrayToBytes(arrays *[][]string) (err error, bytes []byte) {
	f := excelize.NewFile()
	sheetName := "sheet"
	// 创建一个工作表
	index := f.NewSheet(sheetName)
	// 设置工作簿的默认工作表
	f.SetActiveSheet(index)
	var cellName string
	for i, row := range *arrays {
		for j, colCell := range row {
			cellName, err = excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				fmt.Println(err)
				return
			}
			// 设置单元格的值
			err = f.SetCellValue(sheetName, cellName, colCell)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	buffer, err := f.WriteToBuffer()
	if err != nil {
		logger.Error("WriteToBuffer error:", err.Error())
		return
	}
	bytes = buffer.Bytes()
	return
}

/*
 * 根据模版写入 excel 并进行汇总
 * readPath 模版读取路径
 * startRow 开始行
 * startCol 开始列
 * arrays 数据
 * flag  是否删除模版文件
 */
func WriteExcelToBytesAndSum(readPath string, arrays [][]string, startRow int, startCol int, flag bool) (fileBytes []byte, err error) {
	f, err := excelize.OpenFile(readPath)
	if err != nil {
		return nil, errors.Wrap(err, "WriteExcelToBytes")
	}
	sheets := f.GetSheetList()
	for i, row := range arrays {
		for j, colCell := range row {
			cellName, err1 := excelize.CoordinatesToCellName(startCol+j, startRow+i)
			if err1 != nil {
				err = err1
				return
			}
			// 设置单元格的值
			if j > 0 {
				colCellFloat, err1 := strconv.ParseFloat(colCell, 64)
				if err1 != nil {
					err2 := f.SetCellValue(sheets[0], cellName, colCell)
					if err2 != nil {
						fmt.Println(err2)
						err = err2
						return
					}
				} else {
					//
					colName, rowInt, err1 := excelize.SplitCellName(cellName)
					if err1 != nil {
						err = err1
						return
					}
					if rowInt == startRow {
						endRow := len(arrays) + startRow - 1
						last := len(arrays) + startRow
						formulaStr := "=SUM(" + cellName + ":" + colName + strconv.Itoa(endRow) + ")"
						err1 := f.SetCellFormula(sheets[0], colName+strconv.Itoa(last), formulaStr)
						if err1 != nil {
							err = err1
							return
						}
						err2 := f.MergeCell(sheets[0], "A"+strconv.Itoa(last), "D"+strconv.Itoa(last))
						if err2 != nil {
							err = err2
							return
						}
						err3 := f.SetCellValue(sheets[0], "A"+strconv.Itoa(last), "合计")
						if err3 != nil {
							err = err3
							return
						}
					}
					err2 := f.SetCellValue(sheets[0], cellName, colCellFloat)
					if err2 != nil {
						err = err2
						return
					}
				}
			} else {
				err2 := f.SetCellValue(sheets[0], cellName, colCell)
				if err2 != nil {
					err = err2
					return
				}
			}

		}
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, errors.Wrap(err, "WriteExcelToBytes")
	}
	if flag {
		err := os.Remove(readPath)
		if err != nil {
			logger.Error("os Remove err:", err.Error())
			return nil, err
		}
	}
	return buffer.Bytes(), nil
}

/*
 * 根据模版写入 excel
 * readPath 模版读取路径
 * startRow 开始行
 * startCol 开始列
 * arrays 数据
 * flag  是否删除模版文件
 */
func WriteExcelToBytes(readPath string, arrays [][]string, startRow int, startCol int, flag bool) (fileBytes []byte, err error) {
	f, err := excelize.OpenFile(readPath)
	if err != nil {
		return nil, errors.Wrap(err, "WriteExcelToBytes")
	}
	sheets := f.GetSheetList()
	for i, row := range arrays {
		if row != nil {
			cellName, err1 := excelize.CoordinatesToCellName(startCol, startRow+i)
			if err1 != nil {
				logger.Error("CoordinatesToCellName error: ", err1.Error())
				err = err1
				return
			}
			err2 := f.SetSheetRow(sheets[0], cellName, &row)
			if err2 != nil {
				logger.Error("SetSheetRow error: ", err2.Error())
				err = err2
				return
			}
		}
	}
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, errors.Wrap(err, "WriteExcelToBytes")
	}
	if flag {
		err := os.Remove(readPath)
		if err != nil {
			logger.Error("os Remove err:", err.Error())
			return nil, err
		}
	}
	return buffer.Bytes(), nil
}

/*
 * 根据模版写入 excel
 * readPath 模版读取路径
 * startRow 开始行
 * startCol 开始列
 * arrays 数据
 */
func ConvertDataToExcelBytes(templatePath string, arrays [][]string, startRow int, startCol int) (fileBytes []byte, err error) {
	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return nil, errors.Wrap(err, "ConvertDataToExcelBytes")
	}
	for i, row := range arrays {
		if row != nil {
			cellName, err1 := excelize.CoordinatesToCellName(startCol, startRow+i)
			if err1 != nil {
				logger.Error("CoordinatesToCellName error: ", err1.Error())
				err = err1
				return
			}
			err2 := f.SetSheetRow(f.GetSheetList()[0], cellName, &row)
			if err2 != nil {
				logger.Error("SetSheetRow error: ", err2.Error())
				err = err2
				return
			}
		}
	}
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, errors.Wrap(err, "WriteExcelToBytes")
	}
	return buffer.Bytes(), nil
}

func CreateZipBytes(maps map[string][]byte) (zipBytes []byte, err error) {
	outBuffer := bytes.NewBuffer(nil)
	zipWriter := zip.NewWriter(outBuffer)
	for k, v := range maps {
		create, err1 := zipWriter.Create(k)
		if err1 != nil {
			err = err1
			return
		}
		_, err2 := create.Write(v)
		if err2 != nil {
			err = err2
			return
		}
		zipWriter.Flush()
	}
	zipWriter.Close()
	zipBytes = outBuffer.Bytes()
	return
}

/**
* @Author SF
* @Description 读取上传excel 保存文件头
* @Date 11:27 2021/7/10
* @Param req
* @Param filename 文件名
* @Param row  删除开始行 包含
**/
func SaveHeader(reader io.Reader, row int, tempPath string, done chan error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		logger.Error(err.Error())
		done <- err
		return
	}
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		logger.Error("表内没有 sheet！")
		//return errors.New("表内没有 sheet！")
		done <- errors.New("表内没有 sheet！")
		return
	}
	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		logger.Error(err.Error())
		done <- err
		return
	}
	if len(rows) == 0 {
		logger.Error("未读取到内容：空文件！")
		done <- errors.New("未读取到内容：空文件！")
		return
	}

	for a := len(rows); row <= a; a-- {
		err1 := f.RemoveRow(sheets[0], a)
		if err1 != nil {
			logger.Error("RemoveRow", err1.Error())
			done <- err1
			return
		}
	}
	err = f.SaveAs(tempPath)
	if err != nil {
		logger.Error(err.Error())
	}
	done <- err
	return
}

/**
* @Author SF
* @Description 获取excel文件流
* @Date 14:18 2021/7/10
* @Param
* @return
**/
func GetExcelFile(req *http.Request, filename string) (file multipart.File, fileName string, err error) {
	file, header, err := req.FormFile(filename)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	if path.Ext(header.Filename) != ".xlsx" {
		err = errors.New("文件格式不正确:" + header.Filename)
		logger.Info("文件格式不正确:" + header.Filename)
		return
	}
	fileName = header.Filename
	defer file.Close()
	return
}

/**
* @Author SF
* @Description 读取文件流 获取数据
* @Date 14:35 2021/7/10
* @Param
* @return
**/
func GetArrayFromFile(reader io.Reader) (arrays [][]string, fileBytes []byte, err error) {

	f, err := excelize.OpenReader(reader)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	buffer, err := f.WriteToBuffer()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	fileBytes = buffer.Bytes()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, errors.New("表内没有 sheet！")
	}
	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		logger.Error(err.Error())
		return
	}
	if len(rows) == 0 {
		return nil, nil, errors.New("未读取到内容：空文件！")
	}
	arrays = rows
	return
}
