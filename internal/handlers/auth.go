package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/choiexe1/hongik-academy/internal/db/sqlc"
)

type AuthHandler struct {
	queries *sqlc.Queries
}

func NewAuthHandler(queries *sqlc.Queries) *AuthHandler {
	return &AuthHandler{queries: queries}
}

func (h *AuthHandler) ShowLoginPage(c *gin.Context) {
	// 이미 로그인된 경우 대시보드로 리다이렉트
	session := sessions.Default(c)
	if userID := session.Get("user_id"); userID != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	// 로그아웃 사유 확인
	reason := c.Query("reason")
	errorMsg := ""
	if reason == "session_expired" {
		errorMsg = "다른 기기에서 로그인하여 현재 세션이 종료되었습니다."
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"error": errorMsg,
	})
}

type LoginRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// generateSessionToken generates a random session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "사용자명과 비밀번호를 입력해주세요.",
		})
		return
	}

	user, err := h.queries.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "사용자명 또는 비밀번호가 올바르지 않습니다.",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "사용자명 또는 비밀번호가 올바르지 않습니다.",
		})
		return
	}

	// 세션 토큰 생성
	sessionToken, err := generateSessionToken()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"error": "로그인 처리 중 오류가 발생했습니다.",
		})
		return
	}

	// DB에 세션 토큰 저장 (기존 토큰 덮어쓰기 = 다른 기기 로그아웃)
	h.queries.UpdateSessionToken(c.Request.Context(), sqlc.UpdateSessionTokenParams{
		ID:           user.ID,
		SessionToken: pgtype.Text{String: sessionToken, Valid: true},
	})

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("session_token", sessionToken)
	session.Save()

	c.Redirect(http.StatusFound, "/dashboard")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)

	// DB에서 세션 토큰 제거
	if userID := session.Get("user_id"); userID != nil {
		if id, ok := userID.(int32); ok {
			h.queries.ClearSessionToken(c.Request.Context(), id)
		}
	}

	session.Clear()
	session.Save()

	c.Redirect(http.StatusFound, "/login")
}
