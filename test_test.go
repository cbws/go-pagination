package pagination

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"
)

var data []int

func init() {
	for i := 1; i <= 100; i++ {
		data = append(data, i)
	}
}

func TestIterateForward(t *testing.T) {
	edges, err := Iterate(context.Background(), testDataPaginator)
	if err != nil {
		t.Errorf("error while iterating: %+v", err)

		return
	}

	got := edgesToData(edges)
	if !reflect.DeepEqual(got, data) {
		t.Errorf("got %v want %v", got, data)
	}
}

func TestStreamForward(t *testing.T) {
	edges := make(chan Edge)
	go func() {
		err := Stream(context.Background(), testDataPaginator, edges, OffsetCursor(0))
		if err != nil {
			log.Printf("Error: %+v", err)
		}
	}()

	got := streamToData(edges)
	if !reflect.DeepEqual(got, data) {
		t.Errorf("got %v want %v", got, data)
	}
}

func TestIterateBackward(t *testing.T) {
	edges, err := IterateBackward(context.Background(), testDataPaginator, OffsetCursor(len(data)+1))
	if err != nil {
		t.Errorf("error while iterating: %+v", err)

		return
	}

	got := edgesToData(edges)
	if !reflect.DeepEqual(got, data) {
		t.Errorf("got %v want %v", got, data)
	}
}

func TestIterateBackwardLimits(t *testing.T) {
	for l := 1; l <= len(data); l++ {
		edges, err := IterateBackward(context.Background(), testDataPaginator, OffsetCursor(len(data)+1))
		if err != nil {
			t.Errorf("error while iterating: %+v", err)

			return
		}

		got := edgesToData(edges)
		if !reflect.DeepEqual(got, data) {
			t.Errorf("got %v want %v", got, data)
		}
	}
}

func TestStreamBackward(t *testing.T) {
	edges := make(chan Edge)
	go func() {
		err := StreamBackward(context.Background(), testDataPaginator, edges, OffsetCursor(len(data)+1))
		if err != nil {
			log.Printf("Error: %+v", err)
		}
	}()

	got := reverse(streamToData(edges))
	if !reflect.DeepEqual(got, data) {
		t.Errorf("got %v want %v", got, data)
	}
}

func edgesToData(edges []Edge) []int {
	var data []int
	for _, edge := range edges {
		data = append(data, edge.Node.(int))
	}

	return data
}

func reverse(s []int) []int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}

func streamToData(edges chan Edge) []int {
	var data []int
	for edge := range edges {
		data = append(data, edge.Node.(int))
	}

	return data
}

func testDataPaginator(ctx context.Context, first, last *int, before, after Cursor) (*Connection, error) {
	var limit = 1

	if first != nil {
		limit = *first
	}
	if last != nil {
		limit = *last
	}

	offset := 0
	if after != nil {
		after, ok := after.(OffsetCursor)
		if !ok {
			return nil, fmt.Errorf("expected offset cursor")
		}

		offset = int(after)
	}
	if before != nil {
		before, ok := before.(OffsetCursor)
		if !ok {
			return nil, fmt.Errorf("expected offset cursor")
		}

		offset = int(before)-limit
	}

	end := limit + offset
	if end > len(data) {
		end = len(data)
	}
	if offset < 0 {
		offset = 0
	}
	slice := data[offset:end]
	count := len(slice)

	connection := &Connection{}
	connection.PageInfo = PageInfo{}
	connection.PageInfo.HasNextPage = len(data) > offset + count
	connection.PageInfo.HasPreviousPage = offset != 0
	connection.PageInfo.StartCursor = OffsetCursor(offset)
	connection.PageInfo.EndCursor = DiffOffsetCursor(offset, count)
	connection.Edges = []Edge{}
	for i, n := range slice {
		connection.Edges = append(connection.Edges, Edge{
			Cursor: DiffOffsetCursor(offset, i),
			Node:   n,
		})
	}

	return connection, nil
}

func TestIterateOffsetForward(t *testing.T) {
	edges, err := Iterate(context.Background(), NewOffsetPaginator(testDataOffsetPaginator))
	if err != nil {
		t.Errorf("error while iterating: %+v", err)

		return
	}

	got := edgesToData(edges)
	if !reflect.DeepEqual(got, data) {
		t.Errorf("got %v want %v", got, data)
	}
}

func TestIterateOffsetBackward(t *testing.T) {
	edges, err := IterateBackward(context.Background(), NewOffsetPaginator(testDataOffsetPaginator), OffsetCursor(len(data)+1))
	if err != nil {
		t.Errorf("error while iterating: %+v", err)

		return
	}

	got := edgesToData(edges)
	if !reflect.DeepEqual(got, data) {
		t.Errorf("got %v want %v", got, data)
	}
}

func testDataOffsetPaginator(ctx context.Context, offset int, l *int) (*ResultSet, error) {
	var limit = 1

	if l != nil {
		limit = *l
	}

	log.Printf("Limit: %+v", limit)

	end := limit + offset
	if end > len(data) {
		end = len(data)
	}
	if offset < 0 {
		offset = 0
	}
	slice := data[offset:end]
	count := len(slice)
	log.Printf("Slice: %+v", slice)

	resultSet := &ResultSet{}
	resultSet.HasNextPage = len(data) > offset + count
	resultSet.HasPreviousPage = offset != 0
	resultSet.Nodes = []Node{}
	for _, n := range slice {
		resultSet.Nodes = append(resultSet.Nodes, n)
	}

	return resultSet, nil
}

