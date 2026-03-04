package request_model

type CreateUserBySuperadminRequest struct {
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Password       string  `json:"password"`
	AssignedRoleId *string `json:"assigned_role_id"`
	RoleSystemId   *string `json:"role_system_id"`
	AccountStatus  *string `json:"account_status"`
}

type EditUserBySuperadminRequest struct {
	Name           *string `json:"name"`
	AssignedRoleId *string `json:"assigned_role_id"`
	AccountStatus  *string `json:"account_status"`
}
