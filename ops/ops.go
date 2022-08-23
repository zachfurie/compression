package ops

import (
	"math/big"
)

// lower, split, upper, interval [a,b]
// l *big.Float, s *big.Float, u *big.Float,
func Binary_expansion(a *big.Float, b *big.Float, ret []uint8) []uint8 {
	// a is lower bound of target interval
	// b is upper bound of target interval
	// |----0----|----1----|
	// 0.0       0.5       1.0
	//
	switch {
	case a.Cmp(big.NewFloat(0.5)) == 1:
		// |----0----|----1-[]-|
		// 0.0       0.5    ab 1.0
		// If a > 1/2:
		//     emit 1
		//     a,b = 2(a,b - 1/2)
		ret = append(ret, 1)
		return Binary_expansion(big.NewFloat(1).Mul(big.NewFloat(1).Add(a, big.NewFloat(-0.5)), big.NewFloat(2.0)), big.NewFloat(1).Mul(big.NewFloat(1).Add(b, big.NewFloat(-0.5)), big.NewFloat(2.0)), ret)
	case b.Cmp(big.NewFloat(0.5)) == -1:
		// |----0-[]-|----1----|
		// 0.0    ab 0.5       1.0
		// If b < 1/2:
		//     emit 0
		//     a,b = 2(a,b)
		ret = append(ret, 0)
		return Binary_expansion(big.NewFloat(1).Mul(a, big.NewFloat(2.0)), big.NewFloat(1).Mul(b, big.NewFloat(2.0)), ret)
	default:
		// |----0--[ |  ]-1----|
		// 0.0     a 0.5 b      1.0
		return special_case(a, b, ret)
	}
}

type block struct {
	low  *big.Float //inclusive
	high *big.Float //not inclusive
	bit  uint8
}

func special_case(a *big.Float, b *big.Float, ret []uint8) []uint8 {
	//recurse here

	//INCOMPLETE - need to adjust block sizes, adjust a and b based on contained blocks, and keep track of what to return.

	// |----0----|----0----|----1----|----1----|
	// 0.0       0.25      0.5       0.75      1.0
	//
	blocks := []block{{big.NewFloat(0), big.NewFloat(0.25), 0}, {big.NewFloat(0.25), big.NewFloat(0.5), 0}, {big.NewFloat(0.5), big.NewFloat(0.75), 1}, {big.NewFloat(0.75), big.NewFloat(1), 1}}
	contained := make([]block, 4)
	for _, x := range blocks {
		switch {
		case a.Cmp(x.high) >= 0 || b.Cmp(x.low) < 0:
			continue
		default:
			contained = append(contained, x)
		}
	}
	for {
	}
}
