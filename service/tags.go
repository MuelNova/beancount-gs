package service

import (
	"github.com/beancount-gs/script"
	"github.com/gin-gonic/gin"
)

type Tags struct {
	Value string `bql:"distinct tags" json:"value"`
}

// QueryTags godoc
// @Summary     获取所有标签
// @Description 返回当前账本中使用过的所有交易标签
// @Tags        标签
// @Produce     json
// @Security    LedgerId
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Router      /api/auth/tags [get]
func QueryTags(c *gin.Context) {
	ledgerConfig := script.GetLedgerConfigFromContext(c)
	tags := make([]Tags, 0)
	err := script.BQLQueryList(ledgerConfig, nil, &tags)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	result := make([]string, 0)
	for _, t := range tags {
		if t.Value != "" {
			result = append(result, t.Value)
		}
	}

	OK(c, result)
}
