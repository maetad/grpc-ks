package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/maetad/grpc-ks/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BookServicer interface {
	Count() (int, error)
	List(offset, limit int) ([]*model.Book, error)
	Get(id uuid.UUID) (*model.Book, error)
	Create(book *model.Book) error
	Update(book *model.Book, fields []string) error
	Delete(id uuid.UUID) error
}

type BookService struct {
	logger *zap.Logger
	db     *gorm.DB
}

func NewBookService(logger *zap.Logger, db *gorm.DB) BookServicer {
	return &BookService{logger, db}
}

func (s *BookService) Count() (int, error) {
	var count int64
	if err := s.db.Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func (s *BookService) List(offset, limit int) ([]*model.Book, error) {
	var books = []*model.Book{}

	if err := s.db.Offset(offset).Limit(limit).Find(&books).Error; err != nil {
		return nil, err
	}

	return books, nil
}

func (s *BookService) Get(id uuid.UUID) (*model.Book, error) {
	book := &model.Book{}
	if err := s.db.First(book, id).Error; err != nil {
		return nil, err
	}

	return book, nil
}

func (s *BookService) Create(book *model.Book) error {
	if book == nil {
		return errors.New("book is empty")
	}
	now := time.Now()
	book.ID = uuid.New()
	book.CreatedAt = now
	book.UpdatedAt = now
	book.DeletedAt = gorm.DeletedAt{}

	return s.db.Create(book).Error
}

func (s *BookService) Update(book *model.Book, fields []string) error {
	if book == nil {
		return errors.New("book is empty")
	}

	return s.db.Model(book).Select(fields).Updates(book).Error
}

func (s *BookService) Delete(id uuid.UUID) error {
	return s.db.Delete(&model.Book{}, id).Error
}
