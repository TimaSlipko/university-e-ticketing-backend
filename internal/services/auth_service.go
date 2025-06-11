package services

import (
	"errors"
	"time"

	"eticketing/internal/models"
	"eticketing/internal/repositories"
	"eticketing/internal/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   repositories.UserRepository
	sellerRepo repositories.SellerRepository
	adminRepo  repositories.AdminRepository
	jwtManager *utils.JWTManager
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	UserType int    `json:"user_type" binding:"required,oneof=1 2 3"` // 1=user, 2=seller, 3=admin
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Surname  string `json:"surname" binding:"required"`
	UserType int    `json:"user_type" binding:"required,oneof=1 2"` // Only user or seller can register
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	User         *UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uint            `json:"id"`
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Name     string          `json:"name"`
	Surname  string          `json:"surname"`
	UserType models.UserType `json:"user_type"`
}

func NewAuthService(
	userRepo repositories.UserRepository,
	sellerRepo repositories.SellerRepository,
	adminRepo repositories.AdminRepository,
	jwtManager *utils.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		sellerRepo: sellerRepo,
		adminRepo:  adminRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Register(req *RegisterRequest) (*TokenResponse, error) {
	// Validate input
	if !utils.ValidateEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	if valid, validationErrors := utils.ValidateUsername(req.Username); !valid {
		return nil, errors.New("username validation failed: " + validationErrors[0])
	}

	if valid, validationErrors := utils.ValidatePassword(req.Password); !valid {
		return nil, errors.New("password validation failed: " + validationErrors[0])
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Check if user already exists in any table
	if req.UserType == 1 { // User
		if existingUser, _ := s.userRepo.GetByEmail(req.Email); existingUser != nil {
			return nil, errors.New("user with this email already exists")
		}
		if existingUser, _ := s.userRepo.GetByUsername(req.Username); existingUser != nil {
			return nil, errors.New("user with this username already exists")
		}

		// Create user
		user := &models.User{
			Username:     utils.SanitizeString(req.Username),
			Email:        utils.SanitizeString(req.Email),
			PasswordHash: hashedPassword,
			Name:         utils.SanitizeString(req.Name),
			Surname:      utils.SanitizeString(req.Surname),
		}

		if err := s.userRepo.Create(user); err != nil {
			return nil, errors.New("failed to create user")
		}

		return s.generateTokenResponseForUser(user)

	} else if req.UserType == 2 { // Seller
		if existingSeller, _ := s.sellerRepo.GetByEmail(req.Email); existingSeller != nil {
			return nil, errors.New("seller with this email already exists")
		}
		if existingSeller, _ := s.sellerRepo.GetByUsername(req.Username); existingSeller != nil {
			return nil, errors.New("seller with this username already exists")
		}

		// Create seller
		seller := &models.Seller{
			Username:     utils.SanitizeString(req.Username),
			Email:        utils.SanitizeString(req.Email),
			PasswordHash: hashedPassword,
			Name:         utils.SanitizeString(req.Name),
			Surname:      utils.SanitizeString(req.Surname),
		}

		if err := s.sellerRepo.Create(seller); err != nil {
			return nil, errors.New("failed to create seller")
		}

		return s.generateTokenResponseForSeller(seller)
	}

	return nil, errors.New("invalid user type")
}

func (s *AuthService) Login(req *LoginRequest) (*TokenResponse, error) {
	switch req.UserType {
	case 1: // User
		user, err := s.userRepo.GetByEmail(req.Email)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("invalid email or password")
			}
			return nil, errors.New("failed to find user")
		}

		if !utils.CheckPassword(req.Password, user.PasswordHash) {
			return nil, errors.New("invalid email or password")
		}

		return s.generateTokenResponseForUser(user)

	case 2: // Seller
		seller, err := s.sellerRepo.GetByEmail(req.Email)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("invalid email or password")
			}
			return nil, errors.New("failed to find seller")
		}

		if !utils.CheckPassword(req.Password, seller.PasswordHash) {
			return nil, errors.New("invalid email or password")
		}

		return s.generateTokenResponseForSeller(seller)

	case 3: // Admin
		admin, err := s.adminRepo.GetByEmail(req.Email)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("invalid email or password")
			}
			return nil, errors.New("failed to find admin")
		}

		if !utils.CheckPassword(req.Password, admin.PasswordHash) {
			return nil, errors.New("invalid email or password")
		}

		return s.generateTokenResponseForAdmin(admin)

	default:
		return nil, errors.New("invalid user type")
	}
}

func (s *AuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Generate new tokens based on user type
	switch claims.UserType {
	case models.UserTypeUser:
		user, err := s.userRepo.GetByID(claims.UserID)
		if err != nil {
			return nil, errors.New("user not found")
		}
		return s.generateTokenResponseForUser(user)

	case models.UserTypeSeller:
		seller, err := s.sellerRepo.GetByID(claims.UserID)
		if err != nil {
			return nil, errors.New("seller not found")
		}
		return s.generateTokenResponseForSeller(seller)

	case models.UserTypeAdmin:
		admin, err := s.adminRepo.GetByID(claims.UserID)
		if err != nil {
			return nil, errors.New("admin not found")
		}
		return s.generateTokenResponseForAdmin(admin)

	default:
		return nil, errors.New("invalid user type in token")
	}
}

func (s *AuthService) generateTokenResponseForUser(user *models.User) (*TokenResponse, error) {
	userInfo := &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Surname:  user.Surname,
		UserType: models.UserTypeUser,
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Username, user.Email, models.UserTypeUser)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Username, user.Email, models.UserTypeUser)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Hour * 24 * 7), // 7 days in seconds
		User:         userInfo,
	}, nil
}

func (s *AuthService) generateTokenResponseForSeller(seller *models.Seller) (*TokenResponse, error) {
	userInfo := &UserInfo{
		ID:       seller.ID,
		Username: seller.Username,
		Email:    seller.Email,
		Name:     seller.Name,
		Surname:  seller.Surname,
		UserType: models.UserTypeSeller,
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(seller.ID, seller.Username, seller.Email, models.UserTypeSeller)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(seller.ID, seller.Username, seller.Email, models.UserTypeSeller)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Hour * 24 * 7), // 7 days in seconds
		User:         userInfo,
	}, nil
}

func (s *AuthService) generateTokenResponseForAdmin(admin *models.Admin) (*TokenResponse, error) {
	userInfo := &UserInfo{
		ID:       admin.ID,
		Username: admin.Username,
		Email:    admin.Email,
		Name:     admin.Name,
		Surname:  admin.Surname,
		UserType: models.UserTypeAdmin,
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(admin.ID, admin.Username, admin.Email, models.UserTypeAdmin)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(admin.ID, admin.Username, admin.Email, models.UserTypeAdmin)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Hour * 24 * 7), // 7 days in seconds
		User:         userInfo,
	}, nil
}
