package main

import (
	"reflect"
	"testing"
	"time"
)

func Test_nextMonthlyRun(t *testing.T) {

	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "next monthly run",
			now:  time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
			want: time.Date(1, time.January, 26, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "next monthly run",
			now:  time.Date(2025, time.January, 27, 0, 0, 0, 0, time.UTC),
			want: time.Date(2025, time.January, 28, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextMonthlyRun(tt.now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nextMonthlyRun() = %v, want %v", got, tt.want)
			}
		})
	}
}
