package bookhandler

import (
	"context"

	book "github.com/maetad/grpc-ks-proto-gen/go/book/v1alpha1"
	"github.com/maetad/grpc-ks/internal/model"
	"github.com/maetad/grpc-ks/internal/service"
	cursorpagination "github.com/maetad/grpc-ks/pkg/cursor-pagination"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCBookHandler struct {
	book.UnimplementedBookServiceServer
	logger      *zap.Logger
	bookservice service.BookServicer
}

func NewGRPCBookHandler(logger *zap.Logger, bookservice service.BookServicer) *GRPCBookHandler {
	return &GRPCBookHandler{logger: logger, bookservice: bookservice}
}

func (h *GRPCBookHandler) ListBooks(ctx context.Context, req *book.ListBooksRequest) (*book.ListBooksResponse, error) {
	c, _ := cursorpagination.Decode(string(req.GetPageToken()))
	pageSize := 25
	if req.GetPageSize() > 0 {
		pageSize = int(req.GetPageSize())
	}
	books, err := h.bookservice.List(c.Offset, pageSize)
	if err != nil {
		return nil, err
	}

	c.Offset += pageSize

	var res = &book.ListBooksResponse{
		Books:         []*book.Book{},
		NextPageToken: []byte(cursorpagination.Encode(c)),
	}

	for _, book := range books {
		res.Books = append(res.Books, book.ToProto())
	}

	return res, nil
}

func (h *GRPCBookHandler) CreateBook(ctx context.Context, req *book.CreateBookRequest) (*book.CreateBookResponse, error) {
	var (
		b   *model.Book
		err error
	)

	if err = req.Book.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	b, err = model.NewBookFromProto(req.GetBook())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = h.bookservice.Create(b); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &book.CreateBookResponse{
		Book: b.ToProto(),
	}, nil
}

func (h *GRPCBookHandler) GetBook(ctx context.Context, req *book.GetBookRequest) (*book.GetBookResponse, error) {
	var (
		b   = &model.Book{}
		err error
	)
	if err = b.SetIDFromRRN(req.GetName()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if b, err = h.bookservice.Get(b.ID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &book.GetBookResponse{
		Book: b.ToProto(),
	}, nil
}

func (h *GRPCBookHandler) UpdateBook(ctx context.Context, req *book.UpdateBookRequest) (*book.UpdateBookResponse, error) {
	var (
		b   *model.Book
		err error
	)

	if err = req.Book.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	b, err = model.NewBookFromProto(req.GetBook())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = h.bookservice.Update(b, req.GetUpdateMask().GetPaths()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &book.UpdateBookResponse{
		Book: b.ToProto(),
	}, nil
}

func (h *GRPCBookHandler) DeleteBook(ctx context.Context, req *book.DeleteBookRequest) (*emptypb.Empty, error) {
	var (
		b   = &model.Book{}
		err error
	)
	if err = b.SetIDFromRRN(req.GetName()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = h.bookservice.Delete(b.ID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
