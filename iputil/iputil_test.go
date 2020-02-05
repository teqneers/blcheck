package iputil

import (
	"testing"
)

func TestReverseIP(t *testing.T) {
	testIP := "1.2.3.4"
	reverseIP := Reverse(testIP)

	if reverseIP != "4.3.2.1" {
		t.Errorf("Reversing ip failed: got %v want %v", reverseIP, "4.3.2.1")
	}
}

func TestValidateIP(t *testing.T) {
	testIP := "1.2.3.4"
	isValid := ValidateIP(testIP)

	if isValid == false {
		t.Errorf("Validating IP failed, IP is invalid but should be valid")
	}
}

func TestValidateIPButInvalid(t *testing.T) {
	testIP := "300.2.3.4"
	isValid := ValidateIP(testIP)

	if isValid == true {
		t.Errorf("Validating IP failed, IP is valid but should be invalid")
	}
}

func TestValidateIPButTooLong(t *testing.T) {
	testIP := "1.2.3.4.5"
	isValid := ValidateIP(testIP)

	if isValid == true {
		t.Errorf("Validating IP failed, IP is valid but should be invalid")
	}
}
