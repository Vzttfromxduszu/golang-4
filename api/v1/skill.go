package v1

import (
	"mall/consts"
	util "mall/pkg/utils"
	"mall/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ImportSkillGoods 处理导入秒杀商品文件的请求
func ImportSkillGoods(c *gin.Context) {
	var skillGoodsImport service.SkillGoodsImport
	// 从请求中获取上传的文件
	file, _, _ := c.Request.FormFile("file")
	// 将请求绑定到 skillGoodsImport 结构体
	if err := c.ShouldBind(&skillGoodsImport); err == nil {
		// 调用 Import 方法进行导入逻辑处理
		res := skillGoodsImport.Import(c.Request.Context(), file)
		c.JSON(consts.StatusOK, res)
	} else {
		// 如果绑定失败，返回错误信息
		c.JSON(consts.IlleageRequest, ErrorResponse(err))
		util.LogrusObj.Infoln(err, "ImportSkillGoods")
	}
}

// InitSkillGoods 初始化秒杀商品信息
func InitSkillGoods(c *gin.Context) {
	var skillGoods service.SkillGoodsService
	// 将请求绑定到 skillGoods 结构体
	if err := c.ShouldBind(&skillGoods); err == nil {
		// 调用 InitSkillGoods 方法进行初始化逻辑
		res := skillGoods.InitSkillGoods(c.Request.Context())
		c.JSON(consts.StatusOK, res)
	} else {
		// 如果绑定失败，返回错误信息
		c.JSON(consts.IlleageRequest, ErrorResponse(err))
		util.LogrusObj.Infoln(err, "InitSkillGoods")
	}
}

// SkillGoods 处理秒杀下单请求
func SkillGoods(c *gin.Context) {
	var skillGoods service.SkillGoodsService
	// 解析登录用户的 Token 获取 claim
	claim, _ := util.ParseToken(c.GetHeader("Authorization"))
	// 将请求绑定到 skillGoods 结构体
	if err := c.ShouldBind(&skillGoods); err == nil {
		// 调用 SkillGoods 方法进行秒杀下单逻辑
		res := skillGoods.SkillGoods(c.Request.Context(), claim.ID)
		c.JSON(consts.StatusOK, res)
	} else {
		// 如果绑定失败，返回错误信息
		c.JSON(consts.IlleageRequest, ErrorResponse(err))
		util.LogrusObj.Infoln(err, "SkillGoods")
	}
}

// CancelSkillOrder 处理取消秒杀订单请求
func CancelSkillOrder(c *gin.Context) {
	// 获取订单 ID
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid order ID"})
		return
	}

	// 创建取消订单服务
	service := service.CancelOrderService{
		OrderID: uint(orderID),
	}

	// 调用取消订单逻辑
	response := service.CancelOrder(c)
	c.JSON(200, response)
}
