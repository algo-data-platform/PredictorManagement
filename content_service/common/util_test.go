package common

import (
	"reflect"
	"testing"
	"math"
)

func TestFindComplement(t *testing.T) {
	tables := []struct {
		src  []string
		dest []string
		res  []string
	}{
		{[]string{"a"}, []string{"b"}, []string{"a"}},
		{[]string{"a", "b"}, []string{"b"}, []string{"a"}},
		{[]string{"a", "b"}, []string{"b", "c"}, []string{"a"}},
		{[]string{"a", "b"}, []string{"b", "c", "a"}, []string{}},
		{[]string{"a", "b", "c", "d"}, []string{"c", "a"}, []string{"b", "d"}},
	}

	for _, table := range tables {
		res := FindComplement(table.src, table.dest)
		if !reflect.DeepEqual(res, table.res) {
			t.Errorf("TestFindComplement(%v, %v) failed, got: %v, want: %v.",
				table.src, table.dest, res, table.res)
		}
	}
}

func TestDivideSlices(t *testing.T) {
	tables := []struct {
		s         []string
		num       int
		expectRes [][]string
	}{
		{
			[]string{"1", "2", "3", "4", "5"},
			2,
			[][]string{
				[]string{"1", "2", "3"},
				[]string{"4", "5"},
			},
		},
		{
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"},
			3,
			[][]string{
				[]string{"1", "2", "3", "4", "5", "6", "7", "8"},
				[]string{"9", "10", "11", "12", "13", "14", "15", "16"},
				[]string{"17", "18", "19", "20", "21", "22", "23"},
			},
		},
		{
			[]string{"1", "2", "3", "4", "5"},
			0,
			[][]string{},
		},
		{
			[]string{},
			3,
			[][]string{},
		},
	}

	for _, table := range tables {
		res := DivideSlices(table.s, table.num)
		if !reflect.DeepEqual(res, table.expectRes) {
			t.Errorf("DivideSlices(%v, %v) failed, got: %v, want: %v.",
				table.s, table.num, res, table.expectRes)
		}
	}
}

func TestGetIndexFromChildList(t *testing.T) {
	tables := []struct {
		item       string
		childList [][]string
		expectRes  bool
		expectPos  int
	}{
		{
			"3",
			[][]string{
				[]string{"1", "2", "3"},
				[]string{"4", "5"},
			},
			true,
			0,
		},
		{
			"23",
			[][]string{
				[]string{"1", "2", "3", "4", "5", "6", "7", "8"},
				[]string{"9", "10", "11", "12", "13", "14", "15", "16"},
				[]string{"17", "18", "19", "20", "21", "22", "23"},
			},
			true,
			2,
		},
		{
			"9",
			[][]string{
				[]string{"1", "2", "3"},
				[]string{"4", "5"},
			},
			false,
			0,
		},
		{
			"9",
			[][]string{},
			false,
			0,
		},
	}

	for _, table := range tables {
		res, pos := GetIndexFromChildList(table.item, table.childList)
		if res != table.expectRes || pos != table.expectPos {
			t.Errorf("GetIndexFromChildList(%v, %v) failed, res: %v, expectRes: %v, pos: %v, expectPos: %v.",
				table.item, table.childList, res, table.expectRes, pos, table.expectPos)
		}
	}
}

func TestIsFloatVectorCosineSim(t *testing.T) {
	tables := []struct {
		vec_a  []float64
		vec_b []float64
		expSim  float64
	}{
		{[]float64{0.5, 0.6, 0.7, 0.8}, []float64{0.5, 0.6, 0.7, 0.8}, 1},
		{[]float64{0.5, 0.6, 0.7, 0.8}, []float64{0.55, 0.6, 0.7, 0.8}, 0.999402},
		{[]float64{0.5, 0.6, 0.7, 0.8}, []float64{0.5, 0.66, 0.7, 0.8}, 0.999213},
		{[]float64{0.5, 0.6, 0.77, 0.8}, []float64{0.5, 0.6, 0.7, 0.8}, 0.999044},
		{[]float64{0.5, 0.6, 0.7, 0.88}, []float64{0.5, 0.6, 0.7, 0.8}, 0.998920},
	}
	for _, table := range tables {
		is_similar, sim := IsFloatVectorCosineSim(table.vec_a, table.vec_b)
		if !is_similar && (math.Abs(sim-table.expSim)>1e-6) {
			t.Errorf("IsFloatVectorCosineSim(%v, %v) failed, sim: %v, expSim: %v.",
				table.vec_a, table.vec_b, sim, table.expSim)
		}
	}
}
