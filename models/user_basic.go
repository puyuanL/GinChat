package models

import (
	"GinChat/utils"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserBasic struct {
	gorm.Model
	Name          string
	PassWord      string
	Phone         string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email         string `valid:"email"`
	Avatar        string //头像
	Identity      string
	ClientIp      string
	ClientPort    string
	Salt          string
	LoginTime     time.Time
	HeartbeatTime time.Time
	LoginOutTime  time.Time `gorm:"column:login_out_time" json:"login_out_time"`
	IsLogout      bool
	DeviceInfo    string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func GetUserList() []*UserBasic {
	data := make([]*UserBasic, 10)
	utils.DB.Find(&data)
	for _, v := range data {
		fmt.Println(v)
	}
	return data
}

func RefreshSQLToken(user UserBasic, token string) error {
	result := utils.DB.Model(&user).Where("id = ?", user.ID).Update("identity", token)
	return result.Error
}

// FindUserByName 根据Name查找某个用户
func FindUserByName(name string) (UserBasic, error) {
	user := UserBasic{}
	result := utils.DB.Where("name = ?", name).First(&user)
	// can't find this username
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return user, gorm.ErrRecordNotFound
		} else {
			panic(fmt.Sprintf("error: %v", result.Error))
		}
	}
	return user, nil
}

// FindUserByPhone 根据Phone查找某个用户
func FindUserByPhone(phone string) (UserBasic, error) {
	user := UserBasic{}
	result := utils.DB.Where("Phone = ?", phone).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return user, gorm.ErrRecordNotFound
		} else {
			panic(fmt.Sprintf("error: %v", result.Error))
		}
	}
	return user, nil
}

// FindUserByEmail 根据Email查找某个用户
func FindUserByEmail(email string) (UserBasic, error) {
	user := UserBasic{}
	result := utils.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return user, gorm.ErrRecordNotFound
		} else {
			panic(fmt.Sprintf("error: %v", result.Error))
		}
	}
	return user, nil
}

// FindByID 根据ID查找某个用户
func FindByID(id uint) UserBasic {
	user := UserBasic{}
	utils.DB.Where("id = ?", id).First(&user)
	return user
}

func CreateUser(user UserBasic) *gorm.DB {
	return utils.DB.Create(&user)
}
func DeleteUser(user UserBasic) *gorm.DB {
	return utils.DB.Delete(&user)
}
func UpdateUser(user UserBasic) *gorm.DB {
	return utils.DB.Model(&user).Updates(
		UserBasic{
			Name:     user.Name,
			PassWord: user.PassWord,
			Phone:    user.Phone,
			Email:    user.Email,
			Avatar:   user.Avatar,
		})
}
