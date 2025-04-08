package v1

import (
	"mall/consts"
	util "mall/pkg/utils"
	"mall/service"

	gin "github.com/gin-gonic/gin"
)

// AdminRegister 管理员注册接口
func AdminRegister(c *gin.Context) {
	var AdminRegisterService service.AdminService //相当于创建了一个AdminRegisterService对象，调用这个对象中的Register方法。
	if err := c.ShouldBind(&AdminRegisterService); err == nil {
		res := AdminRegisterService.Register(c.Request.Context())
		c.JSON(consts.StatusOK, res)
	} else {
		c.JSON(consts.IlleageRequest, ErrorResponse(err))
		util.LogrusObj.Infoln(err)
	}
}

// AdminLogin 管理员登陆接口
func AdminLogin(c *gin.Context) {
	var AdminLoginService service.AdminService
	if err := c.ShouldBind(&AdminLoginService); err == nil {
		res := AdminLoginService.AdminLogin(c.Request.Context())
		c.JSON(consts.StatusOK, res)
	} else {
		c.JSON(consts.IlleageRequest, ErrorResponse(err))
		util.LogrusObj.Infoln(err)
	}
}
