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
	"time"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/google/uuid"
)

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
			return response.Error(http.StatusInternalServerError, "db save failed")
		}
		existingCV = cv
	} else {
		// CV exists, update it
		existingCV.Filename = fileHeader.Filename
		existingCV.Path = objectKey
		existingCV.ParsedData = nil    // Reset parsed data
		existingCV.Status = "UPLOADED" // Reset status

		if err := u.gormRepo.UpdateCV(ctx, existingCV); err != nil {
			return response.Error(http.StatusInternalServerError, "db update failed")
		}
	}

	// 4. Publish Event
	go func(cv *gorm_model.CV) {
		// Use a detached context or background context for async publishing
		// to avoid cancellation if the request context is cancelled.

		if u.mqRepo == nil {
			fmt.Println("mqRepo is nil, skipping event publishing")
			return
		}

		bgCtx := context.Background()
		// Here we just fire and forget with a new context.
		err := u.mqRepo.Publish(bgCtx, os.Getenv("RABBITMQ_QUEUE_CV_UPLOAD"), map[string]interface{}{
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
			// Translate AI generated Job Role names to concrete UUID equivalents
			if roles, ok := parsedMap["applied_roles"].([]interface{}); ok {
				var roleIDs []interface{}
				db := u.gormRepo.GetDB()
				for _, r := range roles {
					if roleName, ok := r.(string); ok && roleName != "" {
						var jr gorm_model.JobRole
						if err := db.Where("name ILIKE ?", "%"+roleName+"%").First(&jr).Error; err == nil {
							roleIDs = append(roleIDs, jr.ID)
						} else {
							roleIDs = append(roleIDs, roleName) // Keep fallback text
						}
					}
				}
				parsedMap["applied_roles"] = roleIDs
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

	return response.Success(map[string]interface{}{
		"id":          cv.ID,
		"parsed_data": parsedData,
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

		// Handle JobRoles Many-to-Many logic converting the string roles to UUID models
		if roles, ok := unifiedData["applied_roles"].([]interface{}); ok {
			var jobRoles []gorm_model.JobRole
			for _, r := range roles {
				if roleStr, ok := r.(string); ok && roleStr != "" {
					var jr gorm_model.JobRole
					// Check if it's already a UUID dropdown assignment, otherwise ILIKE format it
					if _, err := uuid.Parse(roleStr); err == nil {
						if err := u.gormRepo.GetDB().Where("id = ?", roleStr).First(&jr).Error; err == nil {
							jobRoles = append(jobRoles, jr)
						}
					} else {
						if err := u.gormRepo.GetDB().Where("name ILIKE ?", "%"+roleStr+"%").First(&jr).Error; err == nil {
							jobRoles = append(jobRoles, jr)
						}
					}
				}
			}
			if len(jobRoles) > 0 {
				user.JobRoles = jobRoles
				// Replace associations in many2many table
				u.gormRepo.GetDB().Model(user).Association("JobRoles").Replace(jobRoles)

				// Optional: Overwrite array from UUIDs/Mixed strings onto purely JobRole.Name structs for Qdrant index
				var strRoles []string
				for _, jr := range jobRoles {
					strRoles = append(strRoles, jr.Name)
				}
				unifiedData["applied_roles"] = strRoles
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
		return response.Error(http.StatusInternalServerError, "failed to update cv")
	}

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
		ID:   cv.ID,
		Name: cv.Filename,
		URL:  presignedLink,
	}

	return response.Success(res)
}
