package service

import (
	"fmt"
	"github.com/beancount-gs/script"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
	"time"
)

// QueryLedgerSourceFileDir godoc
// @Summary     获取账本文件目录结构
// @Description 返回账本数据目录下所有 .bean 文件的相对路径列表
// @Tags        文件
// @Produce     json
// @Security    LedgerId
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Router      /api/auth/file/dir [get]
func QueryLedgerSourceFileDir(c *gin.Context) {
	ledgerConfig := script.GetLedgerConfigFromContext(c)
	result, err := dirs(ledgerConfig.DataPath, ledgerConfig.DataPath)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	OK(c, result)
}

func dirs(parent string, dirPath string) ([]string, error) {
	result := make([]string, 0)
	rd, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, dir := range rd {
		parentDir := dirPath + "/" + dir.Name()
		if dir.IsDir() {
			// 跳过备份文件夹
			if dir.Name() == "bak" {
				continue
			}
			files, err := dirs(parent, parentDir)
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
		} else {
			fmt.Println(parentDir)
			result = append(result, strings.ReplaceAll(parentDir, parent+"/", ""))
		}
	}
	return result, nil
}

// QueryLedgerSourceFileContent godoc
// @Summary     读取账本文件内容
// @Description 返回指定相对路径的 .bean 源文件内容
// @Tags        文件
// @Produce     json
// @Security    LedgerId
// @Param       path query string true "文件相对路径"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Router      /api/auth/file/content [get]
func QueryLedgerSourceFileContent(c *gin.Context) {
	ledgerConfig := script.GetLedgerConfigFromContext(c)
	queryParams := script.GetQueryParams(c)
	if queryParams.Path == "" {
		BadRequest(c, "params must not be blank")
		return
	}
	bytes, err := script.ReadFile(ledgerConfig.DataPath + "/" + queryParams.Path)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	OK(c, string(bytes))
}

type UpdateSourceFileForm struct {
	Path    string `form:"path" binding:"required"`
	Content string `form:"content"`
}

// UpdateLedgerSourceFileContent godoc
// @Summary     更新账本文件内容
// @Description 更新指定账本文件的内容，先自动备份再写入
// @Tags        文件
// @Accept      json
// @Produce     json
// @Security    LedgerId
// @Param       body body UpdateSourceFileForm true "文件更新表单"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Router      /api/auth/file [post]
func UpdateLedgerSourceFileContent(c *gin.Context) {
	ledgerConfig := script.GetLedgerConfigFromContext(c)

	var updateSourceFileForm UpdateSourceFileForm
	if err := c.ShouldBindJSON(&updateSourceFileForm); err != nil {
		BadRequest(c, err.Error())
		return
	}

	sourceFilePath := ledgerConfig.DataPath + "/" + updateSourceFileForm.Path
	targetFilePath := ledgerConfig.DataPath + "/bak/" + time.Now().Format("20060102150405") + "_" + strings.ReplaceAll(updateSourceFileForm.Path, "/", "_")
	// 备份数据
	if ledgerConfig.IsBak {
		err := script.CopyFile(sourceFilePath, targetFilePath)
		if err != nil {
			InternalError(c, err.Error())
			return
		}
	}

	err := script.WriteFile(sourceFilePath, updateSourceFileForm.Content)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	// 更新外币种源文件后，更新缓存
	if strings.Contains(updateSourceFileForm.Path, "currency.json") {
		err = script.LoadLedgerCurrencyMap(ledgerConfig)
		if err != nil {
			InternalError(c, err.Error())
			return
		}
	}

	OK(c, nil)
}
