package serializer

import (
	"mall/conf"
	"mall/consts"
	"mall/repository/db/model"
)

type Admin1 struct {
	ID        uint   `json:"id"`
	AdminName string `json:"admin_name"`
	Type      int    `json:"type"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	CreateAt  int64  `json:"create_at"`
}

// BuildAdmin 序列化用户
func BuildAdmin(Admin *model.Admin) *Admin1 {
	u := &Admin1{
		ID:        Admin.ID,
		AdminName: Admin.AdminName,
		CreateAt:  Admin.CreatedAt.Unix(),
	}

	if conf.UploadModel == consts.UploadModelOss {
		u.Avatar = Admin.Avatar
	}

	return u
}

func BuildAdmins(items []*model.Admin) (Admins []*Admin1) {
	for _, item := range items {
		Admin := BuildAdmin(item)
		Admins = append(Admins, Admin)
	}
	return Admins
}
