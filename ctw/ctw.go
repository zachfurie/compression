package ctw

import (
	"fmt"
	"log"
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
	left  *node   //adds 1 to code
	right *node   //adds 0 to code
	c0    int     //count of 0s
	c1    int     //count of 1s
	p     float64 //weighted probability that next bit is 1
	d     int     //depth
	//parent *node   //possibly unneccessary
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

// Krichevsky–Trofimov estimator.
// Recursively update probabilities of all nodes by calling this func on the root node.
func updateProb(n *node, update uint8) {
	newP := 0.0
	if update == 0 {
		newP = n.p * (float64(n.c0) + 0.5) / (float64(n.c0) + float64(n.c1) + 1.0)
	} else if update == 1 {
		newP = n.p * (float64(n.c1) + 0.5) / (float64(n.c0) + float64(n.c1) + 1.0)
	}
	if newP != 0 {
		//fmt.Println(newP)
	}

	if n.d == Depth {
		n.p = newP
	} else {
		updateProb(n.left, update)
		updateProb(n.right, update)
		n.p = 0.5*newP + 0.5*n.left.p*n.right.p
	}
}

// Update prediction data for suffix nodes.
func updateCount(n *node, update uint8) {
	if n.d == Depth {
		if update == 0 {
			n.c0++
		} else {
			n.c1++
		}
	} else {
		bit := update & uint8(128)
		newUpdate := update << 1
		if bit == 0 {
			updateCount(n.right, newUpdate)
			n.c0 = n.left.c0 + n.right.c0
		} else {
			updateCount(n.left, newUpdate)
			n.c1 = n.left.c1 + n.right.c1
		}
	}
}

// All probabilities should be initialized to 1.
func initializeNodes(d int, code uint8) *node {
	newNode := node{code: code, c0: 0, c1: 0, d: d, p: 1.0}
	if d < Depth {
		rcode := code << 1
		lcode := rcode | uint8(1)
		newNode.left = initializeNodes(d+1, lcode)
		newNode.right = initializeNodes(d+1, rcode)
		newNode.p = newNode.left.p + newNode.right.p
	}
	return &newNode
}

// PROBLEM: PROBABILITIES ARE TOO SMALL TO REPRESENT WITH FLOAT64. NEED TO MANUALLY WRITE THEM INTO BYTES
// SOLUTION: USE ASSYMETRIC NUMBER SYSTEMS ENCODING https://en.wikipedia.org/wiki/Asymmetric_numeral_systems
func Encode(fp string, op string) {
	bytes, err := os.ReadFile(fp)
	if err != nil {
		log.Fatal(err)
	}
	bytes = bytes[:1000]
	length := len(bytes)
	llength := length / 20

	// B( empty_sequence | window) := 0 , where B(x) = # of bits needed to encode x
	// B is related to interval for window
	// I'm guessing its needed for decoding
	B := float64(0)
	interval := make([]float64, 2)
	interval[0] = B
	interval[1] = 1
	root := initializeNodes(0, uint8(0))
	recCheck(root, []int{})

	low := 0.0
	high := 1.0

	for i, bt := range bytes {
		bits := getBits(bt)
		for _, bit := range bits {
			// Do you update probabilities before or after encoding step?
			// Probably doesnt matter as long as you do it the same order in decoder
			// ACTUALLY, probably need to encode before updating, since you wont know the bit you are updating on until you decode it.
			updateWin(bit)
			updateCount(root, window)
			testy := root.c1
			test := root.p
			updateProb(root, bit)
			if testy < root.c1 {
				low = low + (root.p * (high - low))
				showIntervals(test, root.p)
			} else {
				high = high - (root.p * (high - low))
				showIntervals(root.p, test)
			}
		}
		if i%llength == 0 {
			cnt := 5 * i / llength
			fmt.Println(cnt, "%")
		}
	}
	fmt.Println(100, "%")

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
	fmt.Println("PROB: ", root.p)

	recCheck(root, []int{})

	return

	fmt.Println(root, B, bytes) // get rid of red lines for unused variables
}

//Weighted coding distribution = root.p

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

// As of now, this mist be called by update function.
// Could instead maybe implement this as a goroutine so it can just be called once from encode func.
func showIntervals(l float64, h float64) {
	low := int(l * 100)
	high := int(h * 100)
	if low < 1 && high < 1 {
		return
	}
	fmt.Println("")
	fmt.Print("|")
	for x := range [100]byte{} {
		if x == low {
			fmt.Print("[")
			continue
		}
		if x > low && x < high {
			fmt.Print(" ")
			continue
		}
		if x == high {
			fmt.Print("]")
			continue
		}
		fmt.Print("-")
	}
	fmt.Println("|")
	fmt.Println(low, "       ", high)
	fmt.Println(" ")
}
