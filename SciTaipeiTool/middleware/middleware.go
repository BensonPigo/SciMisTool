package middleware

import (
	"net/http"
	"strings"

	"SciTaipeiTool/internal/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 驗證 JWT 的中間件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 Authorization 標頭提取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供有效的 Token"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 驗證 Token
		userID, err := auth.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "無效的 Token"})
			c.Abort()
			return
		}

		// 將 userID 保存到上下文
		c.Set("UserId", userID)

		c.Next()
	}
}
