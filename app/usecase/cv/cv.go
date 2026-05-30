package usecase_cv

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var cvProgressMap sync.Map // Map[string]int (cvID -> progress)

func SetCVProgress(cvID string, progress int) {
	cvProgressMap.Store(cvID, progress)
}

func GetCVProgress(cvID string) int {
	val, ok := cvProgressMap.Load(cvID)
	if !ok {
		return 0
	}
	return val.(int)
}

func DeleteCVProgress(cvID string) {
	cvProgressMap.Delete(cvID)
}


// levenshtein computes the edit distance between two strings (case-insensitive).
func levenshtein(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			m := ins
			if del < m {
				m = del
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// toTitleCase converts a string to "Title Case" (first letter of each word uppercase).
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// fuzzyMatchJobRole finds the best matching job role using Levenshtein distance.
// Returns the matched JobRole and true if a match within maxDistance is found.
func fuzzyMatchJobRole(allRoles []gorm_model.JobRole, input string, maxDistance int) (*gorm_model.JobRole, bool) {
	inputLower := strings.ToLower(strings.TrimSpace(input))
	bestDist := maxDistance + 1
	var bestMatch *gorm_model.JobRole
	for i := range allRoles {
		dist := levenshtein(inputLower, strings.ToLower(allRoles[i].Name))
		if dist < bestDist {
			bestDist = dist
			bestMatch = &allRoles[i]
		}
	}
	if bestMatch != nil && bestDist <= maxDistance {
		return bestMatch, true
	}
	return nil, false
}

// getOrCreateUndefinedSector fetches or creates the "undefined" sector.
func getOrCreateUndefinedSector(db *gorm.DB) (*gorm_model.Sector, error) {
	var sector gorm_model.Sector
	if err := db.Where("LOWER(name) = ?", "undefined").First(&sector).Error; err == nil {
		return &sector, nil
	}
	sector = gorm_model.Sector{
		ID:       uuid.New().String(),
		Name:     "Undefined",
		IsActive: true,
	}
	if err := db.Create(&sector).Error; err != nil {
		return nil, err
	}
	return &sector, nil
}

// getOrCreateSectorByName fetches or creates a sector by its name (case-insensitive).
func getOrCreateSectorByName(db *gorm.DB, name string) (*gorm_model.Sector, error) {
	var sector gorm_model.Sector
	if err := db.Where("LOWER(name) = ?", strings.ToLower(name)).First(&sector).Error; err == nil {
		return &sector, nil
	}
	sector = gorm_model.Sector{
		ID:       uuid.New().String(),
		Name:     toTitleCase(name),
		IsActive: true,
	}
	if err := db.Create(&sector).Error; err != nil {
		return nil, err
	}
	return &sector, nil
}

type CVUsecase interface {
	UploadCV(ctx context.Context, userID string, fileHeader *multipart.FileHeader) response.Base
	GetParsedCV(ctx context.Context, userID string) response.Base
	ConfirmCV(ctx context.Context, userID string, editedData map[string]interface{}) response.Base
	GenerateCVLink(ctx context.Context, userID string) response.Base
}

type cvUsecase struct {
	gormRepo    domain.GormRepo
	storageRepo domain.StorageRepo
	mqRepo      domain.MessageBroker // <--- Inject MessageBroker
	timeout     time.Duration
}

func NewCVUsecase(gormRepo domain.GormRepo, storageRepo domain.StorageRepo, mqRepo domain.MessageBroker, timeout time.Duration) CVUsecase {
	return &cvUsecase{
		gormRepo:    gormRepo,
		storageRepo: storageRepo,
		mqRepo:      mqRepo,
		timeout:     timeout,
	}
}

func (u *cvUsecase) UploadCV(ctx context.Context, userID string, fileHeader *multipart.FileHeader) response.Base {
	// 1. Open File
	file, err := fileHeader.Open()
	if err != nil {
		return response.Error(http.StatusBadRequest, "failed to open file")
	}
	defer file.Close()

	// 2. Upload to Storage (S3/Minio)
	// Generate a unique key, e.g., "cvs/userID/filename.pdf"
	objectKey := fmt.Sprintf("cvs/%s/%s", userID, fileHeader.Filename)
	_, err = u.storageRepo.UploadFilePublic(objectKey, file, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return response.Error(http.StatusInternalServerError, "upload failed: "+err.Error())
	}

	// 3. Save Metadata to DB (Upsert Logic)
	existingCV, err := u.gormRepo.GetCVByUserID(ctx, userID)
	if err != nil {
		// CV not found, insert new
		cv := &gorm_model.CV{
			ID:       uuid.New().String(),
			UserID:   userID,
			Filename: fileHeader.Filename,
			Path:     objectKey, // Store relative path
			// ParsedData is omitted so it correctly inserts NULL
		}

		if err := u.gormRepo.CreateCV(ctx, cv); err != nil {
			helpers.LogActivity(ctx, u.gormRepo, "Upload", "CV", userID, nil, false)
			return response.Error(http.StatusInternalServerError, "db save failed")
		}
		existingCV = cv
		SetCVProgress(existingCV.ID, 20)
	} else {
		// CV exists, update it
		existingCV.Filename = fileHeader.Filename
		existingCV.Path = objectKey
		existingCV.ParsedData = nil    // Reset parsed data
		existingCV.Status = "UPLOADED" // Reset status

		if err := u.gormRepo.UpdateCV(ctx, existingCV); err != nil {
			helpers.LogActivity(ctx, u.gormRepo, "Upload", "CV", userID, nil, false)
			return response.Error(http.StatusInternalServerError, "db update failed")
		}
		SetCVProgress(existingCV.ID, 20)
	}

	helpers.LogActivity(ctx, u.gormRepo, "Upload", "CV", userID, nil, true)

	// 4. Publish Event (Only if AI Module is Enabled)
	go func(cv *gorm_model.CV) {
		bgCtx := context.Background()

		// Check if AI module is enabled
		settings, err := u.gormRepo.GetSystemSetting(bgCtx)
		if err != nil || !settings.IsAIModeEnabled {
			fmt.Println("AI Module is disabled or failed to fetch settings, skipping parsing event")
			return
		}

		if u.mqRepo == nil {
			fmt.Println("mqRepo is nil, skipping event publishing")
			return
		}

		// Here we just fire and forget with a new context.
		err = u.mqRepo.Publish(bgCtx, os.Getenv("RABBITMQ_QUEUE_CV_UPLOAD"), map[string]interface{}{
			"event":       "cv_uploaded",
			"user_id":     cv.UserID,
			"cv_id":       cv.ID,
			"path":        cv.Path,
			"uploaded_at": time.Now(),
		})
		if err != nil {
			fmt.Printf("failed to publish message: %v\n", err)
		}
	}(existingCV)

	existingCV.Progress = 20
	return response.Success(existingCV.ToCVResp())
}

func (u *cvUsecase) GetParsedCV(ctx context.Context, userID string) response.Base {
	cv, err := u.gormRepo.GetCVByUserID(ctx, userID)
	if err != nil {
		return response.Error(http.StatusNotFound, "cv not found")
	}

	// Double check for logic safety
	if cv.UserID != userID {
		return response.Error(http.StatusForbidden, "not authorized to view this cv")
	}

	var parsedData interface{}
	if cv.ParsedData != nil {
		var parsedMap map[string]interface{}
		if err := json.Unmarshal([]byte(*cv.ParsedData), &parsedMap); err == nil {
			// Translate AI generated Job Role names using fuzzy matching, return names (not IDs)
			if roles, ok := parsedMap["applied_roles"].([]interface{}); ok {
				var roleNames []interface{}
				db := u.gormRepo.GetDB()

				// Load all job roles once for fuzzy matching
				var allRoles []gorm_model.JobRole
				db.Find(&allRoles)

				for _, r := range roles {
					var roleName string
					var sectorName string

					if m, ok := r.(map[string]interface{}); ok {
						roleName, _ = m["name"].(string)
						sectorName, _ = m["sector"].(string)
					} else if s, ok := r.(string); ok {
						roleName = s
					}

					if roleName == "" {
						continue
					}

					if match, found := fuzzyMatchJobRole(allRoles, roleName, 2); found {
						// Found existing role, use its canonical name and its sector name
						sName := "Undefined"
						if match.Sector != nil {
							sName = match.Sector.Name
						}
						roleNames = append(roleNames, map[string]string{
							"name":   match.Name,
							"sector": sName,
						})
					} else {
						// New role, preserve the sector suggested by AI, or fallback to Undefined
						if sectorName == "" {
							sectorName = "Undefined"
						}
						roleNames = append(roleNames, map[string]string{
							"name":   toTitleCase(roleName),
							"sector": toTitleCase(sectorName),
						})
					}
				}
				parsedMap["applied_roles"] = roleNames
			}

			// Translate AI generated location "Kabupaten, Provinsi" string to concrete UUID
			if kabName, ok := parsedMap["kabupaten_kota"].(string); ok && kabName != "" {
				searchName := strings.Split(kabName, ",")[0]
				searchName = strings.TrimSpace(searchName)
				var kb gorm_model.KabupatenKota
				if err := u.gormRepo.GetDB().Where("name ILIKE ?", "%"+searchName+"%").First(&kb).Error; err == nil {
					parsedMap["kabupaten_kota"] = kb.ID
				}
			}

			parsedData = parsedMap
		} else {
			// Fallback raw if unpack fails
			parsedData = json.RawMessage(*cv.ParsedData)
		}
	}

	progress := GetCVProgress(cv.ID)
	if cv.Status == "PARSED" {
		progress = 100
		DeleteCVProgress(cv.ID)
	} else if progress == 0 {
		progress = 20
	}

	return response.Success(map[string]interface{}{
		"id":          cv.ID,
		"parsed_data": parsedData,
		"status":      cv.Status,
		"progress":    progress,
	})
}

func (u *cvUsecase) ConfirmCV(ctx context.Context, userID string, editedData map[string]interface{}) response.Base {
	cv, err := u.gormRepo.GetCVByUserID(ctx, userID)
	if err != nil {
		return response.Error(http.StatusNotFound, "cv not found")
	}

	if cv.UserID != userID {
		return response.Error(http.StatusForbidden, "not authorized to confirm this cv")
	}

	if cv.Status != "PARSED" {
		return response.Error(http.StatusBadRequest, "cv is not in PARSED status")
	}

	// Unpack the existing Database ParsedData into a root map
	var unifiedData map[string]interface{}
	if cv.ParsedData != nil {
		if err := json.Unmarshal([]byte(*cv.ParsedData), &unifiedData); err != nil {
			return response.Error(http.StatusInternalServerError, "failed to parse existing cv data")
		}
	} else {
		unifiedData = make(map[string]interface{})
	}

	// Merge all incoming EditedData (the patch) directly on top of the root map
	for key, value := range editedData {
		unifiedData[key] = value
	}
	kabID := ""
	if k, ok := unifiedData["kabupaten_kota_id"].(string); ok {
		kabID = k
	}

	// Update User Table Columns
	user, err := u.gormRepo.FetchOneUser(ctx, gorm_model.UserFilter{DefaultFilter: gorm_model.DefaultFilter{ID: userID}})
	if err == nil && user != nil {
		if val, ok := unifiedData["school_university"].(string); ok {
			user.SchoolUniversity = &val
		}

		if val, ok := unifiedData["major"].(string); ok {
			user.Major = &val
		}

		if val, ok := unifiedData["gpa"]; ok {
			if v, ok := val.(string); ok {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					user.Gpa = &f
				}
			} else if v, ok := val.(float64); ok {
				user.Gpa = &v
			}
		}

		if val, ok := unifiedData["years_experience"]; ok {
			if v, ok := val.(string); ok {
				if i, err := strconv.Atoi(v); err == nil {
					user.YearsExperience = &i
				}
			} else if v, ok := val.(float64); ok {
				i := int(v)
				user.YearsExperience = &i
			}
		}

		if val, ok := unifiedData["summary"].(string); ok {
			user.Summary = &val
		}

		if skillsArr, ok := unifiedData["tech_stack"]; ok {
			if marshaled, err := json.Marshal(skillsArr); err == nil {
				skillsStr := string(marshaled)
				user.TechStack = &skillsStr
			}
		}

		if kabID != "" {
			user.KabupatenKotaId = &kabID
			// Look up the actual name AND its province for the AI Payload since the AI RabbitMQ stream expects pure text
			var kb gorm_model.KabupatenKota
			if err := u.gormRepo.GetDB().Where("id = ?", kabID).First(&kb).Error; err == nil {
				var prov gorm_model.Provinsi
				if err := u.gormRepo.GetDB().Where("id = ?", kb.ProvinsiID).First(&prov).Error; err == nil {
					unifiedData["kabupaten_kota"] = kb.Name + ", " + prov.Name
				} else {
					unifiedData["kabupaten_kota"] = kb.Name
				}
				delete(unifiedData, "kabupaten_kota_id") // Keep AI payload clean
			}
		} else if kabName, ok := unifiedData["kabupaten_kota"].(string); ok && kabName != "" {
			// Fallback: If no Dropdown ID present, see if AI generated string exists
			var kb gorm_model.KabupatenKota
			if err := u.gormRepo.GetDB().Where("name ILIKE ?", "%"+kabName+"%").First(&kb).Error; err == nil {
				user.KabupatenKotaId = &kb.ID
			}
		}

		// Handle JobRoles Many-to-Many logic with fuzzy matching and auto-generation
		rolesSource := "applied_roles"
		if _, ok := unifiedData["job_role_ids"]; ok {
			rolesSource = "job_role_ids"
		}

		if roles, ok := unifiedData[rolesSource].([]interface{}); ok {
			// Enforce maximum of 3 job roles per user
			const maxJobRoles = 3
			if len(roles) > maxJobRoles {
				roles = roles[:maxJobRoles]
			}

			db := u.gormRepo.GetDB()

			// Load all job roles once for fuzzy matching, including sector info
			var allRoles []gorm_model.JobRole
			db.Preload("Sector").Find(&allRoles)

			var jobRoles []gorm_model.JobRole
			for _, r := range roles {
				var roleName string
				var sectorName string

				if m, ok := r.(map[string]interface{}); ok {
					roleName, _ = m["name"].(string)
					sectorName, _ = m["sector"].(string)
				} else if s, ok := r.(string); ok {
					if strings.HasPrefix(s, "NEW_ROLE:") {
						parts := strings.Split(s, ":")
						if len(parts) >= 3 {
							sectorName = parts[1]
							roleName = parts[2]
						}
					} else {
						roleName = s
					}
				}

				if roleName == "" {
					continue
				}

				var jr gorm_model.JobRole

				// 1. If it's already a UUID, look up by ID directly
				if _, err := uuid.Parse(roleName); err == nil {
					if err := db.Where("id = ?", roleName).Preload("Sector").First(&jr).Error; err == nil {
						jobRoles = append(jobRoles, jr)
					}
					continue
				}

				// 2. Fuzzy match against all existing roles (handles small typos like "forntend developer" → "Frontend Developer")
				if match, found := fuzzyMatchJobRole(allRoles, roleName, 2); found {
					jobRoles = append(jobRoles, *match)
					continue
				}

				// 3. Not found — auto-generate
				titled := toTitleCase(roleName)
				var sector *gorm_model.Sector
				var err error

				if sectorName != "" {
					sector, err = getOrCreateSectorByName(db, sectorName)
				} else {
					sector, err = getOrCreateUndefinedSector(db)
				}

				if err == nil {
					newRole := gorm_model.JobRole{
						ID:       uuid.New().String(),
						Name:     titled,
						SectorID: sector.ID,
					}
					if err := db.Create(&newRole).Error; err == nil {
						jobRoles = append(jobRoles, newRole)
						// Append to allRoles so duplicates in the same confirmation are deduplicated
						allRoles = append(allRoles, newRole)
					}
				}
			}
			if len(jobRoles) > 0 {
				user.JobRoles = jobRoles
				// Replace associations in many2many table
				db.Model(user).Association("JobRoles").Replace(jobRoles)

				// Overwrite applied_roles with canonical objects for the Qdrant index (AI pipeline)
				var roleObjects []map[string]string
				for _, jr := range jobRoles {
					sectorName := "Undefined"
					if jr.Sector != nil {
						sectorName = jr.Sector.Name
					}
					roleObjects = append(roleObjects, map[string]string{
						"name":   jr.Name,
						"sector": sectorName,
					})
				}
				unifiedData["applied_roles"] = roleObjects
			}
		}

		u.gormRepo.UpdateUser(ctx, user)
	}

	// Always inject these strict identifiers so the AI pipeline (`final_cv`) can safely index it
	unifiedData["id"] = cv.ID
	unifiedData["cv_id"] = cv.ID
	unifiedData["user_id"] = cv.UserID
	if user != nil {
		unifiedData["name"] = user.Name
	}

	// Provide the full payload to AI queue, but sanitize the final parsed response locally
	if u.mqRepo != nil {
		go func() {
			bgCtx := context.Background()
			_ = u.mqRepo.Publish(bgCtx, "final_cv", unifiedData) // Using exact parsed string map payload
		}()
	}

	editedJSON, err := json.Marshal(unifiedData)
	if err != nil {
		return response.Error(http.StatusBadRequest, "invalid edited data")
	}

	jsonStr := string(editedJSON)
	cv.ParsedData = &jsonStr
	cv.Status = "CONFIRMED"

	if err := u.gormRepo.UpdateCV(ctx, cv); err != nil {
		helpers.LogActivity(ctx, u.gormRepo, "Confirm", "CV", userID, editedData, false)
		return response.Error(http.StatusInternalServerError, "failed to update cv")
	}

	helpers.LogActivity(ctx, u.gormRepo, "Confirm", "CV", userID, editedData, true)

	return response.Success(cv.ToCVResp())
}

func (u *cvUsecase) GenerateCVLink(ctx context.Context, userID string) response.Base {
	cv, err := u.gormRepo.GetCVByUserID(ctx, userID)
	if err != nil {
		return response.Error(http.StatusNotFound, "cv not found")
	}

	if cv.UserID != userID {
		return response.Error(http.StatusForbidden, "not authorized to view this cv")
	}

	// Generate 1-hour presigned view link
	expireDuration := time.Hour
	presignedLink := u.storageRepo.GetPresignedLink(cv.Path, &expireDuration)

	res := gorm_model.CVPrivateResp{
		ID:        cv.ID,
		Name:      cv.Filename,
		URL:       presignedLink,
		CreatedAt: cv.CreatedAt,
	}

	return response.Success(res)
}
