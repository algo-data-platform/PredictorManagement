package logics

import (
	"server/conf"
	"server/env"
	"testing"
)

func TestGetResetWeight(t *testing.T) {
	env.Env = &env.Environment{
		Conf: &conf.Conf{
			LoadThreshold: conf.LoadThreshold{
				Down_Gap: 100,
			},
		},
	}
	// case 1
	{
		coreNum := 0
		expectWeight := 100
		resetWeight := GetResetWeight(coreNum)
		if resetWeight != expectWeight {
			t.Errorf("TestGetResetWeight() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				resetWeight, expectWeight)
		}
	}
	// case 2
	{
		coreNum := 16
		expectWeight := 100
		resetWeight := GetResetWeight(coreNum)
		if resetWeight != expectWeight {
			t.Errorf("TestGetResetWeight() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				resetWeight, expectWeight)
		}
	}
	// case 3
	{
		coreNum := 32
		expectWeight := 200
		resetWeight := GetResetWeight(coreNum)
		if resetWeight != expectWeight {
			t.Errorf("TestGetResetWeight() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				resetWeight, expectWeight)
		}
	}
	// case 4
	{
		coreNum := 8
		expectWeight := 50
		resetWeight := GetResetWeight(coreNum)
		if resetWeight != expectWeight {
			t.Errorf("TestGetResetWeight() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				resetWeight, expectWeight)
		}
	}
	// case 5
	{
		coreNum := 64
		expectWeight := 400
		resetWeight := GetResetWeight(coreNum)
		if resetWeight != expectWeight {
			t.Errorf("TestGetResetWeight() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				resetWeight, expectWeight)
		}
	}
}
