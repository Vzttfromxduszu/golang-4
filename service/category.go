package service

import (
	"context"

	logging "github.com/sirupsen/logrus"

	"mall/pkg/e"
	"mall/repository/db/dao"
	"mall/repository/db/model"
	"mall/serializer"
)

type ListCategoriesService struct {
}

type CreateCategoryService struct {
	CategoryName string `form:"category_name" json:"category_name"`
}

func (service *ListCategoriesService) List(ctx context.Context) serializer.Response {
	code := e.SUCCESS
	categoryDao := dao.NewCategoryDao(ctx)
	categories, err := categoryDao.ListCategory()
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
		Data:   serializer.BuildCategories(categories),
	}
}

func (service *CreateCategoryService) Create(ctx context.Context) serializer.Response {
	code := e.SUCCESS
	var category *model.Category
	category = &model.Category{
		CategoryName: service.CategoryName,
	}
	categoryDao := dao.NewCategoryDao(ctx)
	err := categoryDao.CreateCategory(category)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}
