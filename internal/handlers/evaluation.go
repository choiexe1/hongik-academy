package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/choiexe1/hongik-academy/internal/db/sqlc"
)

type EvaluationHandler struct {
	queries *sqlc.Queries
}

func NewEvaluationHandler(queries *sqlc.Queries) *EvaluationHandler {
	return &EvaluationHandler{queries: queries}
}

func (h *EvaluationHandler) ListEvaluations(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")

	studentID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students")
		return
	}

	// 원생 정보 조회
	student, err := h.queries.GetStudentByID(c.Request.Context(), int32(studentID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "원생을 찾을 수 없습니다.",
		})
		return
	}

	// 페이지네이션 파라미터
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit != 10 && limit != 20 && limit != 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	// 검색/필터 파라미터
	search := c.Query("search")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// pgtype 변환
	var searchParam pgtype.Text
	if search != "" {
		searchParam = pgtype.Text{String: search, Valid: true}
	}

	var startDateParam pgtype.Date
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDateParam = pgtype.Date{Time: t, Valid: true}
		}
	}

	var endDateParam pgtype.Date
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDateParam = pgtype.Date{Time: t, Valid: true}
		}
	}

	// 총 개수 조회
	totalCount, err := h.queries.CountEvaluationsByStudent(c.Request.Context(), sqlc.CountEvaluationsByStudentParams{
		StudentID: int32(studentID),
		Search:    searchParam,
		StartDate: startDateParam,
		EndDate:   endDateParam,
	})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "평가표 개수를 불러오는데 실패했습니다.",
		})
		return
	}

	// 총 페이지 수 계산
	totalPages := int(totalCount) / limit
	if int(totalCount)%limit > 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}

	// 평가표 목록 조회
	evaluations, err := h.queries.ListEvaluationsByStudent(c.Request.Context(), sqlc.ListEvaluationsByStudentParams{
		StudentID: int32(studentID),
		Limit:     int32(limit),
		Offset:    int32(offset),
		Search:    searchParam,
		StartDate: startDateParam,
		EndDate:   endDateParam,
	})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "평가표 목록을 불러오는데 실패했습니다.",
		})
		return
	}

	c.HTML(http.StatusOK, "evaluations.html", gin.H{
		"student":     student,
		"evaluations": evaluations,
		"username":    username,
		"role":        role,
		"currentPage": "students",
		"page":        page,
		"limit":       limit,
		"totalCount":  totalCount,
		"totalPages":  totalPages,
		"search":      search,
		"startDate":   startDateStr,
		"endDate":     endDateStr,
	})
}

func (h *EvaluationHandler) ShowCreateForm(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")

	studentID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students")
		return
	}

	student, err := h.queries.GetStudentByID(c.Request.Context(), int32(studentID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "원생을 찾을 수 없습니다.",
		})
		return
	}

	c.HTML(http.StatusOK, "evaluation_form.html", gin.H{
		"title":       "평가표 작성",
		"action":      "/students/" + c.Param("id") + "/evaluations",
		"student":     student,
		"evaluation":  nil,
		"username":    username,
		"role":        role,
		"currentPage": "students",
	})
}

func (h *EvaluationHandler) CreateEvaluation(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id").(int32)
	username := session.Get("username")
	role := session.Get("role")

	studentID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students")
		return
	}

	student, err := h.queries.GetStudentByID(c.Request.Context(), int32(studentID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "원생을 찾을 수 없습니다.",
		})
		return
	}

	content := c.PostForm("content")
	if content == "" {
		c.HTML(http.StatusBadRequest, "evaluation_form.html", gin.H{
			"title":       "평가표 작성",
			"action":      "/students/" + c.Param("id") + "/evaluations",
			"student":     student,
			"error":       "평가 내용을 입력해주세요.",
			"username":    username,
			"role":        role,
			"currentPage": "students",
		})
		return
	}

	_, err = h.queries.CreateEvaluation(c.Request.Context(), sqlc.CreateEvaluationParams{
		StudentID: int32(studentID),
		AuthorID:  userID,
		Content:   content,
	})

	if err != nil {
		c.HTML(http.StatusInternalServerError, "evaluation_form.html", gin.H{
			"title":       "평가표 작성",
			"action":      "/students/" + c.Param("id") + "/evaluations",
			"student":     student,
			"error":       "평가표 저장에 실패했습니다.",
			"username":    username,
			"role":        role,
			"currentPage": "students",
		})
		return
	}

	c.Redirect(http.StatusFound, "/students/"+c.Param("id")+"/evaluations")
}

