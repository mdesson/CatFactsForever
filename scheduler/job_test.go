package scheduler

import (
	"fmt"
	"testing"
	"time"
)

func TestCronFieldCheck(t *testing.T) {
	tests := []struct {
		inputInput   string
		inputCompare int
		wantRunnable bool
		wantOk       bool
	}{
		{"*", 5, true, true},
		{"5", 5, true, true},
		{"4,5,6", 5, true, true},
		{"0-10", 5, true, true},
		{"0-10", 5, true, true},
	}
	for _, test := range tests {
		if gotRunnable, gotOk := cronFieldCheck(test.inputInput, test.inputCompare); gotRunnable != test.wantRunnable || gotOk != test.wantOk {
			t.Errorf("intCast(%q, %v) = %v, %v", test.inputInput, test.inputCompare, gotRunnable, gotOk)
		}
	}
}

func TestCronFieldCheckFailure(t *testing.T) {
	tests := []struct {
		inputInput   string
		inputCompare int
		wantRunnable bool
		wantOk       bool
	}{
		{"abc", 5, false, false},
		{"4,,6", 5, false, false},
		{"7,8,9", 5, false, true},
		{"5-4", 5, false, false},
		{"0-4", 5, false, true},
		{"4", 5, false, true},
	}
	for _, test := range tests {
		if gotRunnable, gotOk := cronFieldCheck(test.inputInput, test.inputCompare); gotRunnable != test.wantRunnable || gotOk != test.wantOk {
			t.Errorf("intCast(%q, %v) = %v, %v", test.inputInput, test.inputCompare, gotRunnable, gotOk)
		}
	}
}

func TestCanRun(t *testing.T) {
	now := time.Now()
	hour, min, _ := now.Clock()
	_, month, day := now.Date()
	weekday := now.Weekday()

	tests := []struct {
		input        string
		wantRunnable bool
		wantErr      error
	}{
		{"* * * * *", true, nil},
		{fmt.Sprintf("%v %v %v %v %v", min, hour, day, int(month), int(weekday)), true, nil},
	}

	for _, test := range tests {
		if gotRunnable, gotErr := canRun(test.input); gotRunnable != test.wantRunnable || gotErr != test.wantErr {
			t.Errorf("canRun(%q) = %v, %v", test.input, gotRunnable, gotErr)
		}
	}
}
