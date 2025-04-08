package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"mime/multipart"
	"strconv"
	"time"

	xlsx "github.com/360EntSecGroup-Skylar/excelize"
	logging "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"mall/pkg/e"
	util "mall/pkg/utils"
	"mall/pkg/utils/track"
	"mall/repository/cache"
	"mall/repository/db/dao"
	model2 "mall/repository/db/model"
	"mall/repository/mq"
	"mall/serializer"

	"golang.org/x/time/rate"
)

// SkillGoodsImport 用于处理秒杀商品的导入
type SkillGoodsImport struct {
}

// 定义全局限流器
var limiter = rate.NewLimiter(10, 100) // 每秒生成 10 个令牌，桶容量为 100

// SkillGoodsService 用于处理秒杀商品的相关操作（如初始化、下单等）
type SkillGoodsService struct {
	SkillGoodsId uint   `json:"skill_goods_id" form:"skill_goods_id"` // 秒杀商品ID
	ProductId    uint   `json:"product_id" form:"product_id"`         // 商品ID
	BossId       uint   `json:"boss_id" form:"boss_id"`               // 商家ID
	AddressId    uint   `json:"address_id" form:"address_id"`         // 收货地址ID
	Key          string `json:"key" form:"key"`                       // 秒杀商品的唯一标识
}

// CancelOrderService 用于处理取消订单的逻辑
type CancelOrderService struct {
	OrderID uint `json:"order_id" form:"order_id"` // 订单ID
}

