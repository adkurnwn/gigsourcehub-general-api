package gorm_model

import "strings"

type OnboardingFilter struct {
	JobRoleName    []string
	ProjectName    []string
	EmployeeUser   []string
	Search         *string
}

func (f OnboardingFilter) GetDescription() string {
	var parts []string
	if len(f.JobRoleName) > 0 {
		parts = append(parts, "Job Role: "+strings.Join(f.JobRoleName, ", "))
	}
	if len(f.ProjectName) > 0 {
		parts = append(parts, "Project: "+strings.Join(f.ProjectName, ", "))
	}
	if len(f.EmployeeUser) > 0 {
		parts = append(parts, "Employee: "+strings.Join(f.EmployeeUser, ", "))
	}
	if len(parts) == 0 {
		return "None"
	}
	res := ""
	for i, p := range parts {
		if i > 0 {
			res += ", "
		}
		res += p
	}
	return res
}

type CandidateRecruitmentFilter struct {
	JobRoleName    []string
	ProjectName    []string
	CandidateLevel []string
	Search         *string
}

func (f CandidateRecruitmentFilter) GetDescription() string {
	var parts []string
	if len(f.JobRoleName) > 0 {
		parts = append(parts, "Job Role: "+strings.Join(f.JobRoleName, ", "))
	}
	if len(f.ProjectName) > 0 {
		parts = append(parts, "Project: "+strings.Join(f.ProjectName, ", "))
	}
	if len(f.CandidateLevel) > 0 {
		parts = append(parts, "Level: "+strings.Join(f.CandidateLevel, ", "))
	}
	if len(parts) == 0 {
		return "None"
	}
	res := ""
	for i, p := range parts {
		if i > 0 {
			res += ", "
		}
		res += p
	}
	return res
}
