package handler

import (
	"SciTaipeiTool/internal/auth"
	"SciTaipeiTool/internal/dataprovider/db"
	"SciTaipeiTool/internal/dataprovider/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler 是 HTTP 請求處理器
type LoginHandler struct {
	DatabaseName string // 暫時用不到，但之後紀錄log可以用到
}

// Login 處理用戶登入邏輯
func (lh *LoginHandler) Login(ctx *gin.Context) {

	// post body取JSON
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		log.Println("JSON 參數錯誤:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "請提供正確的 JSON 格式",
		})
		return
	}

	// 檢查Email和密碼是否為空
	if user.Email == "" || user.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Email或密碼不得為空",
		})
		return
	}

	// 驗證使用者
	dbUser, err := db.AuthenticateUser(&user)
	if err != nil {
		log.Println("登入失敗:", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "登入失敗，Email或密碼錯誤",
		})
		return
	}

	// 產生 accessToken
	accessToken, err := auth.GenerateJWT(dbUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "生成 Token 失敗"})
		return
	}

	// _, err = checkRefreshToken(ctx, dbUser.ID)
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"message": "生成 Token 失敗"})
	// 	return
	// }
	// 返回給前端
	ctx.JSON(http.StatusOK, gin.H{
		"message":     "登入成功",
		"accessToken": accessToken,
	})
}

func checkRefreshToken(ctx *gin.Context, userId int) (string, error) {

	// 從 Cookie 中獲取 Refresh Token
	var createRefreshToken = false
	refreshToken := ctx.GetString("RefreshToken")
	// _, err := ctx.Cookie("refreshToken")
	if refreshToken != "" {
		// 找到 Refresh Token，驗證效期，不合法則重塞
		_, err := db.ValidateRefreshToken(refreshToken)
		if err != nil {
			err := db.DeleteRefreshToken(userId)
			if err != nil {
				return "", err
			}
			createRefreshToken = true
		}
	} else {
		// 未找到 Refresh Token
		createRefreshToken = true
	}

	var refreshTokenStr string
	if createRefreshToken {
		// 產生RefreshToken
		refreshTokenStr, err := auth.GenerateRefreshToken(userId)
		if err != nil {
			return "", err
		}

		// 寫入DB
		newRefreshToken := models.RefreshToken{
			Token:     refreshTokenStr,
			UserID:    userId,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 設置過期時間為 7 天
		}
		err = db.CreateRefreshToken(newRefreshToken)
		if err != nil {
			return "", err
		}
	}
	// 設定 Refresh Token 為 HttpOnly Cookie
	// ctx.SetCookie("refreshToken", refreshTokenStr, 7*24*60*60, "/", "", false, true)
	ctx.Set("RefreshToken", refreshTokenStr)
	return refreshTokenStr, nil
}

// Logout 刪除 Refresh Token，確保 Refresh Token 無效
func (lh *LoginHandler) Logout(ctx *gin.Context) {

	// userId, exists := ctx.Get("UserId")

	// if !exists {

	// }

	// aaa, _ := ctx.Get("RefreshToken")
	// // 從 Cookie 中獲取 Refresh Token
	// // refreshTokenStr := ctx.GetString("RefreshToken")
	// refreshTokenStr := aaa.(string)
	// // _, err := ctx.Cookie("refreshToken")
	// if refreshTokenStr == "" {
	// 	log.Println("未找到 Refresh Token:")
	// 	ctx.JSON(http.StatusUnauthorized, gin.H{"message": "用戶未登入或已登出"})
	// 	return
	// }

	// // 刪除DB中的 Refresh Token
	// if err := db.DeleteRefreshToken(userId.(int)); err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"message": "登出失敗，無法刪除 Refresh Token"})
	// 	return
	// }

	// 移除 HttpOnly Cookie
	// ctx.SetCookie("refreshToken", "", -1, "/", "", false, true)
	// ctx.Set("RefreshToken", "")

	ctx.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// Register 處理用戶註冊邏輯
func (lh *LoginHandler) Register(ctx *gin.Context) {

	// post body取JSON
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		log.Println("JSON 參數錯誤:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "請提供正確的 JSON 格式",
		})
		return
	}

	// 檢查是否為空
	if user.Email == "" || user.Username == "" || user.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Email、Username、密碼不得為空",
		})
		return
	}

	// 註冊使用者
	if err := db.RegisterUser(user); err != nil {
		log.Println("註冊失敗:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "註冊失敗，Email已註冊過",
		})
		return
	}

	// 註冊成功
	log.Println("註冊成功!")
	ctx.JSON(http.StatusOK, gin.H{
		"message": "註冊成功",
	})
}

// ResetPassword 處理用戶重設密碼邏輯
func (lh *LoginHandler) ResetPassword(ctx *gin.Context) {

	// post body取JSON
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		log.Println("JSON 參數錯誤:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "請提供正確的 JSON 格式",
		})
		return
	}

	// 檢查Email和密碼是否為空
	if user.Email == "" || user.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "信箱或密碼不得為空",
		})
		return
	}

	// UPDATE
	if err := db.ResetPassword(user); err != nil {
		log.Println("Password修改失敗:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Password修改失敗",
		})
		return
	}

	// 註冊成功
	log.Println("Password修改成功!")
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Password修改成功",
	})
}

// RefreshToken 根據Refresh Token，刷新 Access Token
func (lh *LoginHandler) RefreshToken(ctx *gin.Context) {
	// post body取JSON
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		log.Println("JSON 參數錯誤:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "請提供正確的 JSON 格式",
		})
		return
	}
	refreshTokenStr, err := checkRefreshToken(ctx, user.ID)

	if err != nil {
		log.Println("刷新 Token 失敗:", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "無效的 Refresh Token"})
		return
	}

	// 生成新的 Access Token
	newAccessToken, err := auth.RefreshAccessToken(refreshTokenStr)
	if err != nil {
		log.Println("刷新 Token 失敗:", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "無效的 Refresh Token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": newAccessToken,
	})
}
