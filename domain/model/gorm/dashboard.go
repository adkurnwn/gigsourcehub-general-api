package gorm_model

type RequestsSummaryResp struct {
	TotalRequests                  int     `json:"total_requests"`
	FulfilledRequests              int     `json:"fulfilled_requests"`
	WaitingValidationRequests      int     `json:"waiting_validation_requests"`
	InProgressRequests              int     `json:"in_progress_requests"`
	RequestFulfillmentPercentage   float64 `json:"request_fulfillment_percentage"`
	RequiredHeadcount              int     `json:"required_headcount"`
	FilledHeadcount                int     `json:"filled_headcount"`
	HeadcountFulfillmentPercentage float64 `json:"headcount_fulfillment_percentage"`
}

type DashboardAlertResp struct {
	PendingApprovals         int `json:"pending_approvals"`
	OverdueRequests          int `json:"overdue_requests"`
	InterviewsScheduledToday int `json:"interviews_scheduled_today"`
	UnassignedRequests       int `json:"unassigned_requests"`
	ExpiringPlacements       int `json:"expiring_placements"`
}

type AdminDashboardSummaryResp struct {
	KPIs               map[string]interface{} `json:"kpis"`
	UpcomingInterviews []InterviewResp        `json:"upcoming_interviews"`
	RequestsSummary    RequestsSummaryResp    `json:"requests_summary"`
	Alerts             DashboardAlertResp     `json:"alerts"`
	RecentActivities   []map[string]interface{} `json:"recent_activities"`
}

type TrendPeriodItem struct {
	Period     string `json:"period"`
	Applicants int    `json:"applicants"`
	Interviews int    `json:"interviews"`
	Hires      int    `json:"hires"`
}

type DashboardAnalyticsResp struct {
	Trends []TrendPeriodItem `json:"trends"`
}

type SuperadminDashboardResp struct {
	PendingApprovalsCount int                      `json:"pending_approvals_count"`
	TotalCandidates       int                      `json:"total_candidates"`
	TotalAdmins           int                      `json:"total_admins"`
	TotalEmployees        int                      `json:"total_employees"`
	ActiveJobVacancies    int                      `json:"active_job_vacancies"`
	IsAIModeEnabled       bool                     `json:"is_ai_mode_enabled"`
	RecentApprovals       []map[string]interface{} `json:"recent_approvals"`
}
