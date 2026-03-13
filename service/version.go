package service

import "github.com/gin-gonic/gin"

// QueryVersion godoc
// @Summary     获取服务版本号
// @Description 返回当前 beancount-gs 服务版本号
// @Tags        系统
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Router      /api/version [get]
func QueryVersion(c *gin.Context) {
	OK(c, "v1.2.2")
}
