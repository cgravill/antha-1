package Optimization

import (
	"testing"
)

func TestEndsOK_1(t *testing.T) {
	ends := make([][]string, 4)

	ends[0] = []string{"A"}
	ends[1] = []string{"C"}
	ends[2] = []string{"E"}
	ends[3] = []string{"G"}

	if !endsOK(ends, make(map[string]bool)) {
		t.Errorf("OK ends should be judged OK")
	}
}

func TestEndsOK_2(t *testing.T) {
	ends := make([][]string, 4)

	ends[0] = []string{"A"}
	ends[1] = []string{"D"}
	ends[2] = []string{"F"}
	ends[3] = []string{"A"}

	if endsOK(ends, make(map[string]bool)) {
		t.Errorf("non-OK ends should be judged non-OK")
	}
}

func TestEndsOK_3(t *testing.T) {
	ends := make([][]string, 4)

	ends[0] = []string{"A"}
	ends[1] = []string{"A"}
	ends[2] = []string{"F"}
	ends[3] = []string{"H"}

	if endsOK(ends, make(map[string]bool)) {
		t.Errorf("non-OK ends should be judged non-OK")
	}
}

func TestEndsOK_4(t *testing.T) {
	ends := make([][]string, 4)

	ends[0] = []string{"A"}
	ends[1] = []string{"B"}
	ends[2] = []string{"H"}
	ends[3] = []string{"H"}

	if endsOK(ends, make(map[string]bool)) {
		t.Errorf("non-OK ends should be judged non-OK")
	}
}
