package ctw

import (
	"bufio"
	// ops "compression/ops"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
)

// Based on this paper: https://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.14.352&rep=rep1&type=pdf

// Keep a window of the previous n bits.
// For any sequence of n bits, store the number of 1s and 0s that immediately follow
// any occurence of the sequence in the data that has been read so far (Context Tree).
// Use this to predict the next bit in the data.

// Set of suffixes S.
// Proper - no suffix is a suffix of another suffix.
// Complete - every semi-infinite string has a unique suffix in S.
// Each suffix s in S has a corresponding parameter, which is a value within [0, 1].
// Parameter specifies the distribution over {0, 1}.

// Suffix function maps semi-infinite sequences onto their corresponding suffix s in S.
// Suffix function tells the parameter for generating the next binary digit of the sequence.

// "Model" is equivalent to suffix set. All sequences that share a suffix set S are said to share a model.
// The set of all suffix sets not containing suffixes longer than D is called "model class C_D"

// Parameter represents the chance the next symbol will be a 1.
// Thus, the chance the next bits in the sequence with parameter p will have x 0s and y 1s is:
//     (1-p)^x * p^y

// Context Tree:
// Each node in context tree T_D has a binary string with length <= D.
// Nodes with length == D are leaf nodes.

var Depth = 3
var window = uint8(0)

type node struct {
	code  uint8
	left  *node      //adds 1 to code
	right *node      //adds 0 to code
	c0    *big.Float //count of 0s
	c1    *big.Float //count of 1s
	d     int        //depth
	p     *big.Float //weighted probability of a sequence with c0 "0"s and c1 "1"s
	kt    *big.Float // Krichevsky–Trofimov estimate for p(x=0)
} // In practice, probably don't need kt0 or kt1, but I have them there for now so I can graph them.

// Error Check
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// pop off oldest bit in window, add a new bit from source data
func updateWin(bit uint8) {
	window <<= 1
	window |= bit
}

// Get 8 bits from a byte (bits are represented by bytes with either one or zero nonzero bits)
func getBits(bt byte) []uint8 {
	bits := make([]uint8, 8)
	bits[7] = bt & uint8(1)
	bits[6] = bt & uint8(2)
	bits[5] = bt & uint8(4)
	bits[4] = bt & uint8(8)
	bits[3] = bt & uint8(16)
	bits[2] = bt & uint8(32)
	bits[1] = bt & uint8(64)
	bits[0] = bt & uint8(128)
	for i := range bits {
		if bits[i] != uint8(0) {
			bits[i] = uint8(1)
		}
	}
	return bits
}

func getBitsReverse(bt byte) []uint8 {
	bits := make([]uint8, 8)
	bits[0] = bt & uint8(1)
	bits[1] = bt & uint8(2)
	bits[2] = bt & uint8(4)
	bits[3] = bt & uint8(8)
	bits[4] = bt & uint8(16)
	bits[5] = bt & uint8(32)
	bits[6] = bt & uint8(64)
	bits[7] = bt & uint8(128)
	for i := range bits {
		if bits[i] != uint8(0) {
			bits[i] = uint8(1)
		}
	}
	return bits
}

// Krichevsky–Trofimov estimator.
// Recursively update probabilities of all nodes by calling this func on the root node.
func updateProb(n *node, update uint8) {
	newP := big.NewFloat(0)
	if update&uint8(1) == 0 {
		x := big.NewFloat(0)
		y := big.NewFloat(0)
		newP.Quo((x.Add(n.c0, big.NewFloat(0.5))), y.Add(y.Add(n.c0, n.c1), big.NewFloat(1)))
		n.kt.Mul(n.kt, newP)
	} else {
		x := big.NewFloat(0)
		y := big.NewFloat(0)
		newP.Quo((x.Add(n.c1, big.NewFloat(0.5))), y.Add(y.Add(n.c0, n.c1), big.NewFloat(1)))
		n.kt.Mul(n.kt, newP)
	}
	if n.d == Depth {
		n.p = newP
	} else {
		bit := update & uint8(1)
		newUpdate := update >> 1
		if bit == 0 {
			updateProb(n.right, newUpdate)
		} else {
			updateProb(n.left, newUpdate)
		}
		x := big.NewFloat(0)
		y := big.NewFloat(0)
		n.p.Add(x.Mul(big.NewFloat(0.5), n.kt), y.Mul(y.Mul(n.left.p, n.right.p), big.NewFloat(0.5)))
	}
}

