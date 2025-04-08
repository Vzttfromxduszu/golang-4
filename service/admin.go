package service

import (
	"context"
	"mall/conf"
	"mall/consts"
	"mall/pkg/e"
	util "mall/pkg/utils"
	dao2 "mall/repository/db/dao"
	model2 "mall/repository/db/model"
	"mall/serializer"

	logging "github.com/sirupsen/logrus"
)

// AdminService 管理用户服务
type AdminService struct {
	AdminName string `form:"Admin_name" json:"Admin_name"`
	NickName  string `form:"nick_name" json:"nick_name"`
	Password  string `form:"password" json:"password"`
	Key       string `form:"key" json:"key"` // 前端进行判断
}

func (service AdminService) Register(ctx context.Context) serializer.Response {
	var Admin *model2.Admin
	code := e.SUCCESS
	if service.Key == "" || len(service.Key) != 16 {
		// 为什么是16位？ 因为AES加密算法要求密钥长度为16、24或32字节
		code = e.ERROR
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Data:   "密钥长度不足",
		}
	}
	util.Encrypt.SetKey(service.Key)
	AdminDao := dao2.NewAdminDao(ctx)
	_, exist, err := AdminDao.ExistOrNotByAdminName(service.AdminName)
	if err != nil {
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	if exist {
		code = e.ErrorExistAdmin
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	Admin = &model2.Admin{
		AdminName: service.AdminName,
		NickName:  service.NickName,
	}
	// 加密密码
	if err = Admin.SetPassword(service.Password); err != nil {
		logging.Info(err)
		code = e.ErrorFailEncryption
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	if conf.UploadModel == consts.UploadModelOss {
		Admin.Avatar = "http://q1.qlogo.cn/g?b=qq&nk=294350394&s=640"
	} else {
		Admin.Avatar = "avatar.JPG"
	}
	// 创建用户
	err = AdminDao.CreateAdmin(Admin)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
		Data:   serializer.BuildAdmin(Admin).AdminName,
	}

}

// Login 用户登陆函数
func (service *AdminService) Login(ctx context.Context) serializer.Response {
	var Admin *model2.Admin
	code := e.SUCCESS
	AdminDao := dao2.NewAdminDao(ctx)
	Admin, exist, err := AdminDao.ExistOrNotByAdminName(service.AdminName)
	if !exist { // 如果查询不到，返回相应的错误
		logging.Info(err)
		code = e.ErrorAdminNotFound
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	if Admin.CheckPassword(service.Password) == false {
		code = e.ErrorNotCompare
		util.LogrusObj.Infoln("err", err)
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	token, err := util.GenerateToken(Admin.ID, service.AdminName, 0)
	if err != nil {
		logging.Info(err)
		code = e.ErrorAuthToken
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	return serializer.Response{
		Status: code,
		Data:   serializer.TokenData{Admin: serializer.BuildAdmin(Admin), Token: token},
		Msg:    e.GetMsg(code),
	}
}

// Login 用户登陆函数
func (service *AdminService) AdminLogin(ctx context.Context) serializer.Response {
	var Admin *model2.Admin
	code := e.SUCCESS
	AdminDao := dao2.NewAdminDao(ctx)
	Admin, exist, err := AdminDao.ExistOrNotByAdminName(service.AdminName)
	if !exist { // 如果查询不到，返回相应的错误
		logging.Info(err)
		code = e.ErrorAdminNotFound
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	if Admin.CheckPassword(service.Password) == false {
		code = e.ErrorNotCompare
		util.LogrusObj.Infoln("err", err)
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	token, err := util.GenerateToken(Admin.ID, service.AdminName, 1)
	if err != nil {
		logging.Info(err)
		code = e.ErrorAuthToken
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}
	return serializer.Response{
		Status: code,
		Data:   serializer.TokenData{Admin: serializer.BuildAdmin(Admin), Token: token},
		Msg:    e.GetMsg(code),
	}
}
