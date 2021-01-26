package service

import (
	"reflect"
	"server/util"
	"testing"
)

func TestGetToAllocateMap(t *testing.T) {
	tables := []struct {
		totalNum            int
		allocateConfig      []util.SidHostNum
		allocatedSidHostNum []util.SidHostNum
		expect_res          map[uint]int
	}{
		{
			0,
			[]util.SidHostNum{
				util.SidHostNum{Sid: 2, HostNum: 3},
				util.SidHostNum{Sid: 3, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 1},
			},
			[]util.SidHostNum{},
			map[uint]int{},
		},
		{
			1,
			[]util.SidHostNum{
				util.SidHostNum{Sid: 2, HostNum: 3},
				util.SidHostNum{Sid: 3, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 1},
			},
			[]util.SidHostNum{},
			map[uint]int{2: 1},
		},
		{
			5,
			[]util.SidHostNum{
				util.SidHostNum{Sid: 2, HostNum: 3},
				util.SidHostNum{Sid: 3, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 1},
			},
			[]util.SidHostNum{},
			map[uint]int{2: 3, 3: 2},
		},
		{
			5,
			[]util.SidHostNum{
				util.SidHostNum{Sid: 2, HostNum: 3},
				util.SidHostNum{Sid: 3, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 1},
			},
			[]util.SidHostNum{
				util.SidHostNum{Sid: 2, HostNum: 3},
				util.SidHostNum{Sid: 3, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 0},
			},
			map[uint]int{2: 2, 3: 2, 4: 1},
		},
		{
			20,
			[]util.SidHostNum{
				util.SidHostNum{Sid: 2, HostNum: 3},
				util.SidHostNum{Sid: 3, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 1},
			},
			[]util.SidHostNum{},
			map[uint]int{2: 10, 3: 7, 4: 3},
		},
		{
			20,
			[]util.SidHostNum{
				util.SidHostNum{Sid: 2, HostNum: 3},
				util.SidHostNum{Sid: 3, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 1},
			},
			[]util.SidHostNum{
				util.SidHostNum{Sid: 3, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 3},
				util.SidHostNum{Sid: 5, HostNum: 3},
			},
			map[uint]int{2: 13, 3: 7},
		},
		{
			20,
			[]util.SidHostNum{
				util.SidHostNum{Sid: 2, HostNum: 3},
				util.SidHostNum{Sid: 3, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 1},
			},
			[]util.SidHostNum{
				util.SidHostNum{Sid: 2, HostNum: 2},
				util.SidHostNum{Sid: 4, HostNum: 3},
				util.SidHostNum{Sid: 5, HostNum: 3},
			},
			map[uint]int{2: 11, 3: 9},
		},
		{
			18,
			[]util.SidHostNum{
				util.SidHostNum{Sid: 1, HostNum: 3},
				util.SidHostNum{Sid: 2, HostNum: 2},
				util.SidHostNum{Sid: 3, HostNum: 1},
			},
			[]util.SidHostNum{
				util.SidHostNum{Sid: 1, HostNum: 13},
				util.SidHostNum{Sid: 2, HostNum: 9},
				util.SidHostNum{Sid: 3, HostNum: 3},
			},
			map[uint]int{1: 9, 2: 6, 3: 3},
		},
		{
			18,
			[]util.SidHostNum{
				util.SidHostNum{Sid: 1, HostNum: 0},
				util.SidHostNum{Sid: 2, HostNum: 0},
				util.SidHostNum{Sid: 3, HostNum: 0},
			},
			[]util.SidHostNum{
				util.SidHostNum{Sid: 1, HostNum: 13},
				util.SidHostNum{Sid: 2, HostNum: 9},
				util.SidHostNum{Sid: 3, HostNum: 3},
			},
			map[uint]int{1: 2, 2: 6, 3: 10},
		},
	}

	for _, table := range tables {
		res := getToAllocateMap(table.totalNum, table.allocateConfig, table.allocatedSidHostNum)
		if !reflect.DeepEqual(res, table.expect_res) {
			t.Errorf("TestGetToAllocateMap(%v, %v, %v) failed, got: res=%v, want: expect_res=%v.",
				table.totalNum, table.allocateConfig, table.allocatedSidHostNum, res, table.expect_res)
		}
	}
}
