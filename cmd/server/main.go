package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/choiexe1/hongik-academy/internal/config"
	"github.com/choiexe1/hongik-academy/internal/db/sqlc"
	"github.com/choiexe1/hongik-academy/internal/handlers"
	"github.com/choiexe1/hongik-academy/internal/middleware"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	r := gin.Default()

	store := cookie.NewStore([]byte(cfg.SessionKey))
	store.Options(sessions.Options{
		MaxAge:   86400 * 3, // 3일
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("session", store))

	// 템플릿 함수 등록
	r.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"multiply": func(a, b int) int {
			return a * b
		},
		"iterate": func(count int) []int {
			result := make([]int, count)
			for i := range result {
				result[i] = i
			}
			return result
		},
		"isEven": func(i int) bool {
			return i%2 == 0
		},
		"int": func(i int64) int {
			return int(i)
		},
		"slice": func(s interface{}, start, end int) string {
			str, ok := s.(string)
			if !ok {
				return ""
			}
			runes := []rune(str)
			if start >= len(runes) {
				return ""
			}
			if end > len(runes) {
				end = len(runes)
			}
			return string(runes[start:end])
		},
	})

	r.LoadHTMLGlob("templates/*.html")

	authHandler := handlers.NewAuthHandler(queries)
	dashboardHandler := handlers.NewDashboardHandler()
	studentHandler := handlers.NewStudentHandler(queries)
	userHandler := handlers.NewUserHandler(queries)
	evaluationHandler := handlers.NewEvaluationHandler(queries)

	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/login")
	})
	r.GET("/login", authHandler.ShowLoginPage)
	r.POST("/login", authHandler.Login)
	r.GET("/logout", authHandler.Logout)

	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired(queries))
	{
		authorized.GET("/dashboard", dashboardHandler.ShowDashboard)

		// 원생 관리
		authorized.GET("/students", studentHandler.ListStudents)
		authorized.GET("/students/new", studentHandler.ShowCreateForm)
		authorized.POST("/students", studentHandler.CreateStudent)
		authorized.GET("/students/:id/edit", studentHandler.ShowEditForm)
		authorized.POST("/students/:id", studentHandler.UpdateStudent)
		authorized.POST("/students/:id/delete", studentHandler.DeleteStudent)

		// 평가표 관리
		authorized.GET("/students/:id/evaluations", evaluationHandler.ListEvaluations)
		authorized.GET("/students/:id/evaluations/new", evaluationHandler.ShowCreateForm)
		authorized.POST("/students/:id/evaluations", evaluationHandler.CreateEvaluation)
		authorized.GET("/students/:id/evaluations/:eval_id/edit", evaluationHandler.ShowEditForm)
		authorized.POST("/students/:id/evaluations/:eval_id", evaluationHandler.UpdateEvaluation)
		authorized.POST("/students/:id/evaluations/:eval_id/delete", evaluationHandler.DeleteEvaluation)
	}

	// 관리자 전용 (admin role 필요)
	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired(queries), middleware.AdminRequired())
	{
		// 사용자 관리
		admin.GET("/users", userHandler.ListUsers)
		admin.GET("/users/new", userHandler.ShowCreateForm)
		admin.POST("/users", userHandler.CreateUser)
		admin.GET("/users/:id/edit", userHandler.ShowEditForm)
		admin.POST("/users/:id", userHandler.UpdateUser)
		admin.POST("/users/:id/delete", userHandler.DeleteUser)
	}

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