// Import 导入秒杀商品信息
func (service *SkillGoodsImport) Import(ctx context.Context, file multipart.File) serializer.Response {
	xlFile, err := xlsx.OpenReader(file) // 读取上传的Excel文件
	if err != nil {
		logging.Info(err)
	}
	code := e.SUCCESS
	rows := xlFile.GetRows("Sheet1") // 获取Excel中Sheet1的所有行
	length := len(rows[1:])          // 跳过表头，计算数据行数
	skillGoods := make([]*model2.SkillGoods, length)
	for index, colCell := range rows {
		if index == 0 { // 跳过表头
			continue
		}
		// 从Excel中解析商品信息
		pId, _ := strconv.Atoi(colCell[0])
		bId, _ := strconv.Atoi(colCell[1])
		num, _ := strconv.Atoi(colCell[3])
		money, _ := strconv.ParseFloat(colCell[4], 64)
		skillGood := &model2.SkillGoods{
			ProductId: uint(pId),
			BossId:    uint(bId),
			Title:     colCell[2],
			Money:     money,
			Num:       num,
		}
		skillGoods[index-1] = skillGood
	}
	// 批量将秒杀商品信息存入数据库
	err = dao.NewSkillGoodsDao(ctx).CreateByList(skillGoods)
	if err != nil {
		code = e.ERROR
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Data:   "上传失败",
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// InitSkillGoods 初始化秒杀商品信息，将数据库中的秒杀商品信息加载到Redis中
func (service *SkillGoodsService) InitSkillGoods(ctx context.Context) error {

	r := cache.RedisClient
	bloomFilterKey := "skill_goods_bloom_filter" // 布隆过滤器的 Redis 键

	// 初始化布隆过滤器（如果尚未初始化）
	_, err := r.Do("BF.RESERVE", bloomFilterKey, 0.01, 10000).Result() // 错误率 0.01，容量 10000
	if err != nil && err.Error() != "BUSYKEY Target key name already exists" {
		// 如果布隆过滤器已存在，忽略错误；否则返回错误
		fmt.Errorf("failed to initialize bloom filter: %v", err)
	}

	skillGoods, _ := dao.NewSkillGoodsDao(ctx).ListSkillGoods() // 从数据库中获取秒杀商品列表
	// 将秒杀商品信息存入Redis
	for i := range skillGoods {
		fmt.Println(*skillGoods[i])
		key := cache.SkillGooedKey(skillGoods[i].Id)
		r.HSet(key, "num", skillGoods[i].Num)
		r.HSet(key, "money", skillGoods[i].Money)
		// 将商品 ID 添加到布隆过滤器
		randomExpire := 30*time.Minute + time.Duration(rand.Intn(10))*time.Minute
		r.Expire(key, randomExpire)
		// 布隆过滤器的 Redis 键
		_, err := r.Do("BF.ADD", bloomFilterKey, skillGoods[i].Id).Result()
		if err != nil {
			util.LogrusObj.Errorln("err:", err)
		}
	}
	return nil
}

// SkillGoods 处理秒杀下单请求
func (service *SkillGoodsService) SkillGoods(ctx context.Context, uId uint) serializer.Response {
	// 创建追踪 Span
	span, _ := track.WithSpan(ctx, "SkillGoods")
	defer span.Finish()
	// 检查是否允许请求
	if !limiter.Allow() {
		return serializer.Response{
			Status: e.ErrorFrequencyLimit,
			Msg:    e.GetMsg(e.ErrorFrequencyLimit),
		}
	}
	bloomFilterKey := "skill_goods_bloom_filter" // 布隆过滤器的 Redis 键
	// 从Redis中获取秒杀商品的价格
	var mo float64
	key := cache.SkillGooedKey(service.SkillGoodsId) // 使用 SkillGooedKey 生成 Redis 键

	// 1. 优先从本地缓存获取价格（只缓存相对稳定的价格数据）
	if value, found := cache.LocalCache.Get(key + "_money"); found {
		mo = value.(float64)
	} else {
		// 2. 如果本地缓存不存在，则从Redis中获取
		mo, _ = cache.RedisClient.HGet(key, "money").Float64()
		exists, err := cache.RedisClient.Do("BF.EXISTS", bloomFilterKey, service.SkillGoodsId).Bool()
		if err != nil {
			return serializer.Response{
				Status: e.ERROR,
				Msg:    "布隆过滤器检查失败，请稍后再试",
			}
		}
		if !exists {
			return serializer.Response{
				Status: e.ERROR,
				Msg:    "商品不存在或已下架",
			}
		}
		// 3. 将价格数据存入本地缓存，设置 5 分钟过期时间
		cache.LocalCache.Set(key+"_money", mo, 5*time.Minute)
	}

	// 4. 构造秒杀商品的消息结构体
	sk := &model2.SkillGood2MQ{
		ProductId:   service.SkillGoodsId,
		BossId:      100 + service.SkillGoodsId,
		UserId:      uId,
		AddressId:   uId,
		Key:         key,
		Money:       mo,
		SkillGoodId: service.SkillGoodsId,
	}

	// 5. 调用加锁逻辑处理秒杀请求（通过Redis处理库存）
	err := RedissonSecKillGoods(sk)
	if err != nil {
		return serializer.Response{
			Status: e.ERROR,
			Msg:    err.Error(),
		}
	}

	return serializer.Response{
		Status: e.SUCCESS,
		Msg:    "秒杀成功",
	}
}

// RedissonSecKillGoods 加锁处理秒杀请求
func RedissonSecKillGoods(sk *model2.SkillGood2MQ) error {

	p := strconv.Itoa(int(sk.ProductId)) // 商品ID作为Redis的Key
	uuid := getUuid(p)                   // 生成唯一标识

	// Lua 脚本
	luaScript := `
        -- 参数
        local lockKey = KEYS[1]
        local stockKey = KEYS[2]
        local uuid = ARGV[1]
        local decrement = tonumber(ARGV[2])
		local productId = ARGV[3]

        -- 尝试获取锁
        if redis.call("SET", lockKey, uuid, "NX", "EX", 10) then
            -- 锁成功，扣减库存
            local stock = redis.call("HINCRBY", stockKey, "num", -decrement)
            if stock < 0 then
                -- 如果库存不足，回滚库存并返回错误
                redis.call("HINCRBY", stockKey, "num", decrement)
                redis.call("DEL", lockKey) -- 释放锁
                return -1
            end
            -- 返回成功
            return 1
        else
            -- 获取锁失败
            return 0
        end
    `

	// 执行 Lua 脚本
	result, err := cache.RedisClient.Eval(luaScript, []string{p, sk.Key}, uuid, 1, sk.ProductId, "skill_goods_bloom_filter").Result()
	if err != nil {
		util.LogrusObj.Info(err)
		return errors.New("lua script execution failed")
	}

	// 根据 Lua 脚本的返回值处理逻辑
	switch result.(int64) {
	case -1:
		fmt.Println("stock not enough")
		go SendMessageToUser(sk.UserId, "秒杀失败：库存不足")
		return errors.New("stock not enough")
	case 0:
		fmt.Println("get lock fail")
		go SendMessageToUser(sk.UserId, "秒杀失败：获取锁失败")
		return errors.New("get lock fail")
	case 1:
		fmt.Println("get lock success")
		go SendMessageToUser(sk.UserId, "秒杀成功！订单正在处理中")
	}

	// 将秒杀请求发送到消息队列
	err = SendSecKillGoodsToMQ(sk)
	if err != nil {
		go SendMessageToUser(sk.UserId, "秒杀成功，但订单创建失败，请联系客服")
		fmt.Println("send mq fail")
	}

	// 解锁 Lua 脚本
	unlockScript := `
        local lockKey = KEYS[1]
        local uuid = ARGV[1]

        -- 检查锁是否匹配
        if redis.call("GET", lockKey) == uuid then
            redis.call("DEL", lockKey)
            return 1
        else
            return 0
        end
    `

	// 执行解锁脚本
	unlockResult, err := cache.RedisClient.Eval(unlockScript, []string{p}, uuid).Result()
	if err != nil || unlockResult.(int64) == 0 {
		fmt.Println("unlock fail")
		return errors.New("unlock fail")
	}

	fmt.Println("unlock success")
	return nil
}

// SendSecKillGoodsToMQ 将秒杀请求发送到消息队列
func SendSecKillGoodsToMQ(sk *model2.SkillGood2MQ) error {
	if mq.RabbitMQ == nil {
		return errors.New("RabbitMQ connection is not initialized")
	}
	ch, err := mq.RabbitMQ.Channel() // 获取RabbitMQ的通道
	// 确保通道在函数结束时关闭
	// 声明队列
	if err != nil {
		err = errors.New("rabbitMQ err:" + err.Error())
		return err
	}
	defer ch.Close()
	// 声明队列
	q, err := ch.QueueDeclare("skill_goods", true, false, false, false, nil)
	if err != nil {
		err = errors.New("rabbitMQ err:" + err.Error())
		return err
	}
	// 将秒杀请求序列化为JSON
	body, _ := json.Marshal(sk)
	// 将消息发布到队列
	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})
	if err != nil {
		err = errors.New("rabbitMQ err:" + err.Error())
		return err
	}
	log.Printf("Sent %s", body)
	return nil
}
func ConsumeSecKillGoods() {
	ch, err := mq.RabbitMQ.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
		return
	}
	defer ch.Close() // 确保通道在函数结束时关闭

	msgs, err := ch.Consume(
		"skill_goods", // 队列名称
		"",            // 消费者名称
		true,          // 自动确认
		false,         // 独占
		false,         // 无需等待
		false,         // 其他参数
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	for msg := range msgs {
		var sk model2.SkillGood2MQ
		err := json.Unmarshal(msg.Body, &sk)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// 更新数据库库存
		err = UpdateDatabaseStock(sk)
		if err != nil {
			log.Printf("Failed to update database stock: %v", err)
		}
	}
}
func UpdateDatabaseStock(sk model2.SkillGood2MQ) error {
	skillgoodsdao := dao.NewSkillGoodsDao(context.Background())
	// 扣减库存
	err := skillgoodsdao.DecrementStock(sk.SkillGoodId, 1)
	if err != nil {
		return err
	}

	// 创建订单记录
	order := &model2.Order{
		ProductID: sk.ProductId,
		UserID:    sk.UserId,
		AddressID: sk.AddressId,
		Money:     sk.Money,
		Type:      1, // 订单状态
		// CreatedAt:   time.Now(),
	}
	dao1 := dao.NewOrderDaoByDB(skillgoodsdao.DB)
	err = dao1.CreateOrder(order)
	if err != nil {
		go SendMessageToUser(sk.UserId, "秒杀成功，但订单创建失败，请联系客服")
		return err
	}
	// 推送订单创建成功消息
	go SendMessageToUser(sk.UserId, "订单创建成功！请尽快完成支付")

	return nil
}

