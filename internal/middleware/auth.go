package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/choiexe1/hongik-academy/internal/db/sqlc"
)

func AuthRequired(queries *sqlc.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		sessionToken := session.Get("session_token")

		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 세션 토큰 검증
		if sessionToken != nil {
			id, ok := userID.(int32)
			if !ok {
				session.Clear()
				session.Save()
				c.Redirect(http.StatusFound, "/login")
				c.Abort()
				return
			}

			dbToken, err := queries.GetSessionToken(c.Request.Context(), id)
			if err != nil || !dbToken.Valid || dbToken.String != sessionToken.(string) {
				// 토큰이 일치하지 않음 = 다른 기기에서 로그인함
				session.Clear()
				session.Save()
				c.Redirect(http.StatusFound, "/login?reason=session_expired")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")

		if role != "super_admin" && role != "admin" {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"error": "관리자 권한이 필요합니다.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
