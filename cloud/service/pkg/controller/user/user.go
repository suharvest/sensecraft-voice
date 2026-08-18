package user

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/token"
	"k8s.io/klog/v2"
)

type UserGetter interface {
	User() Interface
}

type Interface interface {
	Register(ctx context.Context, in RegisterRequest) (*model.User, error)
	Login(ctx context.Context, in LoginRequest) (*LoginResponse, error)
	GetById(ctx context.Context, id int64) (*model.User, error)
	List(ctx context.Context, in ListRequest) (*ListResponse, error)
	Update(ctx context.Context, id int64, in UpdateRequest) (*model.User, error)
	Delete(ctx context.Context, id int64) error
	ChangePassword(ctx context.Context, id int64, in ChangePasswordRequest) error
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=128"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User  *model.User `json:"user"`
	Token string      `json:"token,omitempty"` // JWT，管理端接口需在 Authorization: Bearer 中携带
}

type UpdateRequest struct {
	Username string `json:"username,omitempty"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=128"`
}

type ListRequest struct {
	Offset   int    `form:"offset" binding:"min=0"`
	Limit    int    `form:"limit" binding:"min=1,max=100"`
	Username string `form:"username"` // 用户名模糊搜索
}

type ListResponse struct {
	Total int64         `json:"total"`
	Items []*model.User `json:"items"`
}

type user struct {
	cc      config.Config
	factory db.ShareDaoFactory
}

// hashPassword 对密码进行MD5加密
func hashPassword(password string) string {
	hash := md5.Sum([]byte(password))
	return fmt.Sprintf("%x", hash)
}

// verifyPassword 验证密码
func verifyPassword(password, hashedPassword string) bool {
	return password == hashedPassword
}

func (u *user) Register(ctx context.Context, in RegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	if _, err := u.factory.User().GetByUsername(ctx, in.Username); err == nil {
		klog.Errorf("Username %s already exists", in.Username)
		return nil, errors.ErrInvalidRequest
	}

	// 创建新用户
	obj := &model.User{
		Username: in.Username,
		Password: in.Password,
	}

	if err := u.factory.User().Create(ctx, obj); err != nil {
		klog.Errorf("Failed to create user: %v", err)
		return nil, errors.ErrServerInternal
	}
	return obj, nil
}

func (u *user) Login(ctx context.Context, in LoginRequest) (*LoginResponse, error) {
	// 根据用户名查找用户
	user, err := u.factory.User().GetByUsername(ctx, in.Username)
	if err != nil {
		klog.Errorf("User %s not found: %v", in.Username, err)
		return nil, errors.ErrInvalidRequest
	}
	// 验证密码
	if !verifyPassword(in.Password, user.Password) {
		klog.Errorf("Invalid password for user %s", in.Username)
		return nil, errors.ErrInvalidRequest
	}

	jwtToken, err := token.GenerateToken(user.Id, user.Username, []byte(u.cc.Default.JWTKey))
	if err != nil {
		klog.Errorf("Failed to generate token for user %s: %v", in.Username, err)
		return nil, errors.ErrServerInternal
	}

	return &LoginResponse{
		User:  user,
		Token: jwtToken,
	}, nil
}

func (u *user) GetById(ctx context.Context, id int64) (*model.User, error) {
	obj, err := u.factory.User().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get user by id %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}
	return obj, nil
}

func (u *user) List(ctx context.Context, in ListRequest) (*ListResponse, error) {
	if in.Limit == 0 {
		in.Limit = 20
	}

	// 使用新的带过滤条件的查询方法
	total, err := u.factory.User().CountWithFilter(ctx, in.Username)
	if err != nil {
		klog.Errorf("Failed to count users: %v", err)
		return nil, errors.ErrServerInternal
	}

	items, err := u.factory.User().ListWithFilter(ctx, in.Username, in.Offset, in.Limit)
	if err != nil {
		klog.Errorf("Failed to list users: %v", err)
		return nil, errors.ErrServerInternal
	}

	return &ListResponse{
		Total: total,
		Items: items,
	}, nil
}

func (u *user) Update(ctx context.Context, id int64, in UpdateRequest) (*model.User, error) {
	obj, err := u.factory.User().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get user by id %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}

	// 如果更新了用户名，检查是否与其他用户重复
	if in.Username != "" && in.Username != obj.Username {
		if _, err := u.factory.User().GetByUsername(ctx, in.Username); err == nil {
			klog.Errorf("Username %s already exists", in.Username)
			return nil, errors.ErrInvalidRequest
		}
		obj.Username = in.Username
	}

	if err := u.factory.User().Update(ctx, obj); err != nil {
		klog.Errorf("Failed to update user: %v", err)
		return nil, errors.ErrServerInternal
	}
	return obj, nil
}

func (u *user) Delete(ctx context.Context, id int64) error {
	if err := u.factory.User().Delete(ctx, id); err != nil {
		klog.Errorf("Failed to delete user: %v", err)
		return errors.ErrServerInternal
	}
	return nil
}

func (u *user) ChangePassword(ctx context.Context, id int64, in ChangePasswordRequest) error {
	// 获取用户信息
	user, err := u.factory.User().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get user by id %d: %v", id, err)
		return errors.ErrServerInternal
	}

	// 验证旧密码
	if !verifyPassword(in.OldPassword, user.Password) {
		klog.Errorf("Invalid old password for user %d", id)
		return errors.ErrInvalidRequest
	}

	// 更新密码
	user.Password = in.NewPassword
	if err := u.factory.User().Update(ctx, user); err != nil {
		klog.Errorf("Failed to update password: %v", err)
		return errors.ErrServerInternal
	}

	return nil
}

func NewUser(cfg config.Config, f db.ShareDaoFactory) *user {
	return &user{cc: cfg, factory: f}
}
