package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/example/go-diamond/internal/service"
	"github.com/example/go-diamond/internal/watcher"
	"github.com/gin-gonic/gin"
)

type WatchHandler struct {
	configSvc *service.ConfigService
	hub       *watcher.WatcherHub
}

func NewWatchHandler(configSvc *service.ConfigService, hub *watcher.WatcherHub) *WatchHandler {
	return &WatchHandler{
		configSvc: configSvc,
		hub:       hub,
	}
}

func (h *WatchHandler) Watch(c *gin.Context) {
	namespace := c.Param("namespace")
	group := c.Param("group")
	dataID := c.Param("dataId")
	clientMD5 := c.Query("md5")
	timeout, _ := strconv.Atoi(c.DefaultQuery("timeout", "30"))

	cfg, err := h.configSvc.GetConfig(c.Request.Context(), namespace, group, dataID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if cfg == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "config not found"})
		return
	}

	if cfg.ContentMD5 != clientMD5 {
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
		return
	}

	notifyCh := h.hub.Subscribe(namespace, group, dataID)
	defer h.hub.Unsubscribe(namespace, group, dataID, notifyCh)

	select {
	case <-notifyCh:
		cfg, _ = h.configSvc.GetConfig(c.Request.Context(), namespace, group, dataID)
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
	case <-time.After(time.Duration(timeout) * time.Second):
		c.Status(http.StatusNotModified)
	case <-c.Request.Context().Done():
	}
}

func (h *WatchHandler) BatchWatch(c *gin.Context) {
	var req struct {
		WatchItems []struct {
			Namespace string `json:"namespace"`
			Group     string `json:"group"`
			DataID    string `json:"dataId"`
			MD5       string `json:"md5"`
		} `json:"watchItems" binding:"required"`
		Timeout int `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if req.Timeout == 0 {
		req.Timeout = 30
	}

	changed := make([]map[string]string, 0)

	for _, item := range req.WatchItems {
		cfg, err := h.configSvc.GetConfig(c.Request.Context(), item.Namespace, item.Group, item.DataID)
		if err != nil || cfg == nil {
			continue
		}

		if cfg.ContentMD5 != item.MD5 {
			changed = append(changed, map[string]string{
				"namespace": item.Namespace,
				"group":     item.Group,
				"dataId":    item.DataID,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"changed": changed,
		},
	})
}