package pagination

import (
	"context"
	"encoding/base64"
	"strconv"
)

type Paginator func(ctx context.Context, first, last *int, before, after Cursor) (*Connection, error)

type Connection struct {
	Edges []Edge
	PageInfo PageInfo
}

type PageInfo struct {
	HasNextPage bool
	HasPreviousPage bool
	StartCursor Cursor
	EndCursor Cursor
}

type Edge struct {
	Cursor Cursor
	Node Node
}

type Node interface {
}

type Cursor interface {
	GetCursor() string
}

type StringCursor string

func (c StringCursor) GetCursor() string {
	return string(c)
}

func GetOffsetCursor(c Cursor) (*OffsetCursor, error) {
	decodeString, err := base64.StdEncoding.DecodeString(c.GetCursor())
	if err != nil {
		panic("oh no")
	}

	atoi, err := strconv.Atoi(string(decodeString))
	if err != nil {
		panic("oh no 2")
	}

	cursor := OffsetCursor(atoi)
	return &cursor, nil
}
