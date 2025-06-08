// internal/services/user_service.go
package services

import (
	"errors"

	"eticketing/internal/models"
	"eticketing/internal/repositories"
	"eticketing/internal/utils"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo repositories.UserRepository
}

type UpdateProfileRequest struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Username string `json:"username"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetProfile(userID uint) (*UserInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to get user profile")
	}

	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Surname:  user.Surname,
		UserType: models.UserTypeUser,
	}, nil
}

func (s *UserService) UpdateProfile(userID uint, req *UpdateProfileRequest) (*UserInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Validate username if provided
	if req.Username != "" && req.Username != user.Username {
		if valid, validationErrors := utils.ValidateUsername(req.Username); !valid {
			return nil, errors.New("username validation failed: " + validationErrors[0])
		}

		// Check if username is already taken
		if existingUser, _ := s.userRepo.GetByUsername(req.Username); existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("username already taken")
		}
		user.Username = utils.SanitizeString(req.Username)
	}

	// Update other fields
	if req.Name != "" {
		user.Name = utils.SanitizeString(req.Name)
	}
	if req.Surname != "" {
		user.Surname = utils.SanitizeString(req.Surname)
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("failed to update profile")
	}

	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Surname:  user.Surname,
		UserType: models.UserTypeUser,
	}, nil
}

func (s *UserService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify current password
	if !utils.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if valid, validationErrors := utils.ValidatePassword(req.NewPassword); !valid {
		return errors.New("password validation failed: " + validationErrors[0])
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.PasswordHash = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *UserService) DeleteAccount(userID uint) error {
	// TODO: Add business logic to check if user can be deleted
	// For example, check if they have active tickets, pending transfers, etc.

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.userRepo.Delete(user.ID); err != nil {
		return errors.New("failed to delete account")
	}

	return nil
}
