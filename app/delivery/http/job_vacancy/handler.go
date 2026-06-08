package http_job_vacancy

import (
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.JobVacancyAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewJobVacancyHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.JobVacancyAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	// CMS routes — require auth + Admin or Superadmin
	cms := r.Group("/job-vacancies", mdl.Auth(), mdl.AuthRole("Admin", "Superadmin"))
	cms.GET("", handler.FetchAll)
	cms.GET("/:id", handler.FetchData)
	cms.POST("", handler.Create)
	cms.PUT("/:id", handler.Update)
	cms.DELETE("/:id", handler.Delete)
	cms.PATCH("/:id/archive", handler.Archive)

	// Public routes — no auth required
	pub := r.Group("/public/job-vacancies")
	pub.GET("", handler.FetchPublic)
	pub.GET("/:id", handler.FetchPublicByID)
}

// Create Job Vacancy
// @Security BearerAuth
// @Summary Create Job Vacancy
// @Tags Job Vacancy
// @Accept json
// @Produce json
// @Param request body request_model.CreateJobVacancyRequest true "Create Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-vacancies [post]
func (h *routeHandler) Create(c *gin.Context) {
	var req request_model.CreateJobVacancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), req)
	c.JSON(res.Status, res)
}

// Update Job Vacancy
// @Security BearerAuth
// @Summary Update Job Vacancy
// @Tags Job Vacancy
// @Accept json
// @Produce json
// @Param id path string true "Job Vacancy ID"
// @Param request body request_model.UpdateJobVacancyRequest true "Update Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-vacancies/{id} [put]
func (h *routeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request_model.UpdateJobVacancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Update(c.Request.Context(), id, req)
	c.JSON(res.Status, res)
}

// Delete Job Vacancy
// @Security BearerAuth
// @Summary Delete Job Vacancy
// @Tags Job Vacancy
// @Produce json
// @Param id path string true "Job Vacancy ID"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-vacancies/{id} [delete]
func (h *routeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.Delete(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get All Job Vacancies (CMS)
// @Security BearerAuth
// @Summary Get All Job Vacancies
// @Description Get all job vacancies with pagination (Admin & Superadmin)
// @Tags Job Vacancy
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param subrequest_id query string false "Filter by Subrequest ID"
// @Param status query string false "Filter by status (DRAFT|ARCHIVED|PUBLISHED)"
// @Param schema query string false "Filter by schema (ONSITE|REMOTE|HYBRID)"
// @Param search query string false "Search by name"
// @Param published_at_from query string false "Filter published_at from (YYYY-MM-DD)"
// @Param published_at_to query string false "Filter published_at to (YYYY-MM-DD)"
// @Param sort query string false "Sort order: asc or desc (default: desc)"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-vacancies [get]
func (h *routeHandler) FetchAll(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.JobVacancyFilter{}

	if v := c.Query("subrequest_id"); v != "" {
		filter.SubrequestID = &v
	}
	if v := c.Query("status"); v != "" {
		filter.Status = &v
	}
	if v := c.Query("schema"); v != "" {
		filter.Schema = &v
	}
	if v := c.Query("search"); v != "" {
		filter.Search = &v
	}

	// Filter published_at range (menggunakan created_at sebagai proxy)
	if v := c.Query("published_at_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.PublishedAtFrom = &t
		}
	}
	if v := c.Query("published_at_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			// Set to end of day
			end := t.Add(24*time.Hour - time.Second)
			filter.PublishedAtTo = &end
		}
	}

	// Sort order
	sortOrder := "DESC"
	if v := c.Query("sort"); v == "asc" {
		sortOrder = "ASC"
	}
	filter.Sorts = []map[string]string{{"created_at": sortOrder}}

	res := h.Usecase.FetchAll(c.Request.Context(), pagination.Page, pagination.Limit, filter)
	c.JSON(res.Status, res)
}

// Get Job Vacancy By ID (CMS)
// @Security BearerAuth
// @Summary Get Job Vacancy By ID
// @Description Get job vacancy detail by ID (Admin & Superadmin)
// @Tags Job Vacancy
// @Accept json
// @Produce json
// @Param id path string true "Job Vacancy ID"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-vacancies/{id} [get]
func (h *routeHandler) FetchData(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchData(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Archive Job Vacancy (CMS)
// @Security BearerAuth
// @Summary Archive Job Vacancy
// @Description Set job vacancy status to ARCHIVED (Admin & Superadmin)
// @Tags Job Vacancy
// @Produce json
// @Param id path string true "Job Vacancy ID"
// @Success 200 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /job-vacancies/{id}/archive [patch]
func (h *routeHandler) Archive(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.Archive(c.Request.Context(), id)
	c.JSON(res.Status, res)
}

// Get Public Job Vacancies (no auth)
// @Summary Get Public Job Vacancies
// @Description Get all published and active job vacancies (no authentication required)
// @Tags Job Vacancy Public
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param schema query string false "Filter by schema (ONSITE|REMOTE|HYBRID)"
// @Param search query string false "Search by name"
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /public/job-vacancies [get]
func (h *routeHandler) FetchPublic(c *gin.Context) {
	pagination := helpers.GetPagination(c)
	filter := gorm_model.JobVacancyFilter{}

	if v := c.Query("schema"); v != "" {
		filter.Schema = &v
	}
	if v := c.Query("search"); v != "" {
		filter.Search = &v
	}

	res := h.Usecase.FetchPublic(c.Request.Context(), pagination.Page, pagination.Limit, filter)
	c.JSON(res.Status, res)
}

// Get Public Job Vacancy By ID (no auth)
// @Summary Get Public Job Vacancy By ID
// @Description Get a published and active job vacancy by ID (no authentication required)
// @Tags Job Vacancy Public
// @Accept json
// @Produce json
// @Param id path string true "Job Vacancy ID"
// @Success 200 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /public/job-vacancies/{id} [get]
func (h *routeHandler) FetchPublicByID(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.FetchPublicByID(c.Request.Context(), id)
	c.JSON(res.Status, res)
}