// func updateProbD(n *node, update uint8, a *big.Float, b *big.Float) uint8 {
// 	newP := big.NewFloat(0)
// 	if update == 0 {
// 		x := big.NewFloat(0)
// 		y := big.NewFloat(0)
// 		newP.Quo((x.Add(n.c0, big.NewFloat(0.5))), y.Add(y.Add(n.c0, n.c1), big.NewFloat(1)))
// 		n.kt.Mul(n.kt, newP)
// 	} else if update == 1 {
// 		x := big.NewFloat(0)
// 		y := big.NewFloat(0)
// 		newP.Quo((x.Add(n.c1, big.NewFloat(0.5))), y.Add(y.Add(n.c0, n.c1), big.NewFloat(1)))
// 		n.kt.Mul(n.kt, newP)
// 	}
// 	if n.d == Depth {

// 		n.p = newP
// 	} else {
// 		updateProb(n.left, update)
// 		updateProb(n.right, update)
// 		x := big.NewFloat(0)
// 		y := big.NewFloat(0)
// 		n.p.Add(x.Mul(big.NewFloat(0.5), n.kt), y.Mul(y.Mul(n.left.p, n.right.p), big.NewFloat(0.5)))
// 	}
// }

// Update prediction data for suffix nodes.
func updateCount(n *node, update uint8) {
	if n.d == Depth {
		if update&uint8(1) == 0 {
			n.c0.Add(n.c0, big.NewFloat(1))
		} else {
			n.c1.Add(n.c1, big.NewFloat(1))
		}
	} else {
		bit := update & uint8(1)
		newUpdate := update >> 1
		if bit == 0 {
			updateCount(n.right, newUpdate)
			n.c0.Add(n.c0, big.NewFloat(1))
		} else {
			updateCount(n.left, newUpdate)
			n.c1.Add(n.c1, big.NewFloat(1))
		}
	}
}

// All probabilities should be initialized to 1.
func initializeNodes(d int, code uint8) *node {
	newNode := node{code: code, c0: big.NewFloat(0), c1: big.NewFloat(0), d: d, p: big.NewFloat(1), kt: big.NewFloat(1)}
	if d < Depth {
		rcode := code << 1
		lcode := rcode | uint8(1)
		newNode.left = initializeNodes(d+1, lcode)
		newNode.right = initializeNodes(d+1, rcode)
	}
	return &newNode
}

