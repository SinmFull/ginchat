package models

import (
	"ginchat/utils"

	"gorm.io/gorm"
)

type Community struct {
	gorm.Model
	Name    string
	OwnerId uint
	Img     string
	Desc    string
}

func FindCommunityByID(id uint) Community {
	community := Community{}
	utils.DB.Where("id = ?", id).First(&community)
	return community
}

func CreateCommunity(community Community) (int, string) {
	tx := utils.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if len(community.Name) == 0 {
		return -1, "群名称不能为空"
	}
	if community.OwnerId == 0 {
		return -1, "请先登录"
	}
	if err := utils.DB.Create(&community).Error; err != nil {
		tx.Rollback()
		return -1, "建群失败"
	}
	contact := Contact{}
	contact.OwnerId = community.OwnerId
	contact.TargetId = community.ID
	contact.Type = 2
	if err := utils.DB.Create(&contact).Error; err != nil {
		tx.Rollback()
		return -1, "添加群关系失败"
	}
	tx.Commit()
	return 0, "建群成功"
}

func LoadCommunity(ownerId uint) ([]*Community, string) {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type = 2", ownerId).Find(&contacts)

	for _, v := range contacts {
		objIds = append(objIds, uint64(v.TargetId))
	}
	data := make([]*Community, 10)
	utils.DB.Where("id in ?", objIds).Find(&data)
	return data, "查询成功"
}

func JoinGroup(userId uint, comId string) (int, string) {
	//fmt.Println("user>>", userId, "   tarID>>>>>", targetId)
	contact := Contact{}
	contact.OwnerId = userId
	contact.Type = 2

	community := Community{}
	utils.DB.Where("id = ? or name = ?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没有找到群"
	}
	utils.DB.Where("owner_id = ? and target_id = ? and type = 2", userId, community.ID).Find(&contact)
	if !contact.CreatedAt.IsZero() {
		return -1, "已加过该群"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功"
	}
}
