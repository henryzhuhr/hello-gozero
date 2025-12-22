// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	userConstant "hello-gozero/internal/constant/user"
	userDto "hello-gozero/internal/dto/user"
	userEntity "hello-gozero/internal/entity/user"
	"hello-gozero/internal/svc"
)

var (
	ErrUsernameExists = errors.New("username already exists")
)

type CreateUserService struct {
	Logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建用户
func NewCreateUserService(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserService {
	return &CreateUserService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserService) CreateUser(req *userDto.CreateUserReq) (resp *userDto.CreateUserResp, err error) {
	// 1. 检查用户名是否已存在
	exists, err := l.svcCtx.Repository.User.Exists(l.ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %v", err)
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// 2. 检查邮箱是否已存在（如果提供）
	if req.Email != "" {
		existingUser, err := l.svcCtx.Repository.User.GetByEmail(l.ctx, req.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check email existence: %v", err)
		}
		if existingUser != nil {
			return nil, errors.New("email already exists")
		}
	}

	// 3. 检查手机号是否已存在（如果提供）
	if req.Phone != "" {
		existingUser, err := l.svcCtx.Repository.User.GetByPhone(l.ctx, req.Phone)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			l.Logger.Errorf("Failed to check phone existence: %v", err)
			return nil, errors.New("failed to check phone")
		}
		if existingUser != nil {
			return nil, errors.New("phone already exists")
		}
	}

	// 4. 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// 5. 创建用户实体
	user := &userEntity.User{
		// ID: 由 [userEntity.User.BeforeCreate] hook 自动生成
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
		Status:   userConstant.StatusActive, // 默认正常状态
	}

	// 6. 保存到数据库
	if err := l.svcCtx.Repository.User.Create(l.ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// 7. 返回结果
	return &userDto.CreateUserResp{
		Id: string(user.ID),
	}, nil
}
