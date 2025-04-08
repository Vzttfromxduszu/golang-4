package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	api "mall/api/v1"
	"mall/middleware"
)

// 路由配置
func NewRouter() *gin.Engine {
	r := gin.Default()
	store := cookie.NewStore([]byte("something-very-secret"))
	r.Use(middleware.Cors(), middleware.Jaeger())
	r.Use(sessions.Sessions("mysession", store))
	r.StaticFS("/static", http.Dir("./static"))
	v1 := r.Group("api/v1")
	{

		v1.GET("ping", func(c *gin.Context) {
			c.JSON(200, "success")
		})

		// 用户操作
		v1.POST("user/register", api.UserRegister)
		v1.POST("user/login", api.UserLogin)
		v1.POST("admin/register", api.AdminRegister)
		v1.POST("admin/adminlogin", api.AdminLogin)

		// 商品操作
		v1.GET("products", api.ListProducts)
		v1.GET("product/:id", api.ShowProduct)
		v1.POST("products", api.SearchProducts)
		v1.GET("imgs/:id", api.ListProductImg)   // 商品图片
		v1.GET("categories", api.ListCategories) // 商品分类
		v1.GET("carousels", api.ListCarousels)   // 轮播图
		// v1.GET("product/:id",api.SearchProduct)
		// WebSocket 路由
		v1.GET("ws", api.HandleWebSocket)

		Sellerauthed := v1.Group("/") // 需要登陆保护
		Sellerauthed.Use(middleware.JWTSeller())
		{
			Sellerauthed.POST("product", api.CreateProduct)
			Sellerauthed.PUT("product/:id", api.UpdateProduct)
			Sellerauthed.DELETE("product/:id", api.DeleteProduct)

			// 秒杀专场
			Sellerauthed.POST("import_skill_goods", api.ImportSkillGoods)
			Sellerauthed.POST("init_skill_goods", api.InitSkillGoods)
		}
		Userauthed := v1.Group("/") // 需要登陆保护
		Userauthed.Use(middleware.JWT())
		{

			// 用户操作
			Userauthed.PUT("user", api.UserUpdate)
			Userauthed.POST("user/sending-email", api.SendEmail)
			Userauthed.POST("user/valid-email", api.ValidEmail)
			Userauthed.POST("avatar", api.UploadAvatar)       // 上传头像
			Userauthed.POST("user/search_user", api.UserInfo) // 用户信息

			// 商品操作
			// authed.POST("product", api.CreateProduct)
			// authed.PUT("product/:id", api.UpdateProduct)
			// authed.DELETE("product/:id", api.DeleteProduct)
			// 商品分类操作
			Userauthed.POST("categories", api.CreateCategory)
			// 收藏夹
			Userauthed.GET("favorites", api.ShowFavorites)
			Userauthed.POST("favorites", api.CreateFavorite)
			Userauthed.DELETE("favorites/:id", api.DeleteFavorite)

			// 订单操作
			Userauthed.POST("orders", api.CreateOrder)
			Userauthed.GET("orders", api.ListOrders)
			Userauthed.GET("orders/:id", api.ShowOrder)
			Userauthed.DELETE("orders/:id", api.DeleteOrder)

			// 购物车
			Userauthed.POST("carts", api.CreateCart)
			Userauthed.GET("carts", api.ShowCarts)
			Userauthed.PUT("carts/:id", api.UpdateCart) // 购物车id
			Userauthed.DELETE("carts/:id", api.DeleteCart)

			// 收获地址操作
			Userauthed.POST("addresses", api.CreateAddress)
			Userauthed.GET("addresses/:id", api.GetAddress)
			Userauthed.GET("addresses", api.ListAddress)
			Userauthed.PUT("addresses/:id", api.UpdateAddress)
			Userauthed.DELETE("addresses/:id", api.DeleteAddress)
			Userauthed.GET("addresses/byname/:name", api.GetAddressByName)

			// 支付功能
			Userauthed.POST("paydown", api.OrderPay)

			// 显示金额
			Userauthed.POST("money", api.ShowMoney)

			// 秒杀下单
			Userauthed.POST("skill_goods", api.SkillGoods)
			// 秒杀订单操作
			Userauthed.DELETE("skill_orders/:id", api.CancelSkillOrder)
		}
	}
	return r
}
