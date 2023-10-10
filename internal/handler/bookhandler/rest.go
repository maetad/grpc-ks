package bookhandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maetad/grpc-ks/internal/model"
	"github.com/maetad/grpc-ks/internal/service"
	cursorpagination "github.com/maetad/grpc-ks/pkg/cursor-pagination"
	"go.uber.org/zap"
)

type RESTBookHandler struct {
	logger      *zap.Logger
	bookservice service.BookServicer
}

func NewRESTBookHandler(logger *zap.Logger, bookservice service.BookServicer) *RESTBookHandler {
	return &RESTBookHandler{logger, bookservice}
}

func (h *RESTBookHandler) ListBooks(ctx *gin.Context) {
	c, _ := cursorpagination.Decode(ctx.Query("pageToken"))
	pageSize, _ := strconv.Atoi(ctx.Query("pageSize"))
	if pageSize <= 0 {
		pageSize = 25
	}
	books, err := h.bookservice.List(c.Offset, pageSize)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Offset += pageSize

	ctx.JSON(http.StatusOK, gin.H{"books": books, "nextPageToken": cursorpagination.Encode(c)})
}

func (h *RESTBookHandler) CreateBook(ctx *gin.Context) {
	var (
		b   = &model.Book{}
		err error
	)

	if err = ctx.BindJSON(b); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	if err = h.bookservice.Create(b); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, b)
}

func (h *RESTBookHandler) GetBook(ctx *gin.Context) {
	var (
		b   = &model.Book{}
		err error
	)
	if b.ID, err = uuid.Parse(ctx.Param("id")); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	if b, err = h.bookservice.Get(b.ID); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, b)
}
