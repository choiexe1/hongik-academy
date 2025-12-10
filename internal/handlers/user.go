package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/choiexe1/hongik-academy/internal/db/sqlc"
)

type UserHandler struct {
	queries *sqlc.Queries
}

func NewUserHandler(queries *sqlc.Queries) *UserHandler {
	return &UserHandler{queries: queries}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	session := sessions.Default(c)
	currentUserID := session.Get("user_id").(int32)
	currentRole := session.Get("role").(string)

	users, err := h.queries.ListUsers(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "사용자 목록을 불러오는데 실패했습니다.",
		})
		return
	}

	username := session.Get("username")

	c.HTML(http.StatusOK, "users.html", gin.H{
		"users":         users,
		"currentUserID": currentUserID,
		"currentRole":   currentRole,
		"username":      username,
		"role":          currentRole,
		"currentPage":   "users",
	})
}

func (h *UserHandler) ShowCreateForm(c *gin.Context) {
	session := sessions.Default(c)
	currentRole := session.Get("role").(string)
	username := session.Get("username")

	// 최고관리자만 관리자 등록 가능
	if currentRole != "super_admin" {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error": "관리자를 등록할 권한이 없습니다.",
		})
		return
	}

	c.HTML(http.StatusOK, "user_form.html", gin.H{
		"title":       "관리자 등록",
		"action":      "/admin/users",
		"user":        nil,
		"currentRole": currentRole,
		"username":    username,
		"role":        currentRole,
		"currentPage": "users",
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	session := sessions.Default(c)
	currentRole := session.Get("role").(string)

	// 최고관리자만 관리자 등록 가능
	if currentRole != "super_admin" {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error": "관리자를 등록할 권한이 없습니다.",
		})
		return
	}

	username := c.PostForm("username")
	name := c.PostForm("name")
	password := c.PostForm("password")
	role := c.PostForm("role")

	if username == "" || name == "" || password == "" {
		c.HTML(http.StatusBadRequest, "user_form.html", gin.H{
			"title":       "관리자 등록",
			"action":      "/admin/users",
			"error":       "아이디, 이름, 비밀번호는 필수입니다.",
			"currentRole": currentRole,
		})
		return
	}

	// 최고관리자만 super_admin 역할 부여 가능
	if role == "super_admin" && currentRole != "super_admin" {
		role = "admin"
	}
	if role != "super_admin" && role != "admin" {
		role = "admin"
	}

	// 비밀번호 해시
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "user_form.html", gin.H{
			"title":  "관리자 등록",
			"action": "/admin/users",
			"error":  "비밀번호 처리 중 오류가 발생했습니다.",
		})
		return
	}

	_, err = h.queries.CreateUser(c.Request.Context(), sqlc.CreateUserParams{
		Username:     username,
		Name:         name,
		PasswordHash: string(hashedPassword),
		Role:         role,
	})

	if err != nil {
		c.HTML(http.StatusInternalServerError, "user_form.html", gin.H{
			"title":  "관리자 등록",
			"action": "/admin/users",
			"error":  "사용자 등록에 실패했습니다. 아이디가 중복되었을 수 있습니다.",
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

func (h *UserHandler) ShowEditForm(c *gin.Context) {
	session := sessions.Default(c)
	currentUserID := session.Get("user_id").(int32)
	currentRole := session.Get("role").(string)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// 일반 관리자는 자기 자신만 수정 가능
	if currentRole != "super_admin" && currentUserID != int32(id) {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error": "다른 관리자를 수정할 권한이 없습니다.",
		})
		return
	}

	user, err := h.queries.GetUserByID(c.Request.Context(), int32(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// 자기 자신인지 확인 (역할 수정 방지용)
	isSelf := currentUserID == int32(id)

	username := session.Get("username")

	c.HTML(http.StatusOK, "user_form.html", gin.H{
		"title":       "관리자 수정",
		"action":      "/admin/users/" + c.Param("id"),
		"user":        user,
		"currentRole": currentRole,
		"isSelf":      isSelf,
		"username":    username,
		"role":        currentRole,
		"currentPage": "users",
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	session := sessions.Default(c)
	currentUserID := session.Get("user_id").(int32)
	currentRole := session.Get("role").(string)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// 일반 관리자는 자기 자신만 수정 가능
	if currentRole != "super_admin" && currentUserID != int32(id) {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error": "다른 관리자를 수정할 권한이 없습니다.",
		})
		return
	}

	// 대상 사용자 조회
	targetUser, err := h.queries.GetUserByID(c.Request.Context(), int32(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	name := c.PostForm("name")
	password := c.PostForm("password")
	role := c.PostForm("role")
	isSelf := currentUserID == int32(id)

	if name == "" {
		c.HTML(http.StatusBadRequest, "user_form.html", gin.H{
			"title":       "관리자 수정",
			"action":      "/admin/users/" + c.Param("id"),
			"error":       "이름은 필수입니다.",
			"user":        targetUser,
			"currentRole": currentRole,
			"isSelf":      isSelf,
		})
		return
	}

	// 최고관리자의 역할은 변경 불가
	if targetUser.Role == "super_admin" {
		role = "super_admin"
	} else if isSelf {
		// 자기 자신의 역할은 변경 불가 - 기존 역할 유지
		role = targetUser.Role
	} else {
		// 최고관리자만 super_admin 역할 부여 가능
		if role == "super_admin" && currentRole != "super_admin" {
			role = "admin"
		}
		if role != "super_admin" && role != "admin" {
			role = "admin"
		}
	}

	// 사용자 정보 업데이트
	_, err = h.queries.UpdateUser(c.Request.Context(), sqlc.UpdateUserParams{
		ID:   int32(id),
		Name: name,
		Role: role,
	})

	if err != nil {
		c.HTML(http.StatusInternalServerError, "user_form.html", gin.H{
			"title":       "관리자 수정",
			"action":      "/admin/users/" + c.Param("id"),
			"error":       "사용자 수정에 실패했습니다.",
			"user":        targetUser,
			"currentRole": currentRole,
			"isSelf":      isSelf,
		})
		return
	}

	// 비밀번호가 입력된 경우에만 업데이트
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err == nil {
			h.queries.UpdateUserPassword(c.Request.Context(), sqlc.UpdateUserPasswordParams{
				ID:           int32(id),
				PasswordHash: string(hashedPassword),
			})
		}
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	session := sessions.Default(c)
	currentUserID := session.Get("user_id").(int32)
	currentRole := session.Get("role").(string)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// 자기 자신은 삭제 불가
	if currentUserID == int32(id) {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error": "자기 자신은 삭제할 수 없습니다.",
		})
		return
	}

	// 최고관리자만 삭제 가능
	if currentRole != "super_admin" {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error": "최고관리자만 관리자를 삭제할 수 있습니다.",
		})
		return
	}

	h.queries.DeleteUser(c.Request.Context(), int32(id))

	c.Redirect(http.StatusFound, "/admin/users")
}
