package dao

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"mall/repository/db/model"
)

type SkillGoodsDao struct {
	*gorm.DB
}

func NewSkillGoodsDao(ctx context.Context) *SkillGoodsDao {
	return &SkillGoodsDao{NewDBClient(ctx)}
}

func (dao *SkillGoodsDao) Create(in *model.SkillGoods) error {
	return dao.Model(&model.SkillGoods{}).Create(&in).Error
}

func (dao *SkillGoodsDao) CreateByList(in []*model.SkillGoods) error {
	return dao.Model(&model.SkillGoods{}).Create(&in).Error
}

func (dao *SkillGoodsDao) ListSkillGoods() (resp []*model.SkillGoods, err error) {
	err = dao.Model(&model.SkillGoods{}).Where("num > 0").Find(&resp).Error
	return
}

func (dao *SkillGoodsDao) DecrementStock(SkillGoodId uint, quantity uint) (err error) {
	result := dao.Model(&model.SkillGoods{}).
		Where("id = ? AND num >= ?", SkillGoodId, quantity).
		Update("num", gorm.Expr("num - ?", quantity))
	if result.RowsAffected == 0 {
		return errors.New("stock not enough")
	}
	return nil
}
