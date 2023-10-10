package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	book "github.com/maetad/grpc-ks-proto-gen/go/book/v1alpha1"
	"github.com/maetad/grpc-ks/internal/model"
)

type CreateBookRequest struct {
	Title     string `json:"title"`
	Author    string `json:"author"`
	Isbn      string `json:"isbn"`
	Publisher string `json:"publisher"`
}

func GetBookWithRest(n string) (*model.Book, error) {
	url := fmt.Sprintf("%s/%s", os.Getenv("APP_REST_ADDRESS"), n)
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	var book = &model.Book{}
	if err = json.NewDecoder(r.Body).Decode(book); err != nil {
		return nil, err
	}

	return book, nil
}

func CreateBookWithRest(body CreateBookRequest) (*model.Book, error) {
	url := fmt.Sprintf("%s", os.Getenv("APP_REST_ADDRESS"))
	jsonBody := []byte(fmt.Sprintf(`{
		"title": "%s",
		"author": "%s",
		"isbn": "%s",
		"publisher": "%s"
	}`, body.Title, body.Author, body.Isbn, body.Publisher))
	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	r, err := client.Do(req)
	var book = &model.Book{}
	if err = json.NewDecoder(r.Body).Decode(book); err != nil {
		return nil, err
	}

	return book, nil
}

func GetBookWithGRPC(c book.BookServiceClient, ctx context.Context, n string) (*model.Book, error) {
	r, err := c.GetBook(ctx, &book.GetBookRequest{
		Name: fmt.Sprintf("books/%s", n),
	})
	if err != nil {
		return nil, err
	}

	return model.NewBookFromProto(r.GetBook())
}

func CreateBookWithGRPC(c book.BookServiceClient, ctx context.Context, body CreateBookRequest) (*model.Book, error) {
	r, err := c.CreateBook(ctx, &book.CreateBookRequest{
		Book: &book.Book{
			Title:     body.Title,
			Author:    body.Author,
			Isbn:      body.Isbn,
			Publisher: body.Publisher,
		},
	})
	if err != nil {
		return nil, err
	}

	return model.NewBookFromProto(r.GetBook())
}
