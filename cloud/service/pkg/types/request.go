package types

const AllNamespace = "all_namespaces"

type (
	// LoginRequest is the request body struct for user login.
	LoginRequest struct {
		Name     string `json:"name" binding:"required"`     // required
		Password string `json:"password" binding:"required"` // required
	}

	UpdateUserPasswordRequest struct {
		New             string `json:"new" binding:"required,password"`     // required
		Old             string `json:"old" binding:"required"`              // required
		ResourceVersion *int64 `json:"resource_version" binding:"required"` // required
		Reset           bool   `json:"reset"`
	}

	UpdateClusterRequest struct {
		AliasName   *string `json:"alias_name" binding:"omitempty"`  // optional
		Description *string `json:"description" binding:"omitempty"` // optional
		// TODO: put resource version in a common struct for updating request only
		ResourceVersion *int64 `json:"resource_version" binding:"required"` // required
	}

	ProtectClusterRequest struct {
		ResourceVersion *int64 `json:"resource_version" binding:"required"` // required
		Protected       bool   `json:"protected" binding:"omitempty"`       // optional
	}

	CreateTenantRequest struct {
		Name        string  `json:"name" binding:"required"`         // required
		Description *string `json:"description" binding:"omitempty"` // optional
	}

	UpdateTenantRequest struct {
		Name            *string `json:"name" binding:"omitempty"`            // optional
		Description     *string `json:"description" binding:"omitempty"`     // optional
		ResourceVersion *int64  `json:"resource_version" binding:"required"` // required
	}

	CreatePlanConfigRequest struct {
		PlanId      int64  `json:"plan_id"`
		Region      string `json:"region"`
		OSImage     string `json:"os_image" binding:"required"`     // 操作系统
		Description string `json:"description" binding:"omitempty"` // optional

		Kubernetes KubernetesSpec `json:"kubernetes"`
		Network    NetworkSpec    `json:"network"`
		Runtime    RuntimeSpec    `json:"runtime"`
	}

	UpdatePlanConfigRequest struct {
		// TODO:
	}

	// PageRequest 分页配置
	PageRequest struct {
		Page  int `form:"page" json:"page"`   // 页数，表示第几页
		Limit int `form:"limit" json:"limit"` // 每页数量
	}
	// QueryOption 搜索配置
	QueryOption struct {
		LabelSelector string `form:"labelSelector" json:"labelSelector"` // 标签搜索
		NameSelector  string `form:"nameSelector" json:"nameSelector"`   // 名称搜索
	}

	// WebSSHRequest 主机 ssh 跳转请求
	WebSSHRequest struct {
		Host     string `form:"host" json:"host" binding:"required"`
		Port     int    `form:"port" json:"port"`
		User     string `form:"user" json:"user" binding:"required"`
		Password string `form:"password" json:"password"`
	}
)

type (
	// PageResponse 分页查询返回值
	PageResponse struct {
		PageRequest `json:",inline"` // 分页请求属性

		Total int         `json:"total"` // 分页总数
		Items interface{} `json:"items"` // 指定页的元素列表
	}
)
