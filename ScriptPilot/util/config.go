package util

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/viper" // 相對於$GOPATH/pkg/mod的路徑
)

var (
	ProjectRootPath = getProjectRootPath() // 初始化根目錄
)

func getProjectRootPath() string {
	// 1. 嘗試使用當前工作目錄
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("無法取得工作目錄: %v", err))
	}

	// 2. 確認工作目錄是否包含 config 資料夾
	configPath := filepath.Join(wd, "config")
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("使用工作目錄作為根目錄:", wd)
		return wd
	}

	// 3. 如果工作目錄不包含 config，則切換到執行檔所在目錄
	exePath, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("無法取得執行檔路徑: %v", err))
	}
	exeDir := filepath.Dir(exePath)
	fmt.Println("使用執行檔所在目錄作為根目錄:", exeDir)
	return exeDir
}

// Viper可以解析JSON、TOML、YAML、HCL、INI、ENV等格式的設定檔
func CreateConfig(file string, env string) *viper.Viper {
	config := viper.New()
	configPath := path.Join(ProjectRootPath, "config")
	configFile := path.Join(configPath, file+"."+env+".yaml")

	fmt.Println("正在嘗試讀取配置文件:", configFile) // 用於排錯
	config.AddConfigPath(configPath)
	config.SetConfigName(file + "." + env) // 文件名
	config.SetConfigType("yaml")           // 文件類型

	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("找不到配置文件: %s", configFile))
		} else {
			panic(fmt.Errorf("解析配置文件 %s 發生錯誤: %v", configFile, err))
		}
	}
	return config
}
