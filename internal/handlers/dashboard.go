package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"username":    username,
		"role":        role,
		"currentPage": "dashboard",
	})
}