// PROBLEM: PROBABILITIES ARE TOO SMALL TO REPRESENT WITH FLOAT64. NEED TO MANUALLY WRITE THEM INTO BYTES
// POSSIBLE SOLUTION: USE ASSYMETRIC NUMBER SYSTEMS ENCODING https://en.wikipedia.org/wiki/Asymmetric_numeral_systems
func Encode(fp string, op string) {
	bytes, err := os.ReadFile(fp)
	if err != nil {
		log.Fatal(err)
	}

	desiredLength := 1000
	if len(bytes) > desiredLength {
		bytes = bytes[:desiredLength]
	}

	length := len(bytes)
	llength := length / 20
	fmt.Println("Bytes: ", length)

	// B( empty_sequence | window) := 0 , where B(x) = # of bits needed to encode x
	// B is related to interval for window
	// I'm guessing its needed for decoding
	lowerBound := big.NewFloat(0)
	upperBound := big.NewFloat(1)

	// Initialize nodes and do a dummy update on a "0" bit
	root := initializeNodes(0, uint8(0))
	updateProb(root, uint8(0))
	updateCount(root, uint8(0))

	probsfile, err := os.Create("probs.txt")
	check(err)
	w := bufio.NewWriter(probsfile)

	ktprobsfile, err := os.Create("ktprobs.txt")
	check(err)
	w2 := bufio.NewWriter(ktprobsfile)

	lbfile, err := os.Create("lb.txt")
	check(err)
	w3 := bufio.NewWriter(lbfile)

	bdfile, err := os.Create("bytedata.txt")
	check(err)
	w4 := bufio.NewWriter(bdfile)

	totalp := big.NewFloat(0)

	for i, bt := range bytes {
		bits := getBits(bt)
		for _, bit := range bits {
			fmt.Fprintln(w, root.p)
			fmt.Fprintln(w2, upperBound)
			fmt.Fprintln(w3, lowerBound)
			fmt.Fprintln(w4, big.NewInt(int64(bit)))

			updateWin(bit)
			updateProb(root, bit)
			updateCount(root, window)
			if bit == 0 {
				lowerBound.Add(lowerBound, big.NewFloat(0).Mul(root.p, big.NewFloat(0).Add(upperBound, big.NewFloat(0).Mul(big.NewFloat(0), lowerBound))))
				//lowerBound.Add(lowerBound, root.p)
			} else if bit == 1 {
				upperBound.Add(upperBound, big.NewFloat(0).Mul(big.NewFloat(0).Mul(root.p, big.NewFloat(0).Add(upperBound, big.NewFloat(0).Mul(big.NewFloat(0), lowerBound))), big.NewFloat(-1)))
				//upperBound.Add(upperBound, big.NewFloat(0).Mul(big.NewFloat(-1), root.p))
			}
			totalp.Add(totalp, root.p)
		}
		// if i%llength == 0 {
		// 	cnt := 5 * i / llength
		// 	fmt.Println(cnt, "%")
		// }
		fmt.Print(lowerBound, "      | .    ", i, llength) // remove this
	}
	fmt.Println(100, "%")
	w.Flush()
	w2.Flush()
	w3.Flush()
	w4.Flush()

	fmt.Println("total p: ", totalp)

	// NOTE (8/20/22): Arithmetic encoding should return a SINGLE number representing the final probability value
	// Should also return the first x bits of the source data, where x=Depth. This is the information needed to get the decoder started.
	// Also need to return an INTERVAL of probabilities, not just one probability.

	// The decoder will start with their own blank tree and will update it with the x bits given by the encoder from the source data.
	// The decoder will then decode the next but as a 1 or a 0 depending on which option keeps the final interval given by the encoder inside the decoder's new window it is constructiong as it goes.

	// Proof:
	//
	// Base Case:
	//    Decoder knows the first D bits of the source data.
	//    Decoder constructs a new tree and initializes on the first D bits exactly how encoder would.
	//    Decoder knows the interval [i,j] (s.t. 0<=i<j<=1) that was the final result of the encoder's tree.
	//    Decoder keeps track of its own interval which is initially [0,1]
	//
	// Induction:
	//    Decoder gets P(x=0) from the root of its current tree and uses that probability to divide the current interval up into subintervals
	//                                                    0.0                   1.0
	//       Final interval returned by encoder:           |----[]---------------|
	//       Current interval and subintervals of decoder: |-[  0  | 1 ]---------|
	//    Decoder knows the next bit is a 0 because it is following in the footsteps of the encoder, and if the encoder ended up at that final interval, then it neccessarily must have chosen 0 at this interval.
	//    Knowing next bit is 0, decoder updates tree accordingly.
	//    If the interval containing the chosen option is equal to the encoder's final interval, decoding stops.
	//
	// By repeating inductive step, all bits will be decoded.

	//os.WriteFile(op,root.p (converted to byte array),os.ModeDevice)
	fmt.Println()
	fmt.Printf("INTERVAL: [%v, %v)\n", lowerBound, upperBound) //big.NewFloat(0).Add(lowerBound, root.p))

	// binaryCode := ops.Binary_expansion(lowerBound, big.NewFloat(0).Add(lowerBound, root.p), []uint8{})
	// fmt.Println()
	// fmt.Println(binaryCode, len(binaryCode))
	// fmt.Println()
	// ret, n := binaryToFloat(binaryCode)
	// fmt.Println(ret, n)
	// os.WriteFile(op, binaryCode, os.ModeDevice)

	fmt.Println("!!!", root.c0, root.left.c0, root.right.c0)

}

func binaryToFloat(bc []uint8) (*big.Float, *big.Float) {
	// neighbors := big.NewFloat(1)
	// ret := big.NewFloat(0)
	// for i, x := range bc {
	// 	neighbors.Quo(neighbors, big.NewFloat(10))
	// 	if x == 1 {
	// 		ret.Add(ret, big.NewFloat(math.Pow(2, -1.0*float64(i))))
	// 	}
	// }
	// return ret, neighbors
	a := big.NewFloat(0)
	b := big.NewFloat(1)
	for i, x := range bc {
		if x == 0 {
			b.Add(b, big.NewFloat(-1*(math.Pow(2, float64(-(i+1))))))
		} else {
			a.Add(a, big.NewFloat(math.Pow(2, float64(-(i+1)))))
		}
	}
	return a, b
}

