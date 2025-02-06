package db

import (
	"errors"
	"fmt"
	"time"

	"SciTaipeiTool/internal/dataprovider/models"

	"gorm.io/gorm"
)

// GetRefreshToken 查詢 Refresh Token
func GetRefreshToken(token string) (*models.RefreshToken, error) {
	db := GetDB() // 獲取資料庫連接

	var refreshToken models.RefreshToken
	result := db.Where("token = ?", token).First(&refreshToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token 不存在")
		}
		return nil, errors.New("查詢 refresh token 失敗")
	}

	return &refreshToken, nil
}

// GetRefreshTokenByUserID 根據用戶 ID 查詢 Refresh Token
func GetRefreshTokenByUserID(userID int) (string, error) {
	db := GetDB() // 獲取資料庫連接

	var refreshToken models.RefreshToken
	result := db.Where("userid = ?", userID).First(&refreshToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", errors.New("refresh token不存在")
		}
		return "", errors.New("查詢 refresh token 失敗")
	}

	return refreshToken.Token, nil
}

// ValidateRefreshToken 驗證 Refresh Token
func ValidateRefreshToken(token string) (*models.RefreshToken, error) {
	refreshToken, err := GetRefreshToken(token)
	if err != nil {
		return refreshToken, err
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return refreshToken, errors.New("refresh token 已過期")
	}

	return refreshToken, nil
}

// CreateRefreshToken 新增 Refresh Token
func CreateRefreshToken(refreshToken models.RefreshToken) error {
	db := GetDB() // 獲取資料庫連接

	if err := db.Create(&refreshToken).Error; err != nil {
		return fmt.Errorf("儲存 Refresh Token 失敗: %v", err)
	}

	return nil
}

// DeleteRefreshToken 刪除Refresh Token
func DeleteRefreshToken(userID int) error {
	db := GetDB()
	if err := db.Delete(&models.RefreshToken{}, "UserID = ?", userID).Error; err != nil {
		return fmt.Errorf("刪除 refresh token 失敗: %v", err)
	}
	return nil
}
