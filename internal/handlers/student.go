package handlers

import (
	"math"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/choiexe1/hongik-academy/internal/db/sqlc"
)

var nonDigitRegex = regexp.MustCompile(`[^0-9]`)

// sanitizePhone removes all non-numeric characters from phone number
func sanitizePhone(phone string) string {
	return nonDigitRegex.ReplaceAllString(phone, "")
}

type StudentHandler struct {
	queries *sqlc.Queries
}

func NewStudentHandler(queries *sqlc.Queries) *StudentHandler {
	return &StudentHandler{queries: queries}
}

func (h *StudentHandler) ListStudents(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")

	// 검색어
	search := c.Query("search")
	// 성별 필터
	gender := c.Query("gender")
	// 페이지 (기본값 1)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	// 페이지당 항목 수 (기본값 10)
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if perPage != 10 && perPage != 20 && perPage != 50 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	// 필터 파라미터 설정
	searchParam := pgtype.Text{Valid: false}
	if search != "" {
		searchParam = pgtype.Text{String: search, Valid: true}
	}
	genderParam := pgtype.Text{Valid: false}
	if gender != "" && (gender == "M" || gender == "F") {
		genderParam = pgtype.Text{String: gender, Valid: true}
	}

	// 전체 개수 조회
	totalCount, err := h.queries.CountStudents(c.Request.Context(), sqlc.CountStudentsParams{
		Search: searchParam,
		Gender: genderParam,
	})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "원생 수를 조회하는데 실패했습니다.",
		})
		return
	}

	// 학생 목록 조회
	students, err := h.queries.ListStudents(c.Request.Context(), sqlc.ListStudentsParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
		Search: searchParam,
		Gender: genderParam,
	})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "원생 목록을 불러오는데 실패했습니다.",
		})
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))

	c.HTML(http.StatusOK, "students.html", gin.H{
		"students":    students,
		"search":      search,
		"gender":      gender,
		"page":        page,
		"perPage":     perPage,
		"totalCount":  totalCount,
		"totalPages":  totalPages,
		"username":    username,
		"role":        role,
		"currentPage": "students",
	})
}

func (h *StudentHandler) ShowCreateForm(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")

	c.HTML(http.StatusOK, "student_form.html", gin.H{
		"title":       "원생 등록",
		"action":      "/students",
		"student":     nil,
		"username":    username,
		"role":        role,
		"currentPage": "students",
	})
}

func (h *StudentHandler) CreateStudent(c *gin.Context) {
	name := c.PostForm("name")
	gender := c.PostForm("gender")
	phone := sanitizePhone(c.PostForm("phone"))
	parentPhone := sanitizePhone(c.PostForm("parent_phone"))
	remarks := c.PostForm("remarks")

	if name == "" {
		c.HTML(http.StatusBadRequest, "student_form.html", gin.H{
			"title":  "원생 등록",
			"action": "/students",
			"error":  "이름을 입력해주세요.",
		})
		return
	}

	if gender != "M" && gender != "F" {
		c.HTML(http.StatusBadRequest, "student_form.html", gin.H{
			"title":  "원생 등록",
			"action": "/students",
			"error":  "성별을 선택해주세요.",
		})
		return
	}

	_, err := h.queries.CreateStudent(c.Request.Context(), sqlc.CreateStudentParams{
		Name:        name,
		Gender:      gender,
		Phone:       pgtype.Text{String: phone, Valid: phone != ""},
		ParentPhone: pgtype.Text{String: parentPhone, Valid: parentPhone != ""},
		Remarks:     pgtype.Text{String: remarks, Valid: remarks != ""},
	})

	if err != nil {
		c.HTML(http.StatusInternalServerError, "student_form.html", gin.H{
			"title":  "원생 등록",
			"action": "/students",
			"error":  "원생 등록에 실패했습니다.",
		})
		return
	}

	c.Redirect(http.StatusFound, "/students")
}

func (h *StudentHandler) ShowEditForm(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students")
		return
	}

	student, err := h.queries.GetStudentByID(c.Request.Context(), int32(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/students")
		return
	}

	c.HTML(http.StatusOK, "student_form.html", gin.H{
		"title":       "원생 수정",
		"action":      "/students/" + c.Param("id"),
		"student":     student,
		"username":    username,
		"role":        role,
		"currentPage": "students",
	})
}

func (h *StudentHandler) UpdateStudent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students")
		return
	}

	name := c.PostForm("name")
	gender := c.PostForm("gender")
	phone := sanitizePhone(c.PostForm("phone"))
	parentPhone := sanitizePhone(c.PostForm("parent_phone"))
	remarks := c.PostForm("remarks")

	if name == "" {
		c.HTML(http.StatusBadRequest, "student_form.html", gin.H{
			"title":  "원생 수정",
			"action": "/students/" + c.Param("id"),
			"error":  "이름을 입력해주세요.",
		})
		return
	}

	if gender != "M" && gender != "F" {
		c.HTML(http.StatusBadRequest, "student_form.html", gin.H{
			"title":  "원생 수정",
			"action": "/students/" + c.Param("id"),
			"error":  "성별을 선택해주세요.",
		})
		return
	}

	_, err = h.queries.UpdateStudent(c.Request.Context(), sqlc.UpdateStudentParams{
		ID:          int32(id),
		Name:        name,
		Gender:      gender,
		Phone:       pgtype.Text{String: phone, Valid: phone != ""},
		ParentPhone: pgtype.Text{String: parentPhone, Valid: parentPhone != ""},
		Remarks:     pgtype.Text{String: remarks, Valid: remarks != ""},
	})

	if err != nil {
		c.HTML(http.StatusInternalServerError, "student_form.html", gin.H{
			"title":  "원생 수정",
			"action": "/students/" + c.Param("id"),
			"error":  "원생 수정에 실패했습니다.",
		})
		return
	}

	c.Redirect(http.StatusFound, "/students")
}

func (h *StudentHandler) DeleteStudent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusFound, "/students")
		return
	}

	h.queries.DeleteStudent(c.Request.Context(), int32(id))

	c.Redirect(http.StatusFound, "/students")
}
