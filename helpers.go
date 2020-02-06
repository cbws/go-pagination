package pagination

import (
	"context"
	"fmt"
)

func Iterate(ctx context.Context, f Paginator) ([]Edge, error) {
	var after Cursor
	var connection *Connection
	var err error
	var edges []Edge

	for connection == nil || connection.PageInfo.HasNextPage {
		connection, err = f(ctx, nil, nil, nil, after)
		if err != nil {
			return nil, err
		}

		for _, edge := range connection.Edges {
			edges = append(edges, edge)
		}

		after = connection.PageInfo.EndCursor
	}

	return edges, nil
}

func IterateBackward(ctx context.Context, f Paginator, before Cursor) ([]Edge, error) {
	var connection *Connection
	var err error
	var edges []Edge

	for connection == nil || connection.PageInfo.HasPreviousPage {
		connection, err = f(ctx, nil, nil, before, nil)
		if err != nil {
			return nil, err
		}

		for i, _ := range connection.Edges {
			edges = append([]Edge{connection.Edges[len(connection.Edges)-1-i]}, edges...)
		}

		before = connection.PageInfo.StartCursor
	}

	return edges, nil
}

func Stream(ctx context.Context, f Paginator, c chan<- Edge, after Cursor) error {
	errc := make(chan error)

	// This buffer channel allows for pre-fetching the next page before the client needs it
	bufferChan := make(chan Edge, 10)

	go func() {
		var connection *Connection
		var err error

		for connection == nil || connection.PageInfo.HasNextPage {
			connection, err = f(ctx, nil, nil, nil, after)
			if err != nil {
				errc <- err

				return
			}

			for _, edge := range connection.Edges {
				bufferChan <- edge
			}

			after = connection.PageInfo.EndCursor
		}

		close(bufferChan)
	}()

	for {
		select {
		// Error while listing virtual machines
		case err := <- errc:
			return err
		// Context cancelled
		case <-ctx.Done():
			return fmt.Errorf("context was cancelled while listing virtual machines: %v", ctx.Err())
		case edge, ok := <-bufferChan:
			if !ok {
				close(c)

				return nil
			}

			c <- edge
		}
	}
}

func StreamBackward(ctx context.Context, f Paginator, c chan<- Edge, before Cursor) error {
	errc := make(chan error)

	// This buffer channel allows for pre-fetching the next page before the client needs it
	bufferChan := make(chan Edge, 10)

	go func() {
		var connection *Connection
		var err error

		for connection == nil || connection.PageInfo.HasPreviousPage {
			connection, err = f(ctx, nil, nil, before, nil)
			if err != nil {
				errc <- err

				return
			}

			for i, _ := range connection.Edges {
				bufferChan <- connection.Edges[len(connection.Edges)-1-i]
			}

			before = connection.PageInfo.StartCursor
		}

		close(bufferChan)
	}()

	for {
		select {
		// Error while listing virtual machines
		case err := <- errc:
			return err
		// Context cancelled
		case <-ctx.Done():
			return fmt.Errorf("context was cancelled while listing virtual machines: %v", ctx.Err())
		case edge, ok := <-bufferChan:
			if !ok {
				close(c)

				return nil
			}

			c <- edge
		}
	}
}
