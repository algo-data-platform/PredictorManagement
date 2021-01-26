package util

import (
	"testing"
)

func TestPrettyTime(t *testing.T) {
	tables := []struct {
		Second int64
		Res    string
	}{
		{
			30,
			"30 sec",
		},
		{
			2000,
			"33.3 min",
		},
		{
			20000,
			"5.6 hour",
		},
		{
			200000,
			"2.3 day",
		},
	}

	for _, table := range tables {
		res := PrettyTime(table.Second)
		if res != table.Res {
			t.Errorf("TestPrettyTime(%v) failed, got: %v, want: %v.",
				table.Second, res, table.Res)
		}
	}
}

func TestGetTimestampInterval(t *testing.T) {
	tables := []struct {
		timestamp1 string
		timestamp2 string
		interval   int64
	}{
		{
			"20201020_100000",
			"20201020_100001",
			1,
		},
		{
			"20201020_100000",
			"20201020_100020",
			20,
		},
		{
			"20201020_100000",
			"20201020_100200",
			120,
		},
		{
			"20201020_110000",
			"20201020_100000",
			3600,
		},
		{
			"20201020_11000",
			"20201020_10000",
			0,
		},
		{
			"20201020_110000",
			"20201020_",
			0,
		},
		{
			"22222aaaa",
			"20201020_",
			0,
		},
	}

	for _, table := range tables {
		interval := GetTimestampInterval(table.timestamp1, table.timestamp2)
		if interval != table.interval {
			t.Errorf("TestGetTimestampInterval(%v, %v) failed, got: %v, want: %v.",
				table.timestamp1, table.timestamp2, interval, table.interval)
		}
	}
}

func TestSecondsToHM(t *testing.T) {
	tables := []struct {
		seconds  string
		formatHM string
	}{
		{
			"10",
			"0小时0分钟",
		},
		{
			"120",
			"0小时2分钟",
		},
		{
			"7200",
			"2小时0分钟",
		},
		{
			"86400",
			"24小时0分钟",
		},
		{
			"0",
			"7日无更新",
		},
	}

	for _, table := range tables {
		formatHM := SecondsToHM(table.seconds)
		if formatHM != table.formatHM {
			t.Errorf("TestGetTimestampInterval(%v) failed, got: %v, want: %v.",
				table.seconds, formatHM, table.formatHM)
		}
	}
}