func Decode(fp string, op string) {
	window = uint8(0)
	bytes, err := os.ReadFile(fp)
	if err != nil {
		log.Fatal(err)
	}
	lowerBound := big.NewFloat(0)
	// for i, b := range bytes {
	// 	if b == 0 {
	// 		bytes[i] = 1
	// 	}
	// 	if b == 1 {
	// 		bytes[i] = 0
	// 	}
	// }
	a, b := binaryToFloat(bytes)
	a = big.NewFloat(0.0223388672)
	b = big.NewFloat(0.0223388672)
	fmt.Println(bytes, a, b)
	// Initialize nodes and do a dummy update on a "0" bit
	root := initializeNodes(0, uint8(0))
	updateProb(root, uint8(0))
	updateCount(root, uint8(0))

	decfile, err := os.Create(op)
	check(err)
	w := bufio.NewWriter(decfile)

	bitsfile, err := os.Create("bitsfile.txt")
	check(err)
	w2 := bufio.NewWriter(bitsfile)

	decProbfile, err := os.Create("decprobs.txt")
	check(err)
	w3 := bufio.NewWriter(decProbfile)

	counter := 0
	for {
		counter = counter + 1
		if counter == 3 {
			w.Flush()
			w2.Flush()
			w3.Flush()
			os.Exit(0)
		}
		bits := byte(0)
		for y := 0; y < 8; y++ {
			bit := eval(lowerBound, root.p, a, b)
			fmt.Fprintln(w2, bit)
			fmt.Fprintln(w3, lowerBound, root.p)
			if bit == 2 {
				fmt.Println("EOF")
				w.Flush()
				w2.Flush()
				os.Exit(0)
			} else if bit == 1 {
				bits = bits | uint8(1)
				lowerBound.Add(lowerBound, root.p)
			} else if bit == 3 {
				fmt.Println("eval returned nobinary answer")
				w.Flush()
				w2.Flush()
				os.Exit(1)
			}
			updateWin(bit)
			updateProb(root, bit)
			updateCount(root, window)
			bits = bits << 1
		}
		fmt.Fprint(w, string(bits))

		// I dont think I should have to scale the p values since p should decrease exponentially...

	}

	fmt.Print(bytes, lowerBound, a, b)

	// Probably don't need to convert binary back to float

	// have a string of bits. 1 tells you its above 0.5, 0 means below 0.5
	// loop thru tree updates, where you take root.p and then see what will happen if you choose a 1 or a 0:
	// case 1: only one of the choices keeps you in the interval defined by the bit (either the top or bottom half)
	//     Choose that bit for the decoding output.
	//     Move to the next bit of the encoded interval.
	//     Adjust sizes (either make next encoded interval be half the length of the previous one, or make root.p interval double)
	//         - (whatever you do, will need a variable that is maintained throughout loops to keep track of how much scaling is needed, since it will always be an additional factor of 2)
	// case 2: both choices are within the interval
	//     You are done encoding
	// case 3: the interval overlaps both choices
	//     Do procedure from "special case" op.
	//     Need to expand the half

	// decodedData := []byte{}
	// scaler := 1
	// bitIndex := 0
	// for {
	// 	byter := make([]uint8, 8)
	// 	for i := range byter {
	// 		switch {
	// 		case condition:

	// 		}
	// 		byter[i] = ...

	// 	}
	// }

}

func eval(l *big.Float, p *big.Float, a *big.Float, b *big.Float) uint8 {
	h := big.NewFloat(0)
	h.Add(l, p)
	ret := uint8(3)
	// if l.Cmp(a) >= 0 { //&& h.Cmp(b) < 0
	// 	// EOF
	// 	ret = 2
	// } else ...
	if h.Cmp(b) >= 0 {
		ret = 1
	} else if h.Cmp(a) < 0 {
		ret = 0
	} else {
		//fmt.Printf("p: %v | a: %v | b: %v \n", p, a, b)
		ret = 1 //3
	}
	fmt.Printf("lo: %-20v | hi: %-20v | a: %-20v | b: %-20v \n bit: %-20v \n", l, h, a, b, ret)
	return ret

}

//------------------------------------------
//Bug Tests

// Get path of leafnodes in huffman tree
func recCheck(hufT *node, list []int) {
	if hufT.d == Depth {
		fmt.Println(&hufT, *hufT, list)
		return
	}
	fmt.Println(&hufT, *hufT, list)
	l := make([]int, len(list))
	copy(l, list)
	l = append(l, 1)
	r := make([]int, len(list))
	copy(r, list)
	r = append(r, 0)
	recCheck(hufT.left, l)
	recCheck(hufT.right, r)
}
