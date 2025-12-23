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
	// 用户名已存在
	ErrUsernameExists = errors.New("username already exists")

	// 手机号已存在
	ErrPhoneExists = errors.New("phone already exists")
)

type RegisterUserService struct {
	Logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建用户
func NewRegisterUserService(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterUserService {
	return &RegisterUserService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterUserService) RegisterUser(req *userDto.RegisterUserReq) (resp *userDto.RegisterUserResp, err error) {
	// 检查用户名是否已存在
	exists, err := l.svcCtx.Repository.User.ExistsByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %v", err)
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// 检查手机号是否已存在，避免用户重复注册
	exists, err = l.svcCtx.Repository.User.ExistsByPhone(l.ctx, req.PhoneCountryCode, req.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check phone existence: %v", err)
	}
	if exists {
		return nil, ErrPhoneExists
	}

	// 检查邮箱是否已存在（如果提供）
	if req.Email != "" {
		existingUser, err := l.svcCtx.Repository.User.GetByEmail(l.ctx, req.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check email existence: %v", err)
		}
		if existingUser != nil {
			return nil, errors.New("email already exists")
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// 5. 创建用户实体
	user := &userEntity.User{
		// ID 由 [userEntity.User.BeforeCreate] hook 自动生成，不应该显式设置
		Username:         req.Username,
		Password:         string(hashedPassword),
		Email:            req.Email,
		PhoneCountryCode: req.PhoneCountryCode,
		PhoneNumber:      req.PhoneNumber,
		Nickname:         req.Nickname,
		Status:           userConstant.StatusActive, // 默认正常状态
	}

	// 6. 保存到数据库
	if err := l.svcCtx.Repository.User.Create(l.ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// 7. 返回结果
	return &userDto.RegisterUserResp{}, nil
}
