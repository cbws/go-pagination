package pagination

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
)

var SafeDefaultLast = 10

func DiffOffsetCursor(offset int, count int) OffsetCursor {
	return OffsetCursor(offset + count)
}

type OffsetCursor int

func (c OffsetCursor) GetCursor() string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(c))))
}

func NewOffsetPaginator(f OffsetPaginator) Paginator {
	return func(ctx context.Context, first, last *int, before, after Cursor) (*Connection, error) {
		var limit *int

		if first != nil {
			limit = first
		}
		if last != nil {
			limit = last
		}
		// Using backwards pagination and the limit has not been provided
		if before != nil && last == nil {
			limit = &SafeDefaultLast
		}

		var offset = 0
		if after != nil {
			after, err := GetOffsetCursor(after)
			if err != nil {
				return nil, fmt.Errorf("expected offset cursor: %+v", err)
			}

			offset = int(*after)+1
		}
		//log.Printf("Bla offset: %+v", offset)
		if before != nil {
			before, err := GetOffsetCursor(before)
			if err != nil {
				return nil, fmt.Errorf("expected offset cursor: %+v", err)
			}

			// When the client requests nodes that exist before another but lower than the limit, reset the
			// limit so that the client won't get duplicate nodes.
			if limit != nil && before != nil && *limit > int(*before) {
				limit = (*int)(before)
			}

			offset = int(*before)-*limit
		}

		if offset < 0 {
			offset = 0
		}

		// Call the offset paginator
		r, err := f(ctx, offset, limit)
		if err != nil {
			return nil, err
		}


		count := len(r.Nodes)

		connection := &Connection{}
		connection.PageInfo = PageInfo{}
		connection.PageInfo.HasNextPage = r.HasNextPage
		connection.PageInfo.HasPreviousPage = r.HasPreviousPage
		connection.PageInfo.StartCursor = OffsetCursor(offset)
		connection.PageInfo.EndCursor = DiffOffsetCursor(offset, count-1)
		connection.Edges = []Edge{}
		for i, n := range r.Nodes {
			connection.Edges = append(connection.Edges, Edge{
				Cursor: DiffOffsetCursor(offset, i),
				Node:   n,
			})
		}

		return connection, nil
	}
}

type OffsetPaginator func(ctx context.Context, offset int, limit *int) (*ResultSet, error)

type ResultSet struct {
	HasNextPage bool
	HasPreviousPage bool
	Nodes []Node
}
