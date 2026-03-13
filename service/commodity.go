package service

import (
	"fmt"
	"github.com/beancount-gs/script"
	"github.com/gin-gonic/gin"
)

type SyncCommodityPriceForm struct {
	Commodity string `form:"commodity" binding:"required" json:"commodity"`
	Date      string `form:"date" binding:"required" json:"date"`
	Price     string `form:"price" binding:"required" json:"price"`
}

// SyncCommodityPrice godoc
// @Summary     同步商品价格
// @Description 向账本写入一条新的 price 条目，并刷新货币汇率缓存
// @Tags        商品
// @Accept      json
// @Produce     json
// @Security    LedgerId
// @Param       body body SyncCommodityPriceForm true "商品价格表单"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Router      /api/auth/commodity/price [post]
func SyncCommodityPrice(c *gin.Context) {
	var syncCommodityPriceForm SyncCommodityPriceForm
	if err := c.ShouldBindJSON(&syncCommodityPriceForm); err != nil {
		BadRequest(c, err.Error())
		return
	}

	ledgerConfig := script.GetLedgerConfigFromContext(c)
	filePath := script.GetLedgerPriceFilePath(ledgerConfig.DataPath)
	line := fmt.Sprintf("%s price %s %s %s", syncCommodityPriceForm.Date, syncCommodityPriceForm.Commodity, syncCommodityPriceForm.Price, ledgerConfig.OperatingCurrency)
	// 写入文件
	err := script.AppendFileInNewLine(filePath, line)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	// 刷新货币最新汇率值
	err = script.LoadLedgerCurrencyMap(ledgerConfig)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	OK(c, syncCommodityPriceForm)
}

// QueryAllCurrencies godoc
// @Summary     获取所有货币列表
// @Description 返回当前账本中所有货币及其最新汇率
// @Tags        商品
// @Produce     json
// @Security    LedgerId
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Router      /api/auth/commodity/currencies [get]
func QueryAllCurrencies(c *gin.Context) {
	ledgerConfig := script.GetLedgerConfigFromContext(c)
	// 查询货币获取当前汇率
	currency := script.RefreshLedgerCurrency(ledgerConfig)
	OK(c, currency)
}
