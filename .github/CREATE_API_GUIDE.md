# Guide: Creating a New API Endpoint in Clean Architecture

This document maps out the specific steps you should follow whenever you want to add a new API endpoint to the Go application. The system follows strict **Clean Architecture**, meaning data flows from the **Database Mappings** -> **Repository** -> **Usecase** -> **Delivery Handlers**.

---

## Step 1: Update the Database Schema & Migrations
First, decide what data you need to store. We use Atlas for database migrations based on HCL and pure SQL.
1. Check or write your new migration specifications in the local schema mapping.
2. If following the `Makefile`, generate a migration delta by running:
   ```bash
   make migration name="create_new_table"
   ```
   This will generate a `.sql` file tracking your schema changes inside `database/migrations/`.
3. Apply the migration using:
   ```bash
   make migrate-up
   ```

## Step 2: Create the Domain Models
Define how your data is represented structurally.
1. **GORM Model**: Create your mapping struct in `domain/model/gorm/your_model.go`.
   - Add standard metadata fields (`ID`, `CreatedAt`, `UpdatedAt`).
   - Define your generic output JSON structures (e.g., `type YourModelResp struct`) and a `.ToResp()` convertor attached to the original GORM model to strip away sensitive metadata.
2. **Request Model**: If your API accepts a custom JSON body, create a validation struct in `domain/model/request/your_request.go`.

## Step 3: Define & Implement the Repository Layer
The repository layer securely queries the database. Keep business logic completely out of this layer.
1. Add the function signature strictly inside the exact interface blocks in `domain/repository.go`.
   ```go
   FetchYourModel(ctx context.Context, id string) (*gorm_model.YourModel, error)
   ```
3. Write the actual implementation using GORM mapping functions inside your target database file (e.g., `app/repository/gorm/your_model.go`).
   - Add `Preload("Relation")` calls if your `ToResp()` mapping relies on joined data strings securely.

## Step 4: Regenerate the Database Mocks
Because the Usecase tier is strictly isolated from the real database for Unit Testing, you must update the generic mock struct whenever you add a new function to the `domain.GormRepo` interface!
Run this command from your terminal to regenerate `mocks/GormRepo.go`:
```bash
# If using the Windows PowerShell directly: 
mockery --all

# Or if using WSL / bash:
make generate-mocks
```

## Step 5: Define & Implement the Usecase Layer
The usecase layer holds all the business rules. It connects API requests to database actions safely.
1. Explicitly add your method signature to the root module interface in `domain/usecase.go` (e.g., `YourAppUsecase interface { ... }`).
   ```go
   FetchData(ctx context.Context, id string) response.Base
   ```
2. Implement the usecase logic inside the module (e.g., `app/usecase/your_module/your_logic.go`).
   - Use `context.WithTimeout` bindings to prevent long database hangs.
   - For database counting/pagination, calculate totals *before* adding your `.Limit()` filters.
   - Evaluate your data structs, generate S3 URLs or logic here, and bundle your outputs in `response.Success()` or `response.Error()`.
   
   **Sample code for a GET request:**
   ```go
   package usecase_your_module

   import (
       "context"
       "net/http"

       "github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
       // import your gorm model path here
   )

   func (u *appUsecase) FetchData(ctx context.Context, id string) response.Base {
       // 1. Create a timeout context safely
       ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
       defer cancel()

       // 2. Query the repository
       data, err := u.gormDbRepo.FetchYourModel(ctx, id)
       if err != nil {
           // Keep error messages clean for the frontend
           return response.Error(http.StatusInternalServerError, "Failed to fetch data") 
       }
       if data == nil {
           return response.Error(http.StatusNotFound, "Data not found")
       }

       // 3. (Optional) Perform business logic, mappings, or presigning S3 URLs

       // 4. Return formatted JSON success wrapper
       return response.Success(data.ToResp())
   }
   ```

## Step 6: Implement the HTTP Delivery Handlers
This stage exposes your business logic structurally behind REST endpoints.
1. Write the `gin.Context` parser inside your handler file, e.g., `app/delivery/http/your_module/your_handler.go`.
   - Use `c.Param("id")` for fetching path parameters (`/users/:id`).
   - Use `c.ShouldBindJSON(&req)` for request bodies.
   - Use `helpers.GetPagination(c)` to safely extract automated `page`, `limit`, and `cursor` properties from the `GET` search queries!
   **Sample Handler Setup & Usecase Injection:**
   ```go
   package http_your_module

   import (
       "github.com/gin-gonic/gin"
       "github.com/adkurnwn/gigsourcehub-general-api/domain"
       "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
   )

   // 1. Define your Handler struct that holds the Usecase interface
   type routeHandler struct {
       Usecase    domain.YourAppUsecase
       Route      *gin.RouterGroup
       Middleware middleware.Middleware
   }

   // 2. Create a constructor to inject the Usecase and register the routes
   func NewYourModuleHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.YourAppUsecase) {
       handler := &routeHandler{
           Usecase:    uc,
           Route:      r,
           Middleware: mdl,
       }

       // 3. Register your routes here (often moved to a handleYourRoute map)
       api := r.Group("/your_module")
       api.GET("/:id", mdl.Auth(), mdl.AuthRole("Admin"), handler.FetchData)
   }

   // 4. Implement your Gin Handler logic
   func (h *routeHandler) FetchData(c *gin.Context) {
       id := c.Param("id")
       res := h.Usecase.FetchData(c.Request.Context(), id)
       c.JSON(res.Status, res)
   }
   ```

## Step 7: Register the Secured Route
Inside your HTTP delivery's configuration file (e.g., `app/delivery/http/your_module/init.go`), bind the function to an exact route and protect it with Middlewares:
```go
func (h *routeHandler) handleYourRoute(path string) {
    api := h.Route.Group(path)

    // Example of attaching Role Middleware protection:
    api.GET("/:id", h.Middleware.Auth(), h.Middleware.AuthRole("Admin"), h.FetchData)
}
```

## Step 8: Add Swagger Annotations
Directly above your `FetchData` handler function in Step 5, write the standard Swagger comments. This is exactly what the compiler looks for to generate the testing UI.

```go
// FetchData
// @Summary Fetch My Data
// @Description Fetch specific details about the data struct
// @Tags YourModule
// @Accept json
// @Produce json
// @Param id path string true "Data ID"
// @Success 200 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /your_module/{id} [get]
// @Security BearerAuth
func (h *routeHandler) FetchData(c *gin.Context) { ... }
```

Finally, generate the Swagger API UI by running the generator hook:
```bash
make run-swagger
# Or manually format them:
swag fmt
swag init -g ./main.go -o ./docs
```

> **Congratulations!** Your data will now perfectly map from an Atlas migration -> Gorm schema -> Usecase evaluation -> JSON Response -> Interactive Swagger Docs!

---

## Step 9: Run the Test Suite
Now that your API endpoints are mapped and your mock repository is generated, you can securely execute unit tests isolated from the main database instance:

```bash
# Run all tests natively in go
go test -v ./...

# Or if you want a visual coverage metric printed:
make testcoverage
```
