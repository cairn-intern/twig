package evm

import (
	"encoding/binary"
	"math/bits"
)

var rc = [24]uint64{
	0x0000000000000001, 0x0000000000008082, 0x800000000000808a,
	0x8000000080008000, 0x000000000000808b, 0x0000000080000001,
	0x8000000080008081, 0x8000000000008009, 0x000000000000008a,
	0x0000000000000088, 0x0000000080008009, 0x000000008000000a,
	0x000000008000808b, 0x800000000000008b, 0x8000000000008089,
	0x8000000000008003, 0x8000000000008002, 0x8000000000000080,
	0x000000000000800a, 0x800000008000000a, 0x8000000080008081,
	0x8000000000008080, 0x0000000080000001, 0x8000000080008008,
}

var rhoOffsets = [25]uint{
	0, 1, 62, 28, 27,
	36, 44, 6, 55, 20,
	3, 10, 43, 25, 39,
	41, 45, 15, 21, 8,
	18, 2, 61, 56, 14,
}

var pi = [25]int{
	0, 10, 20, 5, 15,
	16, 1, 11, 21, 6,
	7, 17, 2, 12, 22,
	23, 8, 18, 3, 13,
	14, 24, 9, 19, 4,
}

func keccakF1600(a *[25]uint64) {
	for round := 0; round < 24; round++ {
		// Theta
		var c [5]uint64
		for x := 0; x < 5; x++ {
			c[x] = a[x] ^ a[x+5] ^ a[x+10] ^ a[x+15] ^ a[x+20]
		}
		var d [5]uint64
		for x := 0; x < 5; x++ {
			d[x] = c[(x+4)%5] ^ bits.RotateLeft64(c[(x+1)%5], 1)
		}
		for i := 0; i < 25; i++ {
			a[i] ^= d[i%5]
		}

		// Rho & Pi
		var b [25]uint64
		for i := 0; i < 25; i++ {
			b[pi[i]] = bits.RotateLeft64(a[i], int(rhoOffsets[i]))
		}

		// Chi
		for y := 0; y < 25; y += 5 {
			for x := 0; x < 5; x++ {
				a[y+x] = b[y+x] ^ (^b[y+(x+1)%5] & b[y+(x+2)%5])
			}
		}

		// Iota
		a[0] ^= rc[round]
	}
}

// Keccak256 computes the Ethereum Keccak-256 hash of data (padding 0x01).
func Keccak256(data []byte) []byte {
	rate := 136 // 1088 bits = 136 bytes for Keccak-256
	var state [25]uint64

	// Absorb
	p := 0
	for p+rate <= len(data) {
		for i := 0; i < rate/8; i++ {
			state[i] ^= binary.LittleEndian.Uint64(data[p+i*8 : p+(i+1)*8])
		}
		keccakF1600(&state)
		p += rate
	}

	// Pad: Keccak-256 uses 0x01 padding (unlike SHA3 which uses 0x06)
	rem := data[p:]
	block := make([]byte, rate)
	copy(block, rem)
	block[len(rem)] = 0x01
	block[rate-1] |= 0x80

	for i := 0; i < rate/8; i++ {
		state[i] ^= binary.LittleEndian.Uint64(block[i*8 : (i+1)*8])
	}
	keccakF1600(&state)

	// Squeeze 32 bytes
	out := make([]byte, 32)
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(out[i*8:(i+1)*8], state[i])
	}
	return out
}
