package preconditions

import (
	"reflect"
	"testing"
	"time"
)

type testClock struct {
	ret time.Time
}

func (t *testClock) Now() time.Time {
	return t.ret
}

func TestWindowedPrecondition_PerformCheck(t *testing.T) {
	type fields struct {
		FromHour int
		ToHour   int
		clock    Clock
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "hour in between",
			fields: fields{
				FromHour: 0,
				ToHour:   2,
				clock: &testClock{
					ret: time.Date(2023, time.Month(2), 21, 1, 10, 30, 0, time.UTC),
				},
			},
			want: true,
		},
		{
			name: "edge from",
			fields: fields{
				FromHour: 0,
				ToHour:   2,
				clock: &testClock{
					ret: time.Date(2023, time.Month(2), 21, 0, 10, 30, 0, time.UTC),
				},
			},
			want: true,
		},
		{
			name: "edge to",
			fields: fields{
				FromHour: 0,
				ToHour:   2,
				clock: &testClock{
					ret: time.Date(2023, time.Month(2), 21, 2, 10, 30, 0, time.UTC),
				},
			},
			want: false,
		},
		{
			name: "overlapping interval - within range",
			fields: fields{
				FromHour: 18,
				ToHour:   2,
				clock: &testClock{
					ret: time.Date(2023, time.Month(2), 21, 0, 10, 30, 0, time.UTC),
				},
			},
			want: true,
		},
		{
			name: "overlapping interval - outside range",
			fields: fields{
				FromHour: 18,
				ToHour:   2,
				clock: &testClock{
					ret: time.Date(2023, time.Month(2), 21, 4, 10, 30, 0, time.UTC),
				},
			},
			want: false,
		},
		{
			name: "outside interval",
			fields: fields{
				FromHour: 18,
				ToHour:   2,
				clock: &testClock{
					ret: time.Date(2023, time.Month(2), 21, 14, 10, 30, 0, time.UTC),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &WindowedPrecondition{
				FromHour: tt.fields.FromHour,
				ToHour:   tt.fields.ToHour,
				clock:    tt.fields.clock,
			}
			if got := c.PerformCheck(); got != tt.want {
				t.Errorf("PerformCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWindowPreconditionFromMap(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		want    *WindowedPrecondition
		wantErr bool
	}{
		{
			name:    "empty map",
			args:    map[string]any{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success",
			args: map[string]any{
				"from": 12.,
				"to":   8.,
			},
			want: &WindowedPrecondition{
				FromHour: 12,
				ToHour:   8,
				clock:    &realClock{},
			},
			wantErr: false,
		},
		{
			name: "wrong args",
			args: map[string]any{
				"from": 12.,
				"to":   12.,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing from",
			args: map[string]any{
				"to": 12.,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing to",
			args: map[string]any{
				"from": 12.,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wrong type",
			args: map[string]any{
				"from": "12",
				"to":   23,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WindowPreconditionFromMap(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("WindowPreconditionFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WindowPreconditionFromMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}
