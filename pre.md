我正在用 Go 开发一个秒杀系统学习项目，前期设计已完成，
现在需要你帮我逐层实现代码。以下是完整设计文档，请先确认
理解后，等我指令再开始写代码。

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【一、项目定位】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
学习项目，覆盖高并发秒杀完整链路。
硬约束：零超卖 · 每人限购一件 · 10,000 QPS 不宕机 · 请求幂等

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【二、技术栈】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Web 框架：Gin
ORM：GORM v2
缓存：Redis 7（go-redis/v9）
消息队列：RocketMQ 5（官方 Go SDK）
数据库：MySQL 8
限流：golang.org/x/time/rate（令牌桶）
日志：Zap（uber-go）
配置：Viper + yaml
ID 生成：雪花算法（自实现）
依赖注入：手动构造（main.go 组装，不用 wire）

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【三、项目目录结构】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
seckill/
├── cmd/server/main.go          # 入口：组装依赖，启动 HTTP server
├── internal/
│   ├── handler/                # HTTP 层：参数绑定、响应封装
│   │   ├── user.go
│   │   ├── goods.go
│   │   ├── seckill.go          # 核心抢购接口
│   │   └── order.go
│   ├── service/                # 业务层：核心逻辑
│   │   ├── seckill.go          # 库存扣减、幂等校验、MQ 入队
│   │   └── order.go
│   ├── repo/                   # 数据层：MySQL CRUD
│   │   ├── goods.go
│   │   └── order.go
│   ├── cache/                  # Redis 操作
│   │   ├── keys.go             # Key 统一管理
│   │   ├── stock.go            # Lua 脚本扣减
│   │   └── token.go            # 幂等 Token setNX
│   ├── mq/
│   │   ├── producer.go
│   │   └── consumer.go         # 异步创建订单
│   ├── model/                  # DB 模型 + DTO
│   │   ├── goods.go
│   │   ├── order.go
│   │   └── dto.go
│   └── middleware/
│       ├── auth.go             # JWT 鉴权
│       └── ratelimit.go        # 限流
├── pkg/
│   ├── snowflake.go            # 雪花算法
│   ├── response.go             # 统一响应结构
│   └── errors.go              # 业务错误码
├── config/
│   ├── config.go
│   └── config.yaml
└── scripts/schema.sql

依赖方向（单向）：handler → service → repo/cache/mq → model
pkg 可被任意层引用，但不能反向依赖 internal。

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【四、数据模型】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MySQL 三张表：

-- users
id BIGINT PK（雪花）, username VARCHAR(64) UNIQUE,
password_hash VARCHAR(128), created_at DATETIME

-- seckill_goods
id BIGINT PK（雪花）, name VARCHAR(128),
original_price DECIMAL(10,2), seckill_price DECIMAL(10,2),
total_stock INT, available_stock INT,
start_time DATETIME, end_time DATETIME,
status TINYINT（0草稿/1发布/2进行中/3结束）,
img_url VARCHAR(512), created_at/updated_at DATETIME
索引：idx_status_start(status, start_time)

-- seckill_order
id BIGINT PK（雪花）, user_id BIGINT, goods_id BIGINT,
seckill_price DECIMAL(10,2),
status TINYINT（0待支付/1已支付/2已取消）,
idempotent_key VARCHAR(64) UNIQUE（MQ幂等）,
pay_expire_at DATETIME, paid_at DATETIME（null=未支付）,
created_at/updated_at DATETIME
索引：UNIQUE uk_user_goods(user_id,goods_id)
     UNIQUE uk_idempotent_key(idempotent_key)
     IDX idx_status_expire(status, pay_expire_at)

Redis Key 规范（统一用函数生成，见 cache/keys.go）：
sk:stock:{goodsId}          String  库存计数
sk:sold_out:{goodsId}       String  售罄标记位（值="1"）
sk:users:{goodsId}          Set     已购用户集合
sk:token:{userId}:{goodsId} String  幂等Token(setNX TTL 5min)
sk:result:{userId}:{goodsId}String  抢购结果(0排队/1成功/2失败)
sk:url_token:{userId}:{goodsId} String 动态URL token(TTL 5min)

核心 Lua 脚本（三步原子：已购校验→库存扣减→记录已购）：
KEYS[1]=sk:stock, KEYS[2]=sk:sold_out, KEYS[3]=sk:users
ARGV[1]=userId
返回：1=成功  -1=库存不足  -2=已购

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【五、接口契约（6个核心接口）】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
统一响应：{"code":0,"message":"ok","data":{}}
错误码：0成功 10001参数错 10002未授权 10003限流
       20001商品不存在 20002已售罄 20003已购
       20004Token无效 20005订单不存在 20006订单超时
       50001内部错误

POST /api/v1/user/login
  req: {username, password}
  res: {token, expire_at, user_id}

GET  /api/v1/goods/:goodsId
  res: {id,name,original_price,seckill_price,
        available_stock,start_time,end_time,status,img_url}

GET  /api/v1/seckill/token/:goodsId  [需JWT]
  res: {seckill_token, expire_at}
  限流：同一用户同商品 1min 内最多 3 次

POST /api/v1/seckill/do/:goodsId/:token  [需JWT]
  处理链路：校验url_token → JVM本地售罄预检 →
           Redis Lua原子扣减 → MQ入队 →
           写sk:result="queuing" → 立即返回
  成功res: {queue_status:"queuing"}
  失败：code=20002(售罄) 或 code=20003(已购)
  限流：全局10000 QPS，单用户1次/秒

GET  /api/v1/seckill/result/:goodsId  [需JWT]
  res: {status:"queuing"|"success"|"failed", order_id}
  读 Redis sk:result，不查DB

POST /api/v1/order/:orderId/pay  [需JWT]
  校验：归属当前用户 + 状态=待支付 + 未超时
  res: {order_id, status:"paid", paid_at}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【六、编码实施顺序】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. go.mod 初始化 + 依赖引入
2. config/（Viper 配置加载）
3. pkg/（snowflake、response、errors）
4. internal/model/（Go struct）
5. internal/repo/（MySQL CRUD）
6. internal/cache/（keys.go + Lua脚本 + stock.go + token.go）
7. internal/service/（seckill.go 核心逻辑）
8. internal/mq/（producer + consumer）
9. internal/handler/ + middleware/
10. cmd/server/main.go（组装启动）

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

请确认你已理解以上设计，然后等我说从哪一步开始。