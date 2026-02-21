package http_cv

import (
	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	usecase_cv "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/cv"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

type CVHandler struct {
	Usecase usecase_cv.CVUsecase
}

func NewCVHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc usecase_cv.CVUsecase) {
	handler := &CVHandler{Usecase: uc}

	api := r.Group("/cv")
	api.POST("/upload", mdl.Auth(), handler.Upload) // <--- Protected Route
	api.GET("/:id/parsed", mdl.Auth(), handler.GetParsedCV)
	api.POST("/:id/confirm", mdl.Auth(), handler.ConfirmCV)
}

//	CV Upload
//
// @Summary Upload CV
// @Description Upload CV
// @Tags CV
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CV file"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /cv/upload [post]
//
//	@Security		BearerAuth
func (h *CVHandler) Upload(c *gin.Context) {
	// 1. Get File from Request
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, response.Error(400, "file is required"))
		return
	}

	// 2. Get Current User (from Middleware)
	userClaim := c.MustGet("token_data").(domain.JWTClaimUser)

	// 3. Call Usecase
	resp := h.Usecase.UploadCV(c.Request.Context(), userClaim.UserID, file)
	c.JSON(resp.Status, resp)
}

//	Get Parsed CV
//
// @Summary Get parsed CV data
// @Description Fetch the parsed CV data by CV ID
// @Tags CV
// @Produce json
// @Param id path string true "CV ID"
// @Success 200 {object} response.Base
// @Failure 404 {object} response.Base
// @Router /cv/{id}/parsed [get]
//
//	@Security		BearerAuth
func (h *CVHandler) GetParsedCV(c *gin.Context) {
	id := c.Param("id")
	resp := h.Usecase.GetParsedCV(c.Request.Context(), id)
	c.JSON(resp.Status, resp)
}

type ConfirmCVRequest struct {
	EditedData map[string]interface{} `json:"edited_data" binding:"required"`
}

//	Confirm CV
//
// @Summary Confirm parsed CV data
// @Description Confirm the parsed CV data after editing
// @Tags CV
// @Accept json
// @Produce json
// @Param id path string true "CV ID"
// @Param body body ConfirmCVRequest true "Edited CV data"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Router /cv/{id}/confirm [post]
//
//	@Security		BearerAuth
func (h *CVHandler) ConfirmCV(c *gin.Context) {
	id := c.Param("id")
	var req ConfirmCVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid request body"))
		return
	}

	userClaim := c.MustGet("token_data").(domain.JWTClaimUser)
	resp := h.Usecase.ConfirmCV(c.Request.Context(), userClaim.UserID, id, req.EditedData)
	c.JSON(resp.Status, resp)
}
