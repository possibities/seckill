package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"seckill/config"
	"seckill/internal/cache"
	"seckill/internal/handler"
	"seckill/internal/middleware"
	"seckill/internal/model"
	"seckill/internal/mq"
	"seckill/internal/repo"
	"seckill/internal/service"
	"seckill/pkg"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "server start failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("./config")
	if err != nil {
		return err
	}

	logger, err := newLogger(cfg.Log.Level)
	if err != nil {
		return err
	}
	defer logger.Sync()

	db, err := initDB(cfg)
	if err != nil {
		return err
	}
	if err = db.AutoMigrate(&model.User{}, &model.SeckillGoods{}, &model.SeckillOrder{}); err != nil {
		return err
	}

	rdb := initRedis(cfg)
	if err = rdb.Ping(context.Background()).Err(); err != nil {
		return err
	}
	defer rdb.Close()

	snowflake, err := pkg.NewSnowflake(1)
	if err != nil {
		return err
	}

	goodsRepo, _ := repo.NewGoodsRepo(db)
	orderRepo, _ := repo.NewOrderRepo(db)
	userRepo, _ := repo.NewUserRepo(db)

	stockCache, _ := cache.NewStockCache(rdb)
	tokenCache, _ := cache.NewTokenCache(rdb)

	producer, err := mq.NewProducer(cfg.RocketMQ)
	if err != nil {
		return err
	}
	if err = producer.Start(); err != nil {
		return err
	}
	defer producer.Shutdown()

	consumer, err := mq.NewConsumer(cfg.RocketMQ, orderRepo, goodsRepo, stockCache, snowflake)
	if err != nil {
		return err
	}
	if err = consumer.Start(); err != nil {
		return err
	}
	defer consumer.Shutdown()

	userSvc, _ := service.NewUserService(userRepo)
	seckillSvc, _ := service.NewSeckillService(goodsRepo, stockCache, tokenCache, producer)
	orderSvc, _ := service.NewOrderService(orderRepo)

	jwt := middleware.NewJWTManager(cfg.JWT)

	userHandler, _ := handler.NewUserHandler(userSvc, jwt)
	goodsHandler, _ := handler.NewGoodsHandler(goodsRepo)
	seckillHandler, _ := handler.NewSeckillHandler(seckillSvc)
	orderHandler, _ := handler.NewOrderHandler(orderSvc)

	engine := gin.New()
	engine.Use(gin.Recovery())
	handler.RegisterRoutes(engine, handler.RouterDeps{
		JWT:     jwt,
		User:    userHandler,
		Goods:   goodsHandler,
		Seckill: seckillHandler,
		Order:   orderHandler,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server starting", zap.String("addr", srv.Addr))
		errCh <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	case serveErr := <-errCh:
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func newLogger(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if level != "" {
		if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
			return nil, err
		}
	}
	return cfg.Build()
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MySQL.ConnMaxLifetime) * time.Second)
	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})
}
