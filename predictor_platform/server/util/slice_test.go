package util

import (
	"reflect"
	"sort"
	"testing"
)

func TestIsSubSliceString(t *testing.T) {
	tables := []struct {
		sliceA []string
		sliceB []string
		res    bool
	}{
		{
			[]string{"a", "b", "c"},
			[]string{"b"},
			false,
		},
		{
			[]string{"b", "c"},
			[]string{"b", "c", "a"},
			true,
		},
		{
			[]string{"b", "c", "a"},
			[]string{"a", "b", "c"},
			true,
		},
		{
			[]string{},
			[]string{"b", "c"},
			true,
		},
		{
			[]string{"d", "e"},
			[]string{"c", "b", "d", "c", "a"},
			false,
		},
	}

	for _, table := range tables {
		res := IsSubSliceString(table.sliceA, table.sliceB)
		if res != table.res {
			t.Errorf("TestIsSubSliceString(%v, %v) failed, got: %v, want: %v.",
				table.sliceA, table.sliceB, res, table.res)
		}
	}
}

func TestDelSliceFirstItem(t *testing.T) {
	tables := []struct {
		slice    []string
		item     string
		resSlice []string
	}{
		{
			[]string{"a", "b", "c"},
			"b",
			[]string{"a", "c"},
		},
		{
			[]string{"a", "b", "c"},
			"a",
			[]string{"b", "c"},
		},
		{
			[]string{"b", "c"},
			"c",
			[]string{"b"},
		},
		{
			[]string{},
			"c",
			[]string{},
		},
		{
			[]string{"b"},
			"b",
			[]string{},
		},
		{
			[]string{"a", "b", "c"},
			"d",
			[]string{"a", "b", "c"},
		},
	}

	for _, table := range tables {
		reqSlice := make([]string, len(table.slice))
		copy(reqSlice, table.slice)
		DelSliceFirstItem(&table.slice, table.item)
		sort.Strings(table.slice)
		sort.Strings(table.resSlice)
		if !reflect.DeepEqual(table.slice, table.resSlice) {
			t.Errorf("TestDelSliceFirstItem(%v, %v) failed, got: %v, want: %v.",
				reqSlice, table.item, table.slice, table.resSlice)
		}
	}
}

func TestExcludeSliceString(t *testing.T) {
	tables := []struct {
		slice          []string
		toExcludeSlice []string
		resSlice       []string
	}{
		{
			[]string{"a", "b", "c"},
			[]string{"b"},
			[]string{"a", "c"},
		},
		{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
			[]string{},
		},
		{
			[]string{"a", "b", "c"},
			[]string{"c", "a", "d"},
			[]string{"b"},
		},
		{
			[]string{},
			[]string{"c"},
			[]string{},
		},
		{
			[]string{"b"},
			[]string{},
			[]string{"b"},
		},
		{
			[]string{"b", "b", "a", "a"},
			[]string{"a"},
			[]string{"b", "b"},
		},
	}

	for _, table := range tables {
		resSlice := ExcludeSliceString(table.slice, table.toExcludeSlice)
		if !reflect.DeepEqual(resSlice, table.resSlice) {
			t.Errorf("TestExcludeSliceString(%v, %v) failed, got: %v, want: %v.",
				table.slice, table.toExcludeSlice, resSlice, table.resSlice)
		}
	}
}

func TestDiffSliceString(t *testing.T) {
	tables := []struct {
		sliceA   []string
		sliceB   []string
		resSlice []string
	}{
		{
			[]string{"a", "b", "c"},
			[]string{"b"},
			[]string{"a", "c"},
		},
		{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
			[]string{},
		},
		{
			[]string{"a", "b", "c"},
			[]string{"c", "a", "d"},
			[]string{"b"},
		},
		{
			[]string{},
			[]string{"c"},
			[]string{},
		},
		{
			[]string{"b"},
			[]string{},
			[]string{"b"},
		},
		{
			[]string{"b", "b", "a", "a"},
			[]string{"a"},
			[]string{"b", "b"},
		},
	}

	for _, table := range tables {
		resSlice := DiffSliceString(table.sliceA, table.sliceB)
		if !reflect.DeepEqual(resSlice, table.resSlice) {
			t.Errorf("TestDiffSliceString(%v, %v) failed, got: %v, want: %v.",
				table.sliceA, table.sliceB, resSlice, table.resSlice)
		}
	}
}
