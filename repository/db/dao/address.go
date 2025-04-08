package dao

import (
	"context"

	"gorm.io/gorm"

	"mall/repository/db/model"
)

type AddressDao struct {
	*gorm.DB
}

func NewAddressDao(ctx context.Context) *AddressDao {
	return &AddressDao{NewDBClient(ctx)}
}

func NewAddressDaoByDB(db *gorm.DB) *AddressDao {
	return &AddressDao{db}
}

// GetAddressByAid 根据 Address Id 获取 Address
func (dao *AddressDao) GetAddressByAid(aId uint) (address *model.Address, err error) {
	err = dao.DB.Model(&model.Address{}).
		Where("id = ?", aId).First(&address).
		Error
	return
}

// ListAddressByUid 根据 User Id 获取User
func (dao *AddressDao) ListAddressByUid(uid uint) (addressList []*model.Address, err error) {
	err = dao.DB.Model(&model.Address{}).
		Where("user_id=?", uid).Order("created_at desc").
		Find(&addressList).Error
	return
}

// ListAddressByname 根据 User name 获取User
func (dao *AddressDao) ListAddressByName(name string) (addressList []*model.Address, err error) {
	var uid uint
	err = dao.DB.Model(&model.User{}).Select("id").Where("user_name=?", name).Scan(&uid).Error
	if err != nil {
		return nil, err
	}

	err = dao.DB.Model(&model.Address{}).
		Where("user_id=?", uid).Order("created_at desc").
		Find(&addressList).Error
	return addressList, err
}

// CreateAddress 创建地址
func (dao *AddressDao) CreateAddress(address *model.Address) (err error) {
	err = dao.DB.Model(&model.Address{}).Create(&address).Error
	return
}

// DeleteAddressById 根据 id 删除地址
func (dao *AddressDao) DeleteAddressById(aId uint) (err error) {
	err = dao.DB.Where("id=?", aId).Delete(&model.Address{}).Error
	return
}

// UpdateAddressById 通过 id 修改地址信息
func (dao *AddressDao) UpdateAddressById(aId uint, address *model.Address) (err error) {
	err = dao.DB.Model(&model.Address{}).
		Where("id=?", aId).Updates(address).Error
	return
}
