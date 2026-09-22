package kafka

import "testing"

func TestLagFromOffsets(t *testing.T) {
	tests := []struct {
		name      string
		committed int64
		high      int64
		want      float64
	}{
		{name: "positive lag", committed: 10, high: 25, want: 15},
		{name: "caught up", committed: 100, high: 100, want: 0},
		{name: "clamps negative", committed: 50, high: 40, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lagFromOffsets(tt.committed, tt.high)
			if got != tt.want {
				t.Fatalf("lagFromOffsets(%d, %d) = %v, want %v", tt.committed, tt.high, got, tt.want)
			}
		})
	}
}
