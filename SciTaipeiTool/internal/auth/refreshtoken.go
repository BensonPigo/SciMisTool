package auth

import (
	"SciTaipeiTool/internal/dataprovider/db"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// GenerateRefreshToken 生成 Refresh Token字串
func GenerateRefreshToken(userID int) (string, error) {

	// 產生隨機數
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	tokenStr := base64.URLEncoding.EncodeToString(tokenBytes)

	return tokenStr, nil
}

// RefreshAccessToken 根據 Refresh Token ，生成新 Access Token，回傳 Access Token 字串
func RefreshAccessToken(refreshToken string) (string, error) {
	// 先判斷Refresh Token是哪一個UserID
	tokenRecord, err := db.GetRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("無效的 Refresh Token")
	}

	newAccessToken, err := GenerateJWT(tokenRecord.UserID)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}
