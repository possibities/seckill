# Seckill 项目测试流程（新手可直接照做）

你不需要会写测试代码，按下面命令执行并看结果即可。

## 1. 环境准备（只需一次）

- 安装 Go（建议 1.25+）
- 安装 MySQL 8
- 安装 Redis 7
- 安装 RocketMQ 5（后续步骤会用到）

## 2. 拉取依赖

在项目根目录执行：

```powershell
go mod tidy
```

## 3. 快速编译检查（最常用）

```powershell
go test ./...
```

判断标准：
- 看到 `FAIL`：有问题，需要修复
- 没有 `FAIL`，并显示各包结果：通过

## 4. 仅测试某个包（定位问题）

示例（测试 config 包）：

```powershell
go test ./config
```

## 5. 查看详细日志

```powershell
go test -v ./...
```

## 6. 竞态检查（并发问题排查，较慢）

```powershell
go test -race ./...
```

## 7. 覆盖率（可选）

```powershell
go test -cover ./...
```

## 8. 启动前自检（后续 main.go 完成后再用）

```powershell
go run ./cmd/server
```

判断标准：
- 成功监听端口（例如 `:8080`）表示启动正常
- 启动报错则先检查 `config/config.yaml`、MySQL、Redis、RocketMQ 是否可用

## 9. MQ 阶段联调检查（第 8 步后可用）

先跑单元测试（不依赖真实 MQ）：

```powershell
go test ./internal/mq -v
```

再做服务联调（依赖本地 RocketMQ）：

1. 启动 NameServer 和 Broker
2. 启动项目服务（后续 main.go 完成后）
3. 执行一次秒杀请求，观察日志是否出现“消息入队”和“消费创建订单”
4. 查询 `seckill_order` 表确认订单创建
5. 查询 Redis `sk:result:{userId}:{goodsId}` 确认状态为成功

## 10. 你每次更新后只需要执行这 2 条

```powershell
go test ./...
go test -race ./...
```

如果这两条都通过，通常就可以放心继续开发或提交。

---

后续我会持续补充该文件：
- 新增接口后，会加接口测试步骤
- 新增消息队列后，会加 MQ 联调步骤
- 新增压测后，会加压测命令与结果判定
