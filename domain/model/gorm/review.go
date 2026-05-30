package gorm_model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Review struct {
	ID                          string          `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	SubrequestID                string          `gorm:"column:subrequest_id;type:uuid;not null"`
	Subrequest                  *Subrequest     `gorm:"foreignKey:SubrequestID"`
	CandidateUserID             string          `gorm:"column:candidate_user_id;type:uuid;not null"`
	CandidateUser               *User           `gorm:"foreignKey:CandidateUserID"`
	EmployeeUserID              string          `gorm:"column:employee_user_id;type:uuid;not null"`
	EmployeeUser                *User           `gorm:"foreignKey:EmployeeUserID"`
	OnboardHistoryID            string          `gorm:"column:onboard_history_id;type:uuid;not null"`
	OnboardHistory              *OnboardHistory `gorm:"foreignKey:OnboardHistoryID"`
	FinalRecommendation         string          `gorm:"column:final_recommendation;type:review_final_recommendation"`
	Notes                       *string         `gorm:"column:notes;type:text"`
	Answers                     []ReviewAnswer  `gorm:"foreignKey:ReviewID"`
	CreatedAt                   time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                   time.Time       `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt                   gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}

type ReviewQuestion struct {
	ID            string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	Indicator     string         `gorm:"column:indicator;type:review_indicator;not null"`
	QuestionText  string         `gorm:"column:question_text;type:text;not null"`
	QuestionOrder int            `gorm:"column:question_order;type:int;not null"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type ReviewQuestionResp struct {
	ID            string `json:"id"`
	Indicator     string `json:"indicator"`
	QuestionText  string `json:"question_text"`
	QuestionOrder int    `json:"question_order"`
}

type ReviewAnswer struct {
	ID         string         `gorm:"column:id;primarykey;type:uuid;default:uuid_generate_v4()"`
	ReviewID   string         `gorm:"column:review_id;type:uuid;not null"`
	Review     *Review        `gorm:"foreignKey:ReviewID"`
	QuestionID string         `gorm:"column:question_id;type:uuid;not null"`
	Question   *ReviewQuestion `gorm:"foreignKey:QuestionID"`
	Score      int            `gorm:"column:score;type:int;not null"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type ReviewAnswerResp struct {
	QuestionID    string `json:"question_id"`
	Indicator     string `json:"indicator"`
	QuestionText  string `json:"question_text"`
	QuestionOrder int    `json:"question_order"`
	Score         int    `json:"score"`
}

type ReviewResp struct {
	ID                  string             `json:"id"`
	SubrequestID        string             `json:"subrequest_id"`
	CandidateUserID     string             `json:"candidate_user_id"`
	EmployeeUserID      string             `json:"employee_user_id"`
	OnboardHistoryID    string             `json:"onboard_history_id"`
	FinalRecommendation *string            `json:"final_recommendation,omitempty"`
	Notes               *string            `json:"notes,omitempty"`
	Answers             []ReviewAnswerResp `json:"answers"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

func (m *Review) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

func (m *ReviewQuestion) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

func (m *ReviewQuestion) ToReviewQuestionResp() ReviewQuestionResp {
	return ReviewQuestionResp{
		ID:            m.ID,
		Indicator:     m.Indicator,
		QuestionText:  m.QuestionText,
		QuestionOrder: m.QuestionOrder,
	}
}

func (m *ReviewAnswer) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return
}

func (m *Review) ToReviewResp() ReviewResp {
	var finalRec *string
	if m.FinalRecommendation != "" {
		finalRec = &m.FinalRecommendation
	}

	answers := make([]ReviewAnswerResp, 0, len(m.Answers))
	for _, ans := range m.Answers {
		resp := ReviewAnswerResp{
			QuestionID: ans.QuestionID,
			Score:      ans.Score,
		}
		if ans.Question != nil {
			resp.Indicator = ans.Question.Indicator
			resp.QuestionText = ans.Question.QuestionText
			resp.QuestionOrder = ans.Question.QuestionOrder
		}
		answers = append(answers, resp)
	}

	return ReviewResp{
		ID:                  m.ID,
		SubrequestID:        m.SubrequestID,
		CandidateUserID:     m.CandidateUserID,
		EmployeeUserID:      m.EmployeeUserID,
		OnboardHistoryID:    m.OnboardHistoryID,
		FinalRecommendation: finalRec,
		Notes:               m.Notes,
		Answers:             answers,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}
