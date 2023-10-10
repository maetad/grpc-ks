package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	book "github.com/maetad/grpc-ks-proto-gen/go/book/v1alpha1"
	"github.com/maetad/grpc-ks/cmd/client/handler"
	"github.com/maetad/grpc-ks/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("book.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&model.Book{})
}

func main() {
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	router := gin.Default()
	router.GET("health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"messsage": "OK"})
	})

	conn, err := grpc.DialContext(
		ctx,
		os.Getenv("APP_GRPC_ADDRESS"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Panic("cannot connect to book service", zap.Error(err))
	}

	bookclient := book.NewBookServiceClient(conn)

	rest := router.Group("/rest")
	{
		rest.GET("/books", func(ctx *gin.Context) {
			names := strings.Split(ctx.Query("names"), ",")
			var res = []*model.Book{}
			for _, n := range names {
				logger.Sugar().Info(n)
				b, err := handler.GetBookWithRest(n)
				if err != nil {
					logger.Sugar().Error(err)
					continue
				}
				res = append(res, b)
			}

			ctx.JSON(http.StatusOK, res)
		})
		rest.POST("/books", func(ctx *gin.Context) {
			var req handler.CreateBookRequest

			if err = ctx.BindJSON(&req); err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
				return
			}

			b, err := handler.CreateBookWithRest(req)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
				return
			}

			ctx.JSON(http.StatusCreated, b)
		})
	}

	grpcrouter := router.Group("/grpc")
	{
		grpcrouter.GET("/books", func(ctx *gin.Context) {
			names := strings.Split(ctx.Query("names"), ",")
			var res = []*model.Book{}
			for _, n := range names {
				logger.Sugar().Info(n)
				b, err := handler.GetBookWithGRPC(bookclient, ctx.Request.Context(), n)
				if err != nil {
					logger.Sugar().Error(err)
					continue
				}

				res = append(res, b)
			}

			ctx.JSON(http.StatusOK, res)
		})
		grpcrouter.POST("/books", func(ctx *gin.Context) {
			var req handler.CreateBookRequest

			if err = ctx.BindJSON(&req); err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
				return
			}

			b, err := handler.CreateBookWithGRPC(bookclient, ctx.Request.Context(), req)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
				return
			}

			ctx.JSON(http.StatusCreated, b)
		})
	}

	srv := &http.Server{
		Addr:    os.Getenv("APP_LISTEN_ADDRESS"),
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Sugar().Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown:", zap.Error(err))
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		logger.Info("timeout of 5 seconds.")
	}
	logger.Info("Server exiting")
}