// getUuid 生成唯一标识符
func getUuid(gid string) string {
	codeLen := 8
	// 定义原始字符串
	rawStr := "jkwangagDGFHGSERKILMJHSNOPQR546413890_"
	// 定义一个缓冲区
	buf := make([]byte, 0, codeLen)
	b := bytes.NewBuffer(buf)
	// 随机生成字符
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	for rawStrLen := len(rawStr); codeLen > 0; codeLen-- {
		randNum := rand.Intn(rawStrLen)
		b.WriteByte(rawStr[randNum])
	}
	return b.String() + gid
}

// CancelOrder 取消订单并回滚库存
func (service *CancelOrderService) CancelOrder(ctx context.Context) serializer.Response {
	// 获取订单信息
	orderDao := dao.NewOrderDao(ctx)
	order, err := orderDao.GetOrderById(service.OrderID)
	if err != nil {
		go SendMessageToUser(order.UserID, "订单取消失败，无法找到订单，请联系客服")
		return serializer.Response{
			Status: e.ERROR,
			Msg:    "订单不存在",
		}
	}

	// 检查订单状态是否允许取消
	if order.Type != 1 { // 假设 1 表示未完成订单
		go SendMessageToUser(order.UserID, "订单取消失败，该订单无法取消，请联系客服")
		return serializer.Response{
			Status: e.ERROR,
			Msg:    "订单无法取消",
		}
	}

	// 回滚 Redis 中的库存
	stockKey := cache.SkillGooedKey(order.ProductID)
	_, err = cache.RedisClient.HIncrBy(stockKey, "num", 1).Result()
	if err != nil {
		return serializer.Response{
			Status: e.ERROR,
			Msg:    "回滚库存失败",
		}
	}

	// 更新数据库库存
	skillGoodsDao := dao.NewSkillGoodsDao(ctx)
	err = skillGoodsDao.IncrementStock(order.ProductID, 1)
	if err != nil {
		return serializer.Response{
			Status: e.ERROR,
			Msg:    "更新数据库库存失败",
		}
	}

	// 更新订单状态为已取消
	order.Type = 3 // 假设 3 表示已取消订单
	err = orderDao.UpdateOrder(order)
	if err != nil {
		go SendMessageToUser(order.UserID, "订单取消失败，请稍后重试")
		return serializer.Response{
			Status: e.ERROR,
			Msg:    "更新订单状态失败",
		}
	}
	go SendMessageToUser(order.UserID, "订单已取消，库存已回滚")
	return serializer.Response{
		Status: e.SUCCESS,
		Msg:    "订单已取消，库存已回滚",
	}
}
