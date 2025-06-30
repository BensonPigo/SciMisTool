package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CommonHandler 提供通用資訊的 API 介面
// Factories 欄位存放所有可使用的 Factory 名稱
// 在初始化時由主程式填入

type CommonHandler struct {
	Factories []string
}

// GetFactories 回傳所有可用的 Factory 名稱
func (h *CommonHandler) GetFactories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"factories": h.Factories})
}
