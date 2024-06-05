package request

type ListManRoleReq struct {
	Page
	Role         string `json:"role"`
	PlatFormType string `json:"platFormType"`
	RoleStatus   string `json:"roleStatus"`
}

type GetRolePermissionsReq struct {
	Page
	Role           string `json:"role"`
	PlatFormType   string `json:"platFormType"`
	PermissionType string `json:"permissionType"`
}

type CreateRoleReq struct {
	Role         string   `json:"role"`
	RoleName     string   `json:"roleName"`
	RoleStatus   string   `json:"roleStatus"`
	PlatFormType string   `json:"platFormType"`
	Permissions  []string `json:"permissions"`
}

type UpdateRolePermissionsReq struct {
	Role        string   `json:"role"`
	RoleName    string   `json:"roleName"`
	RoleStatus  string   `json:"roleStatus"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleStatusReq struct {
	Role       string `json:"role"`
	RoleStatus string `json:"roleStatus"`
}

type DeleteRoleReq struct {
	Role string `json:"role"`
}