func (h *EvaluationHandler) ShowEditForm(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")

	studentID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students")
		return
	}

	evalID, err := strconv.ParseInt(c.Param("eval_id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students/"+c.Param("id")+"/evaluations")
		return
	}

	student, err := h.queries.GetStudentByID(c.Request.Context(), int32(studentID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "원생을 찾을 수 없습니다.",
		})
		return
	}

	evaluation, err := h.queries.GetEvaluationByID(c.Request.Context(), int32(evalID))
	if err != nil {
		c.Redirect(http.StatusFound, "/students/"+c.Param("id")+"/evaluations")
		return
	}

	c.HTML(http.StatusOK, "evaluation_form.html", gin.H{
		"title":       "평가표 수정",
		"action":      "/students/" + c.Param("id") + "/evaluations/" + c.Param("eval_id"),
		"student":     student,
		"evaluation":  evaluation,
		"username":    username,
		"role":        role,
		"currentPage": "students",
	})
}

func (h *EvaluationHandler) UpdateEvaluation(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")

	studentID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students")
		return
	}

	evalID, err := strconv.ParseInt(c.Param("eval_id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students/"+c.Param("id")+"/evaluations")
		return
	}

	student, err := h.queries.GetStudentByID(c.Request.Context(), int32(studentID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "원생을 찾을 수 없습니다.",
		})
		return
	}

	evaluation, err := h.queries.GetEvaluationByID(c.Request.Context(), int32(evalID))
	if err != nil {
		c.Redirect(http.StatusFound, "/students/"+c.Param("id")+"/evaluations")
		return
	}

	content := c.PostForm("content")
	if content == "" {
		c.HTML(http.StatusBadRequest, "evaluation_form.html", gin.H{
			"title":       "평가표 수정",
			"action":      "/students/" + c.Param("id") + "/evaluations/" + c.Param("eval_id"),
			"student":     student,
			"evaluation":  evaluation,
			"error":       "평가 내용을 입력해주세요.",
			"username":    username,
			"role":        role,
			"currentPage": "students",
		})
		return
	}

	_, err = h.queries.UpdateEvaluation(c.Request.Context(), sqlc.UpdateEvaluationParams{
		ID:      int32(evalID),
		Content: content,
	})

	if err != nil {
		c.HTML(http.StatusInternalServerError, "evaluation_form.html", gin.H{
			"title":       "평가표 수정",
			"action":      "/students/" + c.Param("id") + "/evaluations/" + c.Param("eval_id"),
			"student":     student,
			"evaluation":  evaluation,
			"error":       "평가표 수정에 실패했습니다.",
			"username":    username,
			"role":        role,
			"currentPage": "students",
		})
		return
	}

	c.Redirect(http.StatusFound, "/students/"+c.Param("id")+"/evaluations")
}

func (h *EvaluationHandler) DeleteEvaluation(c *gin.Context) {
	studentID := c.Param("id")

	evalID, err := strconv.ParseInt(c.Param("eval_id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students/"+studentID+"/evaluations")
		return
	}

	h.queries.DeleteEvaluation(c.Request.Context(), int32(evalID))

	c.Redirect(http.StatusFound, "/students/"+studentID+"/evaluations")
}
