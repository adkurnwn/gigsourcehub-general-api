package http_member

import (
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/gin-gonic/gin"
)

// Get Profile
//
//	@Summary		Get Profile
//	@Description	Get detailed profile of current user
//	@Tags			Profile
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	gorm_model.UserResp
//	@Failure		400	{object}	response.Base
//	@Failure		401	{object}	response.Base
//	@Failure		404	{object}	response.Base
//	@Failure		500	{object}	response.Base
//	@Router			/profile [get]
//
//	@Security		BearerAuth
func (r *routeHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()

	// Call the usecase
	response := r.Usecase.GetProfile(ctx, c.MustGet("token_data").(domain.JWTClaimUser))
	c.JSON(response.Status, response)
}

// Upload Profile Picture
//
//	@Summary		Upload Profile Picture
//	@Description	Upload a profile picture for the current user
//	@Tags			Profile
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"Profile picture file (JPEG, PNG, WebP)"
//	@Success		200		{object}	response.Base
//	@Failure		400		{object}	response.Base
//	@Failure		401		{object}	response.Base
//	@Failure		500		{object}	response.Base
//	@Router			/profile/picture [post]
//
//	@Security		BearerAuth
func (r *routeHandler) UploadProfilePicture(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, map[string]any{"status": 400, "message": "file is required"})
		return
	}

	userClaim := c.MustGet("token_data").(domain.JWTClaimUser)
	resp := r.Usecase.UploadProfilePicture(c.Request.Context(), userClaim.UserID, file)
	c.JSON(resp.Status, resp)
}
