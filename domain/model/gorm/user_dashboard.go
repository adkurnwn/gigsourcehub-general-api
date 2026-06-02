package gorm_model

type CandidateTableResp struct {
	UserResp
	SubrequestID *string `json:"subrequest_id,omitempty"`
	RequestID    *string `json:"request_id,omitempty"`
	ProjectName  *string `json:"project_name,omitempty"`
	JobRoleName  *string `json:"job_role_name,omitempty"`
}

func (row *User) ToCandidateTableResp(info *ActiveSubrequestInfo) CandidateTableResp {
	resp := CandidateTableResp{
		UserResp: row.ToUserResp(),
	}

	if info == nil {
		return resp
	}

	resp.SubrequestID = &info.SubrequestID
	resp.RequestID = &info.RequestID
	projectName := info.ProjectName
	resp.ProjectName = &projectName
	jobRoleName := info.JobRole
	resp.JobRoleName = &jobRoleName

	return resp
}