package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maetad/grpc-ks/internal/handler/bookhandler"
	"github.com/maetad/grpc-ks/internal/model"
	"github.com/maetad/grpc-ks/internal/service"
	"go.uber.org/zap"
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
	logger, _ := zap.NewProduction()
	router := gin.Default()

	// Services
	bs := service.NewBookService(logger, db)
	bh := bookhandler.NewRESTBookHandler(logger, bs)

	router.GET("", bh.ListBooks)
	router.POST("", bh.CreateBook)
	router.GET("/:id", bh.GetBook)

	srv := &http.Server{
		Addr:    os.Getenv("REST_LISTEN_ADDRESS"),
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
