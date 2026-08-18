package types

import (
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

type SensecraftVoiceMeta struct {
	// API 对象 ID
	Id int64 `json:"id"`
	// SensecraftVoice 对象版本号
	ResourceVersion int64 `json:"resource_version"`
}

type TimeMeta struct {
	// API 对象创建时间
	GmtCreate time.Time `json:"gmt_create"`
	// API 对象修改时间
	GmtModified time.Time `json:"gmt_modified"`
}

type KubeNode struct {
	Ready    []string `json:"ready"`
	NotReady []string `json:"not_ready"`
}

// Resources kubernetes 的资源信息
// The memory and cpu usage
type Resources struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}

type Tenant struct {
	SensecraftVoiceMeta `json:",inline"`
	TimeMeta      `json:",inline"`

	Name        string `json:"name"`        // 用户名称
	Description string `json:"description"` // 用户描述信息
}

type Audit struct {
	SensecraftVoiceMeta `json:",inline"`
	TimeMeta      `json:",inline"`

	Ip       string                     `json:"ip"`
	Action   string                     `json:"action"`   // 操作动作
	Status   model.AuditOperationStatus `json:"status"`   // 操作状态
	Operator string                     `json:"operator"` // 操作人
	Path     string                     `json:"path"`     // 操作路径
}

type AuthType string

const (
	NoneAuth     AuthType = "none"     // 已开启密码
	KeyAuth      AuthType = "key"      // 密钥
	PasswordAuth AuthType = "password" // 密码
)

type PlanNodeAuth struct {
	Type     AuthType      `json:"type"` // 节点认证模式，支持 key 和 password
	Key      *KeySpec      `json:"key,omitempty"`
	Password *PasswordSpec `json:"password,omitempty"`
}

type KeySpec struct {
	Data string `json:"data,omitempty"`
	File string `json:"-"`
}

type PasswordSpec struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

type PlanConfig struct {
	SensecraftVoiceMeta `json:",inline"`
	TimeMeta      `json:",inline"`

	PlanId     int64          `json:"plan_id,omitempty"` // required
	Region     string         `json:"region"`
	OSImage    string         `json:"os_image"` // 操作系统
	Kubernetes KubernetesSpec `json:"kubernetes"`
	Network    NetworkSpec    `json:"network"`
	Runtime    RuntimeSpec    `json:"runtime"`
}

// TimeSpec 通用时间规格
type TimeSpec struct {
	GmtCreate   interface{} `json:"gmt_create,omitempty"`
	GmtModified interface{} `json:"gmt_modified,omitempty"`
}

type KubeObject struct {
	lock sync.RWMutex

	ReplicaSets []appv1.ReplicaSet
	Pods        []v1.Pod
}

// WebShellOptions ws API 参数定义
type WebShellOptions struct {
	Cluster   string `form:"cluster"`
	Namespace string `form:"namespace"`
	Pod       string `form:"pod"`
	Container string `form:"container"`
	Command   string `form:"command"`
}

// TerminalMessage 定义了终端和容器 shell 交互内容的格式 Operation 是操作类型
// Data 是具体数据内容 Rows和Cols 可以理解为终端的行数和列数，也就是宽、高
type TerminalMessage struct {
	Operation string `json:"operation"`
	Data      string `json:"data"`
	Rows      uint16 `json:"rows"`
	Cols      uint16 `json:"cols"`
}

// TerminalSession 定义 TerminalSession 结构体，实现 PtyHandler 接口
// wsConn 是 websocket 连接
// sizeChan 用来定义终端输入和输出的宽和高
// doneChan 用于标记退出终端
type TerminalSession struct {
	wsConn   *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}
}

type Turn struct {
	StdinPipe io.WriteCloser
	Session   *ssh.Session
	WsConn    *websocket.Conn
}

// ListOptions is the query options to a standard REST list call.
type ListOptions struct {
	Count bool  `form:"count"`
	Limit int64 `form:"limit"`

	PageRequest `json:",inline"` // 分页请求属性
	QueryOption `json:",inline"` // 搜索内容
}

type EventOptions struct {
	Uid        string `form:"uid"`
	Namespace  string `form:"namespace"`
	Name       string `form:"name"`
	Kind       string `form:"kind"`
	Namespaced bool   `form:"namespaced"`
	Limit      int64  `form:"limit"`
}

type PodLogOptions struct {
	Container string `form:"container"`
	TailLines int64  `form:"tailLines"`
}

type KubernetesSpec struct {
	EnablePublicIp    bool   `json:"enable_public_ip"`
	ApiServer         string `json:"api_server"`
	ApiPort           string `json:"api_port"`
	KubernetesVersion string `json:"kubernetes_version"`
	EnableHA          bool   `json:"enable_ha"`
	Register          bool   `json:"register"`
}

type NetworkSpec struct {
	NetworkInterface string `json:"network_interface"` // 网口，默认 eth0
	Cni              string `json:"cni"`
	PodNetwork       string `json:"pod_network"`
	ServiceNetwork   string `json:"service_network"`
	KubeProxy        string `json:"kube_proxy"`
}

type RuntimeSpec struct {
	Runtime string `json:"runtime"`
}

// OpenAI相关类型定义

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	Timeout     int     `json:"timeout"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Model       string  `json:"model"`
}

// OpenAIChatRequest OpenAI聊天请求
type OpenAIChatRequest struct {
	SessionID           string `json:"session_id,omitempty"`
	Message             string `json:"message"`
	UserID              string `json:"user_id"`
	Stream              bool   `json:"stream,omitempty"`
	SystemPromptID      int64  `json:"system_prompt_id,omitempty"`      // 系统提示词ID
	SystemPromptContent string `json:"system_prompt_content,omitempty"` // 系统提示词内容
}

// OpenAIChatResponse OpenAI聊天响应
type OpenAIChatResponse struct {
	SessionID string      `json:"session_id"`
	Message   string      `json:"message"`
	Usage     OpenAIUsage `json:"usage"`
	CreatedAt int64       `json:"created_at"`
}

// OpenAIUsage OpenAI使用情况
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
