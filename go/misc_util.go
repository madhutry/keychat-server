package friezechat

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func InitConfig() {
	home, _ := os.UserHomeDir()
	if os.Getenv("ENVIRONMENT") == "DEV" {
		fmt.Println("loafing dev env")
		viper.SetConfigName("server")
		viper.SetConfigType("json")
		viper.AddConfigPath(filepath.Dir(home + "/"))
		viper.ReadInConfig()
	} else {
		viper.AutomaticEnv()
	}

}

func GetDBUrl() string {
	return viper.GetString("DB_URL")
}
func GetMatrixServerUrl() string {
	return viper.GetString("MATRIX_URL")
}
func GetFriezeChatAPIUrl() string {
	return viper.GetString("FRIEZE_CHAT_API_HOST")
}
func GetMatrixAdminCode() string {
	return viper.GetString("MATRIX_ADMIN_ACCESS_CODE")
}
func GetMatrixAdminUserid() string {
	return viper.GetString("MATRIX_ADMIN_USERID")
}
func GetS3AccessKey() string {
	return viper.GetString("AWS_ACCESS_KEY_ID")
}

func GetS3SecretKey() string {
	return viper.GetString("AWS_SECRET_ACCESS_KEY")
}
func GetS3ParentFolder() string {
	return viper.GetString("S3_FOLDER_NAME")
}

func GetS3ObjectBucket() string {
	return viper.GetString("OBJECT_BUCKET")
}
func GetFCMServerCode() string {
	return viper.GetString("FCM_SERVER_CODE")
}
