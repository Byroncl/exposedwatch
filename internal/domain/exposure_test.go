package domain

import "testing"

func TestExposureString(testRun *testing.T) {
	want := map[Exposure]string{Loopback: "loopback", Private: "private", AllInterfaces: "all-interfaces", Public: "public"}
	for exposure, name := range want {
		if exposure.String() != name {
			testRun.Errorf("%d.String() = %q, want %q", exposure, exposure.String(), name)
		}
	}
}
