# Guide to Creating POST, PUT, and DELETE APIs

This guide details the standard architectural procedure for scaffolding create (`POST`), update (`PUT`), and delete (`DELETE`) endpoints in this codebase. Our application adheres to a strict interface-driven Domain, Repository, Usecase, and HTTP Handler layer segregation.

## 1. Domain Request Models
Define the expected JSON payloads for your endpoints inside `domain/model/request`. Create dedicated structs so the Gin framework can bind and validate them natively using `binding:"required"` tags.

```go
// Example: domain/model/request/feature.go
package request_model

type CreateFeatureRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID string `json:"parent_id" binding:"required"`
}

type UpdateFeatureRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID string `json:"parent_id" binding:"required"`
}
```

*Note: Deletion operations conventionally rely on URL path parameters (e.g., `/:id`) rather than JSON request bodies, so a dedicated request model for DELETE is typically unneeded.*

## 2. Update the Repository Interfaces & Logic
The generic `GormRepo` interface in `domain/repository.go` must declare the database functions for inserting, modifying, and destroying entities.

```go
// domain/repository.go
type GormRepo interface {
	// ... existing methods
	CreateFeature(ctx context.Context, model *gorm_model.Feature) error
	UpdateFeature(ctx context.Context, model *gorm_model.Feature) error
	DeleteFeature(ctx context.Context, id string) error
}
```

Then, implement the underlying GORM instructions within the associated package in `app/repository/gorm/{feature}.go`.

```go
func (r *gormRepo) CreateFeature(ctx context.Context, model *gorm_model.Feature) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) UpdateFeature(ctx context.Context, model *gorm_model.Feature) error {
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormRepo) DeleteFeature(ctx context.Context, id string) error {
	// Assumes standard Gorm soft-delete via DeletedAt indexing
	return r.db.WithContext(ctx).Delete(&gorm_model.Feature{}, "id = ?", id).Error
}
```

## 3. Update the Usecase Interface
Add the corresponding method signatures to the feature's dedicated usecase interface inside `domain/usecase.go`.

```go
type FeatureAppUsecase interface {
	Create(ctx context.Context, req request_model.CreateFeatureRequest) response.Base
	Update(ctx context.Context, id string, req request_model.UpdateFeatureRequest) response.Base
	Delete(ctx context.Context, id string) response.Base
    // ... Any existing Fetch methodologies ...
}
```

## 4. Implement the Usecase Logic
In `app/usecase/{feature}/{feature}.go`, apply the central business logic. This layer governs payload mapping to the `gorm_model`, orchestrating validation, dispatching database calls via the Repository layer, and formatting the safe return `response.Base`.

```go
func (u *appUsecase) Create(ctx context.Context, req request_model.CreateFeatureRequest) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	newFeature := gorm_model.Feature{
		ID:       uuid.New().String(),
		Name:     req.Name,
		ParentID: req.ParentID,
	}
	
	if err := u.gormDbRepo.CreateFeature(ctx, &newFeature); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to create feature")
	}

	return response.Success(newFeature.ToFeatureResp())
}

func (u *appUsecase) Delete(ctx context.Context, id string) response.Base {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.gormDbRepo.DeleteFeature(ctx, id); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to delete feature")
	}

	return response.Success(nil)
}
```

## 5. Expose the HTTP Delivery Handlers
In `app/delivery/http/{feature}/handler.go`, create the final ingress handlers. These functions catch the HTTP request context `*gin.Context`, execute payload binding via `ShouldBindJSON`, trigger the respective Usecase method, and serve the final response JSON. 

Ensure they are systematically wired into the router configuration method mapping to the proper HTTP verb targets (`api.POST`, `api.PUT`, `api.DELETE`).

```go
func NewFeatureHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.FeatureAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/features", mdl.Auth(), mdl.AuthSuperadmin()) // Protect mutations generally!
	api.POST("", handler.Create)
	api.PUT("/:id", handler.Update)
	api.DELETE("/:id", handler.Delete)
}

// Create Feature
// @Security BearerAuth
// @Summary Create Feature
// @Tags Feature
// @Accept json
// @Produce json
// @Param request body request_model.CreateFeatureRequest true "Create Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Router /features [post]
func (h *routeHandler) Create(c *gin.Context) {
	var req request_model.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(400, "invalid payload"))
		return
	}

	res := h.Usecase.Create(c.Request.Context(), req)
	c.JSON(res.Status, res)
}

// Delete Feature By ID
// @Security BearerAuth
// @Summary Delete Feature By ID
// @Tags Feature
// @Accept json
// @Produce json
// @Param id path string true "Feature ID"
// @Success 200 {object} response.Base
// @Router /features/{id} [delete]
func (h *routeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	res := h.Usecase.Delete(c.Request.Context(), id)
	c.JSON(res.Status, res)
}
```

## 6. Regenerate Swagger Annotations
Always conclude your endpoint modifications by running the Swag CLI from the project root. This ensures the frontend teams and `docs/swagger.yaml` schema stay accurately synchronized with your DTO and route updates.
```bash
swag init
```
