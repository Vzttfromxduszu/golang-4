package dao

import (
	"context"

	"gorm.io/gorm"

	"mall/repository/db/model"
)

type AdminDao struct {
	*gorm.DB
}

func NewAdminDao(ctx context.Context) *AdminDao {
	return &AdminDao{NewDBClient(ctx)}
}

func NewAdminDaoByDB(db *gorm.DB) *AdminDao {
	return &AdminDao{db}
}

// GetAdminById 根据 id 获取用户
func (dao *AdminDao) GetAdminById(uId uint) (Admin *model.Admin, err error) {
	err = dao.DB.Model(&model.Admin{}).Where("id=?", uId).
		First(&Admin).Error
	return
}

// UpdateAdminById 根据 id 更新用户信息
func (dao *AdminDao) UpdateAdminById(uId uint, Admin *model.Admin) (err error) {
	return dao.DB.Model(&model.Admin{}).Where("id=?", uId).
		Updates(&Admin).Error
}

// ExistOrNotByAdminName 根据Adminname判断是否存在该名字
func (dao *AdminDao) ExistOrNotByAdminName(AdminName string) (Admin *model.Admin, exist bool, err error) {
	var count int64
	err = dao.DB.Model(&model.Admin{}).Where("admin_name=?", AdminName).Count(&count).Error
	if count == 0 {
		return Admin, false, err
	}
	err = dao.DB.Model(&model.Admin{}).Where("admin_name=?", AdminName).First(&Admin).Error
	if err != nil {
		return Admin, false, err
	}
	return Admin, true, nil
}

// CreateAdmin 创建用户
func (dao *AdminDao) CreateAdmin(Admin *model.Admin) error {
	return dao.DB.Model(&model.Admin{}).Create(&Admin).Error
}
