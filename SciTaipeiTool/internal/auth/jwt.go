package auth

import (
	"SciTaipeiTool/internal/dataprovider/models"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("test_secret_key") // 後續替換為config的密鑰

// Init 初始化 JWT 密鑰
func Init(configKey string) {
	jwtSecret = []byte(configKey)
	if len(jwtSecret) == 0 {
		jwtSecret = []byte(os.Getenv("JWT_SECRET")) // 從環境變數讀取
		if len(jwtSecret) == 0 {
			panic("JWT_SECRET 未設置")
		}
	}
}

// GenerateJWT 生成 JWT，回傳 Token 字串
func GenerateJWT(userID int) (string, error) {
	// 定義 Token 的 Claims
	claims := jwt.MapClaims{
		"UserId": userID,
		"Exp":    time.Now().Add(time.Minute * 30).Unix(), // // 設置 Access Token 過期時間為 30 分鐘
	}

	// 創建 Token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 簽名並生成 Token 字符串
	return accessToken.SignedString(jwtSecret)
}

// ValidateJWT 驗證 JWT，驗證成功回傳用戶 ID
func ValidateJWT(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("無效的簽名方法")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["UserId"].(float64); ok {
			return int(userID), nil
		}
	}
	return 0, errors.New("無效的 Token")
}

// ParseJWT 解析 JWT 並回傳用戶信息
func ParseJWT(tokenString string) (models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("無效的簽名方法")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return models.User{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := models.User{
			ID: int(claims["userID"].(float64)),
		}
		return user, nil
	}
	return models.User{}, errors.New("無效的 Token")
}
