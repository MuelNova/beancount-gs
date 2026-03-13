package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/beancount-gs/script"
	"github.com/gin-gonic/gin"
)

type Event struct {
	Date        string   `form:"date" binding:"required" json:"date"`
	Stage       string   `form:"stage" json:"stage"`
	Type        string   `form:"type" json:"type"`
	Types       []string `form:"types" json:"types"`
	Description string   `form:"description" binding:"required" json:"description"`
}

// Events 切片包含多个事件
type Events []Event

func (e Events) Len() int {
	return len(e)
}

func (e Events) Less(i, j int) bool {
	return strings.Compare(e[i].Date, e[j].Date) < 0
}

func (e Events) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// GetAllEvents godoc
// @Summary     获取所有事件
// @Description 返回账本中所有记录的事件列表，按日期倒序排列
// @Tags        事件
// @Produce     json
// @Security    LedgerId
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Router      /api/auth/event/all [get]
func GetAllEvents(c *gin.Context) {
	ledgerConfig := script.GetLedgerConfigFromContext(c)

	beanFilePath := script.GetLedgerEventsFilePath(ledgerConfig.DataPath)
	bytes, err := script.ReadFile(beanFilePath)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	lines := strings.Split(string(bytes), "\n")
	events := Events{}
	// foreach lines
	for _, line := range lines {
		if strings.Trim(line, " ") == "" {
			continue
		}
		// split line by " "
		words := strings.Fields(line)
		if len(words) < 4 {
			continue
		}
		if words[1] != "event" {
			continue
		}
		events = append(events, Event{
			Date:        words[0],
			Type:        strings.ReplaceAll(words[2], "\"", ""),
			Description: strings.ReplaceAll(words[3], "\"", ""),
		})
	}
	if len(events) > 0 {
		// events 按时间倒序排列
		sort.Sort(sort.Reverse(events))
	}
	OK(c, events)
}

// AddEvent godoc
// @Summary     添加事件
// @Description 向账本中写入一条或多条新的 event 记录
// @Tags        事件
// @Accept      json
// @Produce     json
// @Security    LedgerId
// @Param       body body Event true "事件表单"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Router      /api/auth/event [post]
func AddEvent(c *gin.Context) {
	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		BadRequest(c, err.Error())
		return
	}

	ledgerConfig := script.GetLedgerConfigFromContext(c)
	filePath := script.GetLedgerEventsFilePath(ledgerConfig.DataPath)

	if event.Type != "" {
		event.Types = []string{event.Type}
	}

	// 定义Event类型的数组
	events := make([]Event, 0)

	if event.Types != nil {
		for _, t := range event.Types {
			events = append(events, Event{
				Date:        event.Date,
				Type:        t,
				Description: event.Description,
			})
			line := fmt.Sprintf("%s event \"%s\" \"%s\"", event.Date, t, event.Description)
			// 写入文件
			err := script.AppendFileInNewLine(filePath, line)
			if err != nil {
				InternalError(c, err.Error())
				return
			}
		}
	}

	OK(c, events)
}

// DeleteEvent godoc
// @Summary     删除事件
// @Description 从账本中删除匹配的 event 记录
// @Tags        事件
// @Accept      json
// @Produce     json
// @Security    LedgerId
// @Param       body body Event true "要删除的事件"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Router      /api/auth/event [delete]
func DeleteEvent(c *gin.Context) {
	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		BadRequest(c, err.Error())
		return
	}

	ledgerConfig := script.GetLedgerConfigFromContext(c)
	filePath := script.GetLedgerEventsFilePath(ledgerConfig.DataPath)

	line := fmt.Sprintf("%s event \"%s\" \"%s\"", event.Date, event.Type, event.Description)
	err := script.DeleteLinesWithText(filePath, line)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	OK(c, nil)
}
