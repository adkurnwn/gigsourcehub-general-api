package gormrepo

import (
	"context"
	"time"

	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
)

func (r *gormRepo) CreateCandidateStatusHistory(ctx context.Context, model *gorm_model.CandidateStatusHistory) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormRepo) FetchUpcomingInterviewsForAdmin(ctx context.Context, adminID string, limit int) ([]gorm_model.Interview, error) {
	var list []gorm_model.Interview
	err := r.db.WithContext(ctx).
		Preload("CandidateUser").
		Preload("Stage").
		Preload("Subrequest").
		Preload("Subrequest.Request").
		Preload("Subrequest.JobRole").
		Where("scheduled_at >= NOW() AND status IN ('SCHEDULED', 'RESCHEDULED')").
		Order("scheduled_at ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *gormRepo) FetchRequestsSummaryStats(ctx context.Context) (gorm_model.RequestsSummaryResp, error) {
	var stats gorm_model.RequestsSummaryResp

	var total, fulfilled, pending, inProgress int64
	r.db.WithContext(ctx).Model(&gorm_model.Request{}).Count(&total)
	r.db.WithContext(ctx).Model(&gorm_model.Request{}).Where("status = 'DONE'").Count(&fulfilled)
	r.db.WithContext(ctx).Model(&gorm_model.Request{}).Where("status = 'PENDING'").Count(&pending)
	r.db.WithContext(ctx).Model(&gorm_model.Request{}).Where("status IN ('ACCEPTED', 'PROCESSING')").Count(&inProgress)

	stats.TotalRequests = int(total)
	stats.FulfilledRequests = int(fulfilled)
	stats.WaitingValidationRequests = int(pending)
	stats.InProgressRequests = int(inProgress)

	if total > 0 {
		stats.RequestFulfillmentPercentage = float64(fulfilled) / float64(total) * 100
	}

	var reqHeadcount int64
	var filledHeadcount int64

	r.db.WithContext(ctx).Model(&gorm_model.Request{}).
		Where("status != 'REJECTED'").
		Select("COALESCE(SUM(required_headcount), 0)").
		Row().Scan(&reqHeadcount)

	r.db.WithContext(ctx).Table("subrequests sr").
		Joins("join requests r on sr.request_id = r.id").
		Where("sr.is_filled = ? and r.status != ? and sr.deleted_at is null and r.deleted_at is null", true, "REJECTED").
		Count(&filledHeadcount)

	stats.RequiredHeadcount = int(reqHeadcount)
	stats.FilledHeadcount = int(filledHeadcount)
	if reqHeadcount > 0 {
		stats.HeadcountFulfillmentPercentage = float64(filledHeadcount) / float64(reqHeadcount) * 100
	}

	return stats, nil
}

func (r *gormRepo) FetchDashboardAlertsForAdmin(ctx context.Context, adminID string) (gorm_model.DashboardAlertResp, error) {
	var alerts gorm_model.DashboardAlertResp

	var pendingApprovals int64
	r.db.WithContext(ctx).Model(&gorm_model.ApprovalRequest{}).Where("status = 'PENDING'").Count(&pendingApprovals)
	alerts.PendingApprovals = int(pendingApprovals)

	var overdueRequests int64
	r.db.WithContext(ctx).Model(&gorm_model.Request{}).
		Where("status NOT IN ('DONE', 'REJECTED') AND due_date < CURRENT_DATE").
		Count(&overdueRequests)
	alerts.OverdueRequests = int(overdueRequests)

	var todayInterviews int64
	r.db.WithContext(ctx).Model(&gorm_model.Interview{}).
		Where("status = 'SCHEDULED' AND DATE(scheduled_at) = CURRENT_DATE").
		Count(&todayInterviews)
	alerts.InterviewsScheduledToday = int(todayInterviews)

	var unassignedRequests int64
	r.db.WithContext(ctx).Model(&gorm_model.Request{}).
		Where("status = 'PENDING' AND admin_user_id IS NULL").
		Count(&unassignedRequests)
	alerts.UnassignedRequests = int(unassignedRequests)

	var expiringPlacements int64
	r.db.WithContext(ctx).Model(&gorm_model.OnboardHistory{}).
		Where("end_date BETWEEN CURRENT_DATE AND (CURRENT_DATE + INTERVAL '30 days')").
		Count(&expiringPlacements)
	alerts.ExpiringPlacements = int(expiringPlacements)

	return alerts, nil
}

func (r *gormRepo) FetchRecentActivitiesForAdmin(ctx context.Context, limit int) ([]gorm_model.LogActivity, error) {
	var list []gorm_model.LogActivity
	err := r.db.WithContext(ctx).
		Preload("Actor").
		Where("is_success = ? AND module IN ('Request', 'Interview', 'Offering', 'Candidate', 'Approval')", true).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *gormRepo) FetchDashboardAnalytics(ctx context.Context, period string) (gorm_model.DashboardAnalyticsResp, error) {
	var analytics gorm_model.DashboardAnalyticsResp

	var trendItems []gorm_model.TrendPeriodItem
	for i := 5; i >= 0; i-- {
		t := time.Now().AddDate(0, -i, 0)
		monthStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)
		monthName := t.Format("Jan 2006")

		var appCount, intCount, hireCount int64
		r.db.WithContext(ctx).Model(&gorm_model.SubrequestCandidate{}).
			Where("created_at BETWEEN ? AND ?", monthStart, monthEnd).
			Count(&appCount)
		r.db.WithContext(ctx).Model(&gorm_model.Interview{}).
			Where("scheduled_at BETWEEN ? AND ?", monthStart, monthEnd).
			Count(&intCount)
		r.db.WithContext(ctx).Model(&gorm_model.OnboardHistory{}).
			Where("created_at BETWEEN ? AND ?", monthStart, monthEnd).
			Count(&hireCount)

		trendItems = append(trendItems, gorm_model.TrendPeriodItem{
			Period:     monthName,
			Applicants: int(appCount),
			Interviews: int(intCount),
			Hires:      int(hireCount),
		})
	}
	analytics.Trends = trendItems

	return analytics, nil
}

