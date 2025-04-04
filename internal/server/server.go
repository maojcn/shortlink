package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maojcn/shortlink/internal/config"
	"github.com/maojcn/shortlink/internal/handlers"
	"github.com/maojcn/shortlink/internal/middleware"
	"github.com/maojcn/shortlink/internal/repository"
	"go.uber.org/zap"
)

// Server 封装了HTTP服务器及其依赖项
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	logger     *zap.Logger
	db         *repository.PostgresRepo
	redis      *repository.RedisRepo
}

// New 创建并初始化一个新的Server实例
func New(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// 设置Gin模式
	if cfg.LogLevel == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	router := gin.New()

	// 初始化数据库连接
	db, err := repository.NewPostgresRepo(cfg.Database.DSN(), logger)
	if err != nil {
		return nil, err
	}

	// 初始化Redis连接
	redis, err := repository.NewRedisRepo(ctx, cfg.Redis, logger)
	if err != nil {
		return nil, err
	}

	// 创建服务器实例
	s := &Server{
		router: router,
		httpServer: &http.Server{
			Addr:    cfg.Server.Address,
			Handler: router,
		},
		logger: logger,
		db:     db,
		redis:  redis,
	}

	// 设置中间件
	s.setupMiddleware()

	// 设置路由
	s.setupRoutes()

	return s, nil
}

// setupMiddleware 添加中间件到Gin路由
func (s *Server) setupMiddleware() {
	// 添加请求ID中间件
	s.router.Use(middleware.RequestID())

	// 添加日志中间件
	s.router.Use(middleware.Logger(s.logger))

	// 添加恢复中间件
	s.router.Use(middleware.Recovery(s.logger))

	// 添加CORS中间件
	s.router.Use(middleware.CORS())
}

// setupRoutes 设置所有API路由
func (s *Server) setupRoutes() {
	// 创建处理程序
	h := handlers.NewHandler(s.db, s.redis, s.logger)

	// 健康检查
	s.router.GET("/health", h.HealthCheck)

	// API V1 路由组
	v1 := s.router.Group("/api/v1")
	{
		// 用户相关路由
		users := v1.Group("/users")
		{
			users.GET("", h.ListUsers)
			users.POST("", h.CreateUser)
			users.GET("/:id", h.GetUser)
			users.PUT("/:id", h.UpdateUser)
			users.DELETE("/:id", h.DeleteUser)
		}

		// 可以添加更多路由组...
	}
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown 优雅地关闭服务器及其资源
func (s *Server) Shutdown(ctx context.Context) error {
	// 设置关闭超时
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("HTTP服务器关闭失败", zap.Error(err))
	}

	// 关闭数据库连接
	if err := s.db.Close(); err != nil {
		s.logger.Error("数据库连接关闭失败", zap.Error(err))
	}

	// 关闭Redis连接
	if err := s.redis.Close(); err != nil {
		s.logger.Error("Redis连接关闭失败", zap.Error(err))
	}

	return nil
}
