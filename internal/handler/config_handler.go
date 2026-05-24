package handler

import (
	"net/http"

	"github.com/example/go-diamond/internal/model"
	"github.com/example/go-diamond/internal/service"
	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configSvc *service.ConfigService
}

func NewConfigHandler(configSvc *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configSvc: configSvc}
}

func (h *ConfigHandler) GetConfig(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	dataID := c.Param("dataId")

	cfg, err := h.configSvc.GetConfig(c.Request.Context(), namespace, group, dataID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if cfg == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"dataId":     cfg.DataID,
			"group":      cfg.Group,
			"namespace":  cfg.Namespace,
			"content":    cfg.Content,
			"contentMd5": cfg.ContentMD5,
			"version":    cfg.Version,
			"format":     cfg.Format,
			"updatedAt":  cfg.UpdatedAt,
		},
	})
}

func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var req struct {
		Namespace   string `json:"namespace" binding:"required"`
		Group       string `json:"group" binding:"required"`
		DataID      string `json:"dataId" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Format      string `json:"format"`
		Description string `json:"description"`
		Operator    string `json:"operator" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if req.Format == "" {
		req.Format = "json"
	}

	cfg := &model.Config{
		Namespace:   req.Namespace,
		Group:       req.Group,
		DataID:      req.DataID,
		Content:    req.Content,
		Format:      req.Format,
		Description: req.Description,
		CreatedBy:   req.Operator,
		UpdatedBy:   req.Operator,
	}

	if err := h.configSvc.CreateConfig(c.Request.Context(), cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"version": cfg.Version}})
}

func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	dataID := c.Param("dataId")

	var req struct {
		Content     string `json:"content" binding:"required"`
		Format      string `json:"format"`
		Description string `json:"description"`
		Operator    string `json:"operator" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	existing, err := h.configSvc.GetConfig(c.Request.Context(), namespace, group, dataID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "config not found"})
		return
	}

	cfg := &model.Config{
		Namespace:   namespace,
		Group:       group,
		DataID:      dataID,
		Content:     req.Content,
		Format:      req.Format,
		Description: req.Description,
		UpdatedBy:   req.Operator,
	}

	if err := h.configSvc.UpdateConfig(c.Request.Context(), cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"version": cfg.Version}})
}

func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	dataID := c.Param("dataId")

	var req struct {
		Operator string `json:"operator" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := h.configSvc.DeleteConfig(c.Request.Context(), namespace, group, dataID, req.Operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "deleted"})
}

func (h *ConfigHandler) ListConfigs(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")
	group := c.DefaultQuery("group", "DEFAULT_GROUP")
	page := 1
	pageSize := 20

	configs, total, err := h.configSvc.ListConfigs(c.Request.Context(), namespace, group, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total": total,
			"list":  configs,
		},
	})
}

func (h *ConfigHandler) GetHistories(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	dataID := c.Param("dataId")

	cfg, err := h.configSvc.GetConfig(c.Request.Context(), namespace, group, dataID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if cfg == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "config not found"})
		return
	}

	page := 1
	pageSize := 20

	histories, total, err := h.configSvc.GetHistories(c.Request.Context(), cfg.ID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total": total,
			"list":  histories,
		},
	})
}

func (h *ConfigHandler) Rollback(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	dataID := c.Param("dataId")

	var req struct {
		HistoryID uint64 `json:"historyId" binding:"required"`
		Operator  string `json:"operator" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := h.configSvc.Rollback(c.Request.Context(), namespace, group, dataID, req.HistoryID, req.Operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "rollback success"})
}