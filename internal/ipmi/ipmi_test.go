package ipmi

import "testing"

func TestPowerArgs(t *testing.T) {
	if _, e := PowerArgs("on"); e != nil {
		t.Fatal(e)
	}
	if _, e := PowerArgs("destroy"); e == nil {
		t.Fatal("expected validation")
	}
}
func TestBootArgs(t *testing.T) {
	a, e := BootArgs("pxe", false, true)
	if e != nil || len(a) != 4 {
		t.Fatalf("%v %v", a, e)
	}
	if _, e := BootArgs("usb", false, false); e == nil {
		t.Fatal("expected validation")
	}
}
