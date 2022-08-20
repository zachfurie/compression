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
		fmt.Println(newP)
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
	B := float64(0)
	interval := make([]float64, 2)
	interval[0] = B
	interval[1] = 1
	root := initializeNodes(0, uint8(0))
	recCheck(root, []int{})
	//encodedInts := []float64{} // by "ints" I mean "intervals" ;)

	// for _, bt := range bytes {
	// 	win1, win2 := nextWin(bt)
	// 	encodedInts = append(encodedInts, processWin(win1))
	// 	encodedInts = append(encodedInts, processWin(win2))
	// }

	for i, bt := range bytes {
		bits := getBits(bt)
		for _, bit := range bits {
			// Do you update probabilities before or after encoding step?
			// Probably doesnt matter as long as you do it the same order in decoder
			updateWin(bit)
			updateCount(root, window)
			updateProb(root, bit)
		}
		if i%llength == 0 {
			cnt := 5 * i / llength
			fmt.Println(cnt, "%")
		}
	}
	fmt.Println(100, "%")

	// NOTE (8/20/22): Arithmetic encoding should return a SINGLE number representing the final probability value

	//os.WriteFile(op,root.p (converted to byte array),os.ModeDevice)
	fmt.Println("PROB: ", root.p)

	recCheck(root, []int{})

	return

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
