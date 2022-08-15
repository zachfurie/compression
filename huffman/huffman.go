package Huffman

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
)

var filepath = "enwik8"
var outPath = "enwik8_encoded"

var freq = map[byte]int{}
var dict = map[byte][]int{}

type huffmanNode struct {
	frequency int
	value     byte
	left      *huffmanNode
	right     *huffmanNode
	isLeaf    bool
}

func minTree(sorted []byte) *huffmanNode {
	root := huffmanNode{value: sorted[0], frequency: freq[sorted[0]], isLeaf: true}
	queue := []*huffmanNode{&root}
	i := 1

	for len(queue) != 0 {
		pop := queue[0]
		queue = queue[1:]
		if i == len(sorted) {
			break
		}
		left := huffmanNode{value: sorted[i], frequency: freq[sorted[i]], isLeaf: true}
		pop.left = &left
		queue = append(queue, &left)
		i++
		if i == len(sorted) {
			break
		}
		right := huffmanNode{value: sorted[i], frequency: freq[sorted[i]], isLeaf: true}
		pop.right = &right
		queue = append(queue, &right)
		i++
	}
	return &root
}

func queueSort(queue []*huffmanNode) []*huffmanNode {
	sort.Slice(queue, func(i, j int) bool {
		return queue[i].frequency < queue[j].frequency
	})
	return queue
}

func pop(queue []*huffmanNode) (*huffmanNode, []*huffmanNode) {
	return queue[0], queue[1:]
}

// get mintree,
// get two smallest nodes, create parent node for them
// continue until there is only one parent, return as root
func huffmanTree(tree *huffmanNode) *huffmanNode {
	queue := []*huffmanNode{tree}
	children := []*huffmanNode{}
	root := &huffmanNode{}
	for len(queue) > 0 {
		root, queue = pop(queue)
		if root.isLeaf {
			rooter := *root
			if root.left != nil {
				queue = append(queue, rooter.left)
			}
			if root.right != nil {
				queue = append(queue, rooter.right)
			}
			trimKids := huffmanNode{value: rooter.value, frequency: rooter.frequency, isLeaf: true}
			children = append(children, &trimKids)
		} else {
			children = append(children, root)
		}

		if len(children) == 2 {
			parent := huffmanNode{frequency: children[0].frequency + children[1].frequency, isLeaf: false, left: children[0], right: children[1]}
			queue = append(queue, &parent)
			children = []*huffmanNode{}
		}
	}
	return root
}

func createDict(node *huffmanNode, ret []int) {
	if node.isLeaf {
		dict[node.value] = ret
		return
		//fmt.Println(node.value, string(node.value), ret, node.frequency)
	}
	rl := make([]int, len(ret))
	copy(rl, ret)
	rl = append(rl, 0)
	createDict(node.left, rl)
	rr := make([]int, len(ret))
	copy(rr, ret)
	rr = append(rr, 1)
	createDict(node.right, rr)
	return
}

// Convert byte to int
func writeBits(ints []int) int {
	i := len(ints)
	ex := float64(-1)
	ret := 0
	for i > 0 {
		i--
		ex++
		ret += ints[i] * int(math.Pow(2, ex))
	}
	return ret
}

// Convert int to byte
func readBits(b int) []int {
	ret := []int{0, 0, 0, 0, 0, 0, 0, 0}
	if b >= 256 {
		fmt.Println("ERROR, tried to convert integer >=256 to byte")
		return ret
	}
	i := 8
	ex := float64(0)
	for i >= 1 {
		i--
		ex++
		temp := b % int(math.Pow(2, ex))
		if temp == b {
			ret[i] = 1
			break
		} else if temp != 0 {
			ret[i] = 1
			b = b - temp
		}
	}
	return ret
}

func encode(bytes []byte, outPath string) *huffmanNode {
	for _, x := range bytes {
		freq[x] += 1
	}
	sorted := make([]byte, len(freq))
	y := 0
	for i := range freq {
		sorted[y] = i
		y++
	}
	sort.Slice(sorted, func(i, j int) bool {
		return freq[sorted[i]] < freq[sorted[j]]
	})
	minT := minTree(sorted)
	hufT := huffmanTree(minT)
	fmt.Println("creating dict...")
	createDict(hufT, []int{})
	fmt.Println("encoding...")
	intData := []int{}
	for i, x := range bytes {
		intData = append(intData, dict[x]...)
		if i%(len(bytes)/10) == 0 {
			fmt.Println(i, "/", len(bytes))
		}
	}
	fmt.Println("converting to bytes...")
	data := []byte{}
	i := 0
	for i < len(intData) {
		temp := writeBits(intData[i : i+8])
		data = append(data, byte(temp))
		i += 8
	}
	fmt.Println("writing to file...")
	os.WriteFile(outPath, data, os.ModeDevice)
	fmt.Println("done")
	fmt.Println()

	return hufT
}

func decode(outPath string, hufT *huffmanNode) (string, int) {
	fmt.Println("Reading encoded data...")
	fmt.Println()
	bytes, err := os.ReadFile(outPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Compressed Bytes: ", len(bytes))
	fmt.Println("Compressed Bits: ", len(bytes)*8)
	//intData := make([]int, 0, len(bytes)*8) //NOTE: this should work to preallocate the required memory while still intilializing a slice of length 0
	//readBits() returns a slice of 8 ints, so it is easier to initialize intData without specifying length and just keep appending the list to intData. However, if you want to initialize intData with the correct length, you would have to assign each int returned by readBits to the correct index in intData
	intData := []int{}
	for _, x := range bytes {
		intData = append(intData, readBits(int(x))...)
	}
	data := []byte{}
	root := hufT
	for _, i := range intData {
		if root.isLeaf == true {
			data = append(data, root.value)
			root = hufT
		}
		if i == 0 {
			root = root.left
		} else if i == 1 {
			root = root.right
		}

	}
	return string(data), len(bytes)
}

func HuffMain(fp string, op string) {
	filepath = fp
	outPath = op
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Bytes: ", len(bytes))
	fmt.Println("Bits: ", len(bytes)*8)
	fmt.Println(" ")
	fmt.Println(" ------- ")
	fmt.Println(" ")

	hufT := encode(bytes, outPath)
	_, deced := decode(outPath, hufT)
	fmt.Println(" ")
	fmt.Println("New file is ", 100*deced/len(bytes), "% of the size of origional file")
}

//------------------------------------------
//Bug Tests

// Get path of leafnodes in huffman tree
func recCheck(hufT *huffmanNode, list []int) {
	if hufT.isLeaf {
		fmt.Println(*hufT, list, string(hufT.value))
		return
	}
	fmt.Println(*hufT, list)
	l := make([]int, len(list))
	copy(l, list)
	l = append(l, 0)
	r := make([]int, len(list))
	copy(r, list)
	r = append(r, 1)
	recCheck(hufT.left, l)
	recCheck(hufT.right, r)
}
