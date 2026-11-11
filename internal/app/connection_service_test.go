package app

import (
	"context"
	"errors"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

type stubConnectionSource struct {
	connections []domain.Connection
	err         error
}

func (source stubConnectionSource) ListConnections(context.Context) ([]domain.Connection, error) {
	return source.connections, source.err
}

func TestConnectionServiceSnapshot(testRun *testing.T) {
	connections := []domain.Connection{makeTestConnection(22, "203.0.113.10", 50000)}
	snapshot, err := NewConnectionService(stubConnectionSource{connections: connections}).Snapshot(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	if len(snapshot.Connections) != 1 || snapshot.Summary.Total != 1 || snapshot.Truncated {
		testRun.Fatalf("snapshot = %+v", snapshot)
	}
}

func TestConnectionServiceTruncationIsNotAnError(testRun *testing.T) {
	connections := []domain.Connection{makeTestConnection(22, "203.0.113.10", 50000)}
	source := stubConnectionSource{connections: connections, err: ports.ErrConnectionsTruncated}
	snapshot, err := NewConnectionService(source).Snapshot(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	if !snapshot.Truncated || len(snapshot.Connections) != 1 {
		testRun.Fatalf("snapshot = %+v, want truncated with data", snapshot)
	}
}

func TestConnectionServiceSourceFailureIsReturned(testRun *testing.T) {
	failure := errors.New("proc unreadable")
	_, err := NewConnectionService(stubConnectionSource{err: failure}).Snapshot(context.Background())
	if !errors.Is(err, failure) {
		testRun.Fatalf("err = %v, want wrapped %v", err, failure)
	}
}

func TestConnectionServiceWithoutSource(testRun *testing.T) {
	_, err := NewConnectionService(nil).Snapshot(context.Background())
	if !errors.Is(err, ErrNoConnectionSource) {
		testRun.Fatalf("err = %v, want ErrNoConnectionSource", err)
	}
}
