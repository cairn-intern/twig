package evm

import (
	"math/big"
	"testing"
)

func TestABIStringEncodeDecode(t *testing.T) {
	orig := "testname.twig"
	encoded := EncodeStringParameter(orig)

	decoded, err := DecodeString(encoded, 0)
	if err != nil {
		t.Fatalf("failed to decode string: %v", err)
	}

	if decoded != orig {
		t.Fatalf("expected %s, got %s", orig, decoded)
	}
}

func TestABIDecodeValues(t *testing.T) {
	data := make([]byte, 64)
	data[31] = 1 // bool true or uint 1
	copy(data[44:64], []byte("01234567890123456789"))

	val := DecodeUint256(data, 0)
	if val.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected 1, got %v", val)
	}

	b := DecodeBool(data, 0)
	if !b {
		t.Fatalf("expected true")
	}

	addr := DecodeAddress(data, 32)
	if len(addr) != 42 {
		t.Fatalf("invalid address format: %s", addr)
	}
}
