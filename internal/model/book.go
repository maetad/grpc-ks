package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	book "github.com/maetad/grpc-ks-proto-gen/go/book/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type Book struct {
	ID        uuid.UUID `gorm:"primarykey"`
	Title     string
	Author    string
	Isbn      string
	Publisher string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (b Book) GetResourceName() string {
	return "books"
}

func (b *Book) GetRRN() string {
	return fmt.Sprintf("%s/%s", b.GetResourceName(), b.ID)
}

func (b *Book) SetIDFromRRN(rrn string) (err error) {
	b.ID, err = uuid.Parse(strings.TrimPrefix(rrn, fmt.Sprintf("%s/", b.GetResourceName())))
	return err
}

func (b *Book) ToProto() *book.Book {
	return &book.Book{
		Name:      b.GetRRN(),
		Title:     b.Title,
		Author:    b.Author,
		Isbn:      b.Isbn,
		Publisher: b.Publisher,
		CreatedAt: timestamppb.New(b.CreatedAt),
		UpdatedAt: timestamppb.New(b.UpdatedAt),
		DeletedAt: func() *timestamppb.Timestamp {
			if !b.DeletedAt.Valid {
				return nil
			}

			return timestamppb.New(b.DeletedAt.Time)
		}(),
	}
}

func NewBookFromProto(proto *book.Book) (b *Book, err error) {
	b = &Book{}
	b.SetIDFromRRN(proto.GetName())

	b.Title = proto.GetTitle()
	b.Author = proto.GetAuthor()
	b.Isbn = proto.GetIsbn()
	b.Publisher = proto.GetPublisher()
	b.CreatedAt = proto.GetCreatedAt().AsTime()
	b.UpdatedAt = proto.GetUpdatedAt().AsTime()
	b.DeletedAt = gorm.DeletedAt{
		Time:  proto.GetDeletedAt().AsTime(),
		Valid: !proto.GetDeletedAt().AsTime().IsZero(),
	}

	return b, nil
}