func (r *gormRepo) FetchSuperadminDashboardStats(ctx context.Context) (gorm_model.SuperadminDashboardResp, error) {
	var stats gorm_model.SuperadminDashboardResp

	var pendingApprovals int64
	r.db.WithContext(ctx).Model(&gorm_model.ApprovalRequest{}).Where("status = 'PENDING'").Count(&pendingApprovals)
	stats.PendingApprovalsCount = int(pendingApprovals)

	var totalCandidates, totalAdmins, totalEmployees int64

	r.db.WithContext(ctx).Table("users u").
		Joins("join system_roles sr on u.system_role_id = sr.id").
		Where("sr.name = ? and u.deleted_at is null", "Candidate").
		Count(&totalCandidates)

	r.db.WithContext(ctx).Table("users u").
		Joins("join system_roles sr on u.system_role_id = sr.id").
		Where("sr.name = ? and u.deleted_at is null", "Admin").
		Count(&totalAdmins)

	r.db.WithContext(ctx).Table("users u").
		Joins("join system_roles sr on u.system_role_id = sr.id").
		Where("sr.name = ? and u.deleted_at is null", "Employee").
		Count(&totalEmployees)

	stats.TotalCandidates = int(totalCandidates)
	stats.TotalAdmins = int(totalAdmins)
	stats.TotalEmployees = int(totalEmployees)

	var activeVacancies int64
	r.db.WithContext(ctx).Model(&gorm_model.JobVacancy{}).
		Where("status = 'PUBLISHED' AND (takedown_date IS NULL OR takedown_date >= CURRENT_DATE)").
		Count(&activeVacancies)
	stats.ActiveJobVacancies = int(activeVacancies)

	var aiMode bool
	r.db.WithContext(ctx).Table("system_settings").Select("is_ai_mode_enabled").Row().Scan(&aiMode)
	stats.IsAIModeEnabled = aiMode

	var recent []gorm_model.ApprovalRequest
	r.db.WithContext(ctx).
		Preload("RequestedByAdmin").
		Where("status IN ('APPROVED', 'REJECTED')").
		Order("updated_at DESC").
		Limit(5).
		Find(&recent)

	var approvalsMap []map[string]interface{}
	for _, app := range recent {
		adminName := "-"
		if app.RequestedByAdmin != nil {
			adminName = app.RequestedByAdmin.Name
		}
		approvalsMap = append(approvalsMap, map[string]interface{}{
			"id":                app.ID,
			"table_name":        app.TableName,
			"action":            app.Action,
			"status":            app.Status,
			"requested_by_name": adminName,
			"created_at":        app.CreatedAt,
			"updated_at":        app.UpdatedAt,
		})
	}
	stats.RecentApprovals = approvalsMap

	return stats, nil
}
