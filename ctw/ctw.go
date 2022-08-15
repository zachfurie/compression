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

//right is the child whos code is this code + 0, and left has code that is this code + 1
// c0 must be >= the sum of the c0 values of the node's children (and same with c1)

func updateWin(bit uint8) {
	window <<= 1
	window |= bit
}

// takes in a byte, returns two windows of length = 4 bits.
// window size should be depth + 1, so currently this only works for depth=3
// possibly wrong/useless
// probably wrong/useless
func nextWin(bt byte) (uint8, uint8) {
	win1 := bt >> 4
	win2 := bt & 8
	return uint8(win1), uint8(win2)
}

// Get bits, update
// also possibly wrong/useless
func processWin(win uint8) int {
	bits := make([]uint8, 4)
	bits[0] = win & uint8(1)
	bits[1] = win & uint8(2)
	bits[2] = win & uint8(4)
	bits[3] = win & uint8(8)

	return 0
}

// Maybe not wrong/useless?
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

	if n.d == Depth {
		n.p = newP
	} else {
		updateProb(n.left, update)
		updateProb(n.right, update)
		n.p = 0.5*newP + 0.5*n.left.p*n.right.p
	}
}

// Update prediction data for suffix nodes
func updateCount(n *node, update uint8) {
	if n.d == Depth {
		winMinusOne := window >> 1
		if n.code == winMinusOne {
			if update == 0 {
				n.c0++
			} else {
				n.c1++
			}
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

func initializeNodes(d int, code uint8) *node {
	newNode := node{code: code, c0: 0, c1: 0, d: d, p: 0.0}
	if d < Depth {
		rcode := code << 1
		lcode := rcode | uint8(1)
		newNode.left = initializeNodes(d+1, lcode)
		newNode.right = initializeNodes(d+1, rcode)
	}
	return &newNode
}

func Encode(fp string, op string) {
	bytes, err := os.ReadFile(fp)
	if err != nil {
		log.Fatal(err)
	}

	// B( empty_sequence | window) := 0 , where B(x) = # of bits needed to encode x
	// B is related to interval for window
	B := float64(0)
	interval := make([]float64, 2)
	interval[0] = B
	interval[1] = 1
	root := initializeNodes(0, uint8(0))
	encodedInts := []float64{} // by "ints" I mean "intervals" ;)

	// for _, bt := range bytes {
	// 	win1, win2 := nextWin(bt)
	// 	encodedInts = append(encodedInts, processWin(win1))
	// 	encodedInts = append(encodedInts, processWin(win2))
	// }

	for _, bt := range bytes {
		bits := getBits(bt)
		for _, bit := range bits {
			// Update window CHECK
			// Update counts CHECK (not really)
			// Update probabilities CHECK
			// Arithmetic encoding
			updateWin(bit)
			updateCount(root, window)
			updateProb(root, bit)

		}
	}

	// Implement Arithmetic encoding:

	// intervals on [0,1) are determined by root.p
	// - [0,1-p) is 0
	// - [1-p,1) is 1

	// Get last bit in window (suffix of sequence in window). Let that bit be b

	// Create new intervals within the previous ionterval that corresponded to b
	// (if first bit was 0, then create two new intervals within [0,1-p). )

	// repeat on remaining 2 bits in window (assuming Depth=3)

	// Resulting interval can then be arithmetically encoded.
	// Since a bigger interval requires less space to write, the more likely sequences will take up less space

	// -------

	// OR is the window supposed to keep shifting 1 bit at a time, thus always using an interval of an interval of an interval
	// This feels more right

	fmt.Println(root, B, bytes, encodedInts) // get rid of red lines for unused variables
}

//Weighted coding distribution = root.p
