package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	book "github.com/maetad/grpc-ks-proto-gen/go/book/v1alpha1"
	"github.com/maetad/grpc-ks/internal/handler/bookhandler"
	"github.com/maetad/grpc-ks/internal/model"
	"github.com/maetad/grpc-ks/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	_, cancel := context.WithCancel(context.Background())

	logger, _ := zap.NewProduction()

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	grpcSrv := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(interceptorLogger(logger), opts...),
			recovery.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(interceptorLogger(logger), opts...),
			recovery.UnaryServerInterceptor(),
		),
	)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		sig := <-sigCh
		logger.Sugar().Infof("got signal %v, attempting graceful shutdown", sig)
		cancel()
		grpcSrv.GracefulStop()
		wg.Done()
	}()

	listener, err := net.Listen("tcp", os.Getenv("GRPC_LISTEN_ADDRESS"))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// Services
	bs := service.NewBookService(logger, db)

	// GRPC Handlers
	bh := bookhandler.NewGRPCBookHandler(logger, bs)
	book.RegisterBookServiceServer(grpcSrv, bh)

	go func() {
		logger.Sugar().Info("starting grpc server")
		if err := grpcSrv.Serve(listener); err != nil {
			logger.Fatal("could not serve", zap.Error(err))
		}
	}()

	wg.Wait()
	logger.Info("clean shutdown")
}

func interceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			l.Debug(msg)
		case logging.LevelInfo:
			l.Info(msg)
		case logging.LevelWarn:
			l.Warn(msg)
		case logging.LevelError:
			l.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
