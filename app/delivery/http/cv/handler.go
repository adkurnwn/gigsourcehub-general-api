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
}

//	CV Upload
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
