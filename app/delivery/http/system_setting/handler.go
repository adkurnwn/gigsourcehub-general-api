package system_setting

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/app/usecase/system_setting"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

type SystemSettingHandler struct {
	uc system_setting.Usecase
}

func NewSystemSettingHandler(api *gin.RouterGroup, mdl middleware.Middleware, uc system_setting.Usecase) {
	h := &SystemSettingHandler{uc: uc}

	group := api.Group("/system-settings")
	{
		// Public-ish: any authenticated user can check if AI is enabled
		group.GET("/ai-mode", mdl.Auth(), h.FetchAiMode)

		// Superadmin only
		group.GET("", mdl.Auth(), mdl.AuthSuperadmin(), h.Fetch)
		group.PUT("", mdl.Auth(), mdl.AuthSuperadmin(), h.Update)
	}
}

func (h *SystemSettingHandler) FetchAiMode(c *gin.Context) {
	data, err := h.uc.Fetch(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]bool{
		"is_ai_mode_enabled": data.IsAIModeEnabled,
	}))
}

func (h *SystemSettingHandler) Fetch(c *gin.Context) {
	data, err := h.uc.Fetch(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(data))
}

func (h *SystemSettingHandler) Update(c *gin.Context) {
	var req request_model.UpdateSystemSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}

	err := h.uc.Update(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
