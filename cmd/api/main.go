package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"ai-language-notes/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func main() {
	cfg := config.LoadDatabaseConfig()
	connString := cfg.GetConnectionString()

	var err error
	dbPool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// 测试数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = dbPool.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")

	r := gin.Default()

	// 健康检查路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 示例 API 路由 (假设你需要访问数据库)
	r.GET("/api/example", func(c *gin.Context) {
		// 在这里可以使用 dbPool 执行数据库操作
		c.JSON(http.StatusOK, gin.H{
			"data": "This is an example API response.",
		})
	})

	r.Run(":8080")
}
