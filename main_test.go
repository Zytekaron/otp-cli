package main

import "testing"

func TestParseSizeBytes(t *testing.T) {
	bytes, err := parseSizeBytes("15")
	if err != nil {
		t.Errorf("expected error to be nil, but got %v", err)
	} else if bytes != 15 {
		t.Errorf("expected %d but got %d", 1, bytes)
	}

	bytes, err = parseSizeBytes("1M")
	if err != nil {
		t.Errorf("expected error to be nil, but got %v", err)
	} else if bytes != megabyte {
		t.Errorf("expected %d but got %d", megabyte, bytes)
	}

	bytes, err = parseSizeBytes("3 tb")
	if err != nil {
		t.Errorf("expected error to be nil, but got %v", err)
	} else if bytes != 3*terabyte {
		t.Errorf("expected %d but got %d", 3*terabyte, bytes)
	}

	bytes, err = parseSizeBytes("15 gib")
	if err != nil {
		t.Errorf("expected error to be nil, but got %v", err)
	} else if bytes != 15*gibibyte {
		t.Errorf("expected %d but got %d", 15*gibibyte, bytes)
	}
}
