package models

import (
	"GinChat/utils"
	"errors"

	"gorm.io/gorm"
)

// Contact 人员关系
type Contact struct {
	gorm.Model
	OwnerId  uint //谁的关系信息
	TargetId uint //对应的谁 /群 ID
	Type     int  //对应的类型  1好友  2群  3xx
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}

func SearchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type=1", userId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users
}

// AddFriend 添加好友
// @Param userId 自己的ID
// @Param targetName	好友的ID
func AddFriend(userId uint, targetName string) (int, string) {

	if targetName != "" {
		targetUser, err := FindUserByName(targetName)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			if targetUser.ID == userId {
				return -1, "不能加自己"
			}
			contact0 := Contact{}
			utils.DB.Where("owner_id =? and target_id =? and type=1", userId, targetUser.ID).Find(&contact0)
			if contact0.ID != 0 {
				return -1, "不能重复添加"
			}

			contacts := []Contact{
				{OwnerId: userId, TargetId: targetUser.ID, Type: 1},
				{OwnerId: targetUser.ID, TargetId: userId, Type: 1},
			}

			// Method 1 -- Transaction
			////事务一旦开始，不论什么异常最终都会
			//tx := utils.DB.Begin()
			//defer func() {
			//	if r := recover(); r != nil {
			//		tx.Rollback()
			//	}
			//}()
			//if err := utils.DB.Create(&contacts[0]).Error; err != nil {
			//	tx.Rollback()
			//	return -1, "添加好友失败"
			//}
			//if err := utils.DB.Create(&contacts[1]).Error; err != nil {
			//	tx.Rollback()
			//	return -1, "添加好友失败"
			//}
			//tx.Commit()

			// Method 2 -- Insert Multi Data
			if err := utils.DB.Create(&contacts).Error; err != nil {
				return -1, "添加好友失败：" + err.Error()
			}

			return 0, "添加好友成功"
		}
		return -1, "没有找到此用户"
	}
	return -1, "好友ID不能为空"
}

func SearchUserByGroupId(communityId uint) []uint {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("target_id = ? and type=2", communityId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, v.OwnerId)
	}
	return objIds
}
