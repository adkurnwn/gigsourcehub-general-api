package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ActivityLogRepo interface {
	CreateLogActivity(ctx context.Context, model *gorm_model.LogActivity) error
}

func LogActivity(ctx context.Context, repo ActivityLogRepo, action, module, target string, metadata interface{}, isSuccess bool) {
	ipAddress := GetIPAddress(ctx)
	userAgent := GetUserAgent(ctx)
	endpoint := GetEndpoint(ctx)

	// Title Case formatting
	caser := cases.Title(language.Indonesian)
	titleAction := caser.String(strings.ReplaceAll(action, "_", " "))
	titleModule := caser.String(strings.ReplaceAll(module, "_", " "))

	// Auto-generated description: "Action Module[: Target]"
	description := fmt.Sprintf("%s %s", titleAction, titleModule)
	if target != "" {
		description = fmt.Sprintf("%s: %s", description, target)
	}

	// JSON metadata serialization & Enrichment
	metaMap := make(map[string]interface{})
	if userAgent != "" {
		metaMap["user_agent"] = userAgent
	}
	if endpoint != "" {
		metaMap["endpoint"] = endpoint
	}

	if metadata != nil {
		if s, ok := metadata.(string); ok && s != "" {
			// If it's a JSON string, try to decode into our map
			var existing map[string]interface{}
			if err := json.Unmarshal([]byte(s), &existing); err == nil {
				for k, v := range existing {
					metaMap[k] = v
				}
			} else {
				// Not valid JSON? Store as raw 'data'
				metaMap["data"] = s
			}
		} else if m, ok := metadata.(map[string]interface{}); ok {
			for k, v := range m {
				metaMap[k] = v
			}
		} else {
			// Try to marshal any other type into JSON bytes then unmarshal into metaMap
			marshaled, _ := json.Marshal(metadata)
			var existing map[string]interface{}
			json.Unmarshal(marshaled, &existing)
			for k, v := range existing {
				metaMap[k] = v
			}
		}
	}

	var metaStr *string
	if len(metaMap) > 0 {
		marshaled, _ := json.Marshal(metaMap)
		ms := string(marshaled)
		metaStr = &ms
	}

	var actorIDPtr *string
	if aid := GetActorID(ctx); aid != "" {
		actorIDPtr = &aid
	}

	log := &gorm_model.LogActivity{
		ActorID:     actorIDPtr,
		ActionType:  titleAction,
		Module:      titleModule,
		Description: &description,
		Metadata:    metaStr,
		IPAddress:   &ipAddress,
		IsSuccess:   isSuccess,
	}

	repo.CreateLogActivity(ctx, log)
}
