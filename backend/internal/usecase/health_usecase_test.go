package usecase

import (
	"context"
	"errors"
	"testing"

	"backend/internal/domain"
)

// mockHealthReader is a local test double implementing HealthReader.
type mockHealthReader struct {
	stats domain.HealthStats
	err   error
}

func (m *mockHealthReader) Health(_ context.Context) (domain.HealthStats, error) {
	return m.stats, m.err
}

func TestGetHealth_Messages(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.HealthStats
		wantMsg string
	}{
		{
			name:    "default healthy message",
			input:   domain.HealthStats{OpenConnections: 5},
			wantMsg: "It's healthy",
		},
		{
			name:    "high open connections",
			input:   domain.HealthStats{OpenConnections: 41},
			wantMsg: "The database is experiencing heavy load.",
		},
		{
			name:    "exactly 40 connections is not heavy load",
			input:   domain.HealthStats{OpenConnections: 40},
			wantMsg: "It's healthy",
		},
		{
			name:    "high wait count",
			input:   domain.HealthStats{OpenConnections: 5, WaitCount: 1001},
			wantMsg: "The database has a high number of wait events, indicating potential bottlenecks.",
		},
		{
			name:    "exactly 1000 wait count is not a bottleneck",
			input:   domain.HealthStats{OpenConnections: 5, WaitCount: 1000},
			wantMsg: "It's healthy",
		},
		{
			name: "max idle closed exceeds half of open connections",
			// OpenConnections=10, int64(10)/2 = 5; MaxIdleClosed=6 > 5 → triggers
			input:   domain.HealthStats{OpenConnections: 10, MaxIdleClosed: 6},
			wantMsg: "Many idle connections are being closed, consider revising the connection pool settings.",
		},
		{
			name: "max idle closed does not exceed half of open connections",
			// OpenConnections=10, int64(10)/2 = 5; MaxIdleClosed=5 is not > 5
			input:   domain.HealthStats{OpenConnections: 10, MaxIdleClosed: 5},
			wantMsg: "It's healthy",
		},
		{
			name: "max lifetime closed exceeds half of open connections",
			// OpenConnections=10, int64(10)/2 = 5; MaxLifetimeClosed=6 > 5 → triggers
			input:   domain.HealthStats{OpenConnections: 10, MaxLifetimeClosed: 6},
			wantMsg: "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern.",
		},
		{
			name: "max lifetime closed does not exceed half of open connections",
			// OpenConnections=10, int64(10)/2 = 5; MaxLifetimeClosed=5 is not > 5
			input:   domain.HealthStats{OpenConnections: 10, MaxLifetimeClosed: 5},
			wantMsg: "It's healthy",
		},
		{
			// The conditions are evaluated in sequence and the last matching one wins.
			// With OpenConnections=0, int64(0)/2=0 so MaxLifetimeClosed=1 > 0 is true.
			name:    "max lifetime closed branch wins over max idle closed when both trigger",
			input:   domain.HealthStats{OpenConnections: 0, MaxIdleClosed: 1, MaxLifetimeClosed: 1},
			wantMsg: "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewHealthUseCase(&mockHealthReader{stats: tc.input})
			got, err := uc.GetHealth(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Message != tc.wantMsg {
				t.Errorf("message mismatch\n  got:  %q\n  want: %q", got.Message, tc.wantMsg)
			}
		})
	}
}

func TestGetHealth_RepoErrorPropagates(t *testing.T) {
	sentinel := errors.New("db unavailable")
	uc := NewHealthUseCase(&mockHealthReader{err: sentinel})

	_, err := uc.GetHealth(context.Background())
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
}
