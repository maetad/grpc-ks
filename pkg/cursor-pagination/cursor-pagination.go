package cursorpagination

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"io"
	"strings"
)

type Cursor struct {
	Offset int
}

type Counter interface {
	Count() (int, error)
}

func Decode(s string) (c Cursor, err error) {
	dec := gob.NewDecoder(base64.NewDecoder(base64.URLEncoding, strings.NewReader(s)))
	if err = dec.Decode(&c); err != nil && !errors.Is(err, io.EOF) {
		return
	}
	return c, nil
}

func Encode(c Cursor) string {
	var b strings.Builder
	base64Encoder := base64.NewEncoder(base64.URLEncoding, &b)
	gobEncoder := gob.NewEncoder(base64Encoder)
	_ = gobEncoder.Encode(c)
	_ = base64Encoder.Close()

	return b.String()
}

func GetNextPageToken(counter Counter, nextOffset int) (string, error) {
	total, err := counter.Count()
	if err != nil {
		return "", err
	}

	if nextOffset > total {
		return "", nil
	}

	return Encode(Cursor{
		Offset: nextOffset,
	}), nil
}
