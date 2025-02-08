package db

//封裝與 User 相關的操作

import (
	"errors"
	"fmt"

	"SciTaipeiTool/internal/dataprovider/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterUser 註冊使用者
func RegisterUser(user models.User) error {
	db := GetDB() // 獲取資料庫連接

	// 檢查是否已存在
	result := db.Where("email = ?", user.Email).First(&user)
	if result.RowsAffected > 0 {
		return errors.New("使用者已存在")
	}

	// 雜湊密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密碼雜湊失敗: %v", err)
	}

	// 儲存到資料庫
	user = models.User{Email: user.Email, Username: user.Username, Password: string(hashedPassword)}
	err = db.Create(&user).Error
	if err != nil {
		return fmt.Errorf("儲存使用者失敗: %v", err)
	}

	return nil
}

// AuthenticateUser 驗證使用者
func AuthenticateUser(inputUser *models.User) (*models.User, error) {
	db := GetDB() // 獲取資料庫連接

	// 查詢使用者
	providedPassword := inputUser.Password // 先把外面傳進來的密碼存下來

	var dbUser models.User
	result := db.Where("email = ?", inputUser.Email).First(&dbUser) // 取得user
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("使用者不存在")
		}
		return nil, fmt.Errorf("查詢使用者失敗: %v", result.Error)
	}

	// 驗證密碼
	err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(providedPassword))
	if err != nil {
		return nil, errors.New("密碼錯誤")
	}

	return &dbUser, nil
}

// ResetPassword 密碼重設
func ResetPassword(user models.User) error {
	db := GetDB() // 獲取資料庫連接

	newPwd := user.Password
	// 檢查是否已存在
	result := db.Where("email = ?", user.Email).First(&user)
	if result.RowsAffected == 0 {
		return errors.New("使用者不存在")
	}

	// 雜湊密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密碼雜湊失敗: %v", err)
	}

	// 更新到資料庫
	user.Password = string(hashedPassword)
	err = db.Save(&user).Error
	if err != nil {
		return fmt.Errorf("更新使用者失敗: %v", err)
	}

	return nil
}
