package ctw

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

var D = 3
var window = "" // replace with bits when you import math.bits package. Length(window) = D
// Use bitshifting (<< or >> in python, might need math.bits import in golang) to keep track of window storing previous bits.

type node struct {
	code  string
	left  *node   //adds 1 to code
	right *node   //adds 0 to code
	c0    int     //count of 0s
	c1    int     //count of 1s
	p     float64 //weighted probability, initialize at 0
}

//right is the child whos code is this code + 0, and left has code that is this code + 1
// c0 must be >= the sum of the c0 values of the node's children (and same with c1)

// Recursively update probabilities of all nodes by calling this func on the root node.
func updateProb(n *node, update int) {
	newP := 0.0
	if update == 0 {
		newP = n.p * (float64(n.c0) + 0.5) / (float64(n.c0) + float64(n.c1) + 1.0)
	} else if update == 1 {
		newP = n.p * (float64(n.c1) + 0.5) / (float64(n.c0) + float64(n.c1) + 1.0)
	}

	if len(n.code) == D {
		n.p = newP
	} else {
		updateProb(n.left, update)
		updateProb(n.right, update)
		n.p = 0.5*newP + 0.5*n.left.p*n.right.p
	}
}

// Update prediction data for suffix nodes
func updateCount(n *node, update int) {
	if len(n.code) == D {
		if n.code == window {
			if update == 0 {
				n.c0++
			} else if update == 1 {
				n.c1++
			}
		}
	} else {
		updateCount(n.left, update)
		updateCount(n.right, update)
		if update == 0 {
			n.c0 = n.left.c0 + n.right.c0
		} else if update == 1 {
			n.c1 = n.left.c1 + n.right.c1
		}
	}
}

// For updateCount() might be more efficient to specifically search for the node that matches the current window by navigating the tree, instead of calling recursion on all nodes
// just need to change the else statement to only call on the child that corresponds to the next bit in the window. pass window recursively, removing 1 bit each time.

func Encode(fp string, op string) {

}
