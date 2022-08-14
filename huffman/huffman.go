package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
)

var filepath = "enwik8"

// var filepath = "filey"
var outPath = "enwik8_encoded"

var freq = map[byte]int{}

// var dict = map[byte]string{}
var dict = map[byte][]int{}

type treeNode struct {
	value byte
	left  *treeNode
	right *treeNode
}

type huffmanNode struct {
	frequency int
	value     byte
	left      *huffmanNode
	right     *huffmanNode
	isLeaf    bool
}

func minTree(sorted []byte) *huffmanNode {
	root := huffmanNode{value: sorted[0], frequency: freq[sorted[0]], isLeaf: true}
	//ret := &root
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

func huffmanTree(tree *huffmanNode) *huffmanNode {

	//get mintree
	//get two smallest nodes, create parent node for them
	//continue until there is only one parent, return as root

	//--------
	queue := []*huffmanNode{tree}
	children := []*huffmanNode{}
	root := &huffmanNode{}
	for len(queue) > 0 {
		root, queue = pop(queue)
		if root.isLeaf {
			rooter := *root //delete
			if root.left != nil {
				queue = append(queue, rooter.left) //change rooter to root
			}
			if root.right != nil {
				queue = append(queue, rooter.right) //change rooter to root
			}
			trimKids := huffmanNode{value: rooter.value, frequency: rooter.frequency, isLeaf: true} //delete
			children = append(children, &trimKids)                                                  //delete
		} else { //delete
			children = append(children, root) //should be out of else statement
		} //delete

		if len(children) == 2 {
			parent := huffmanNode{frequency: children[0].frequency + children[1].frequency, isLeaf: false, left: children[0], right: children[1]}
			queue = append(queue, &parent)
			children = []*huffmanNode{}
		}
	}
	// if len(children) == 1 {
	// 	parent := huffmanNode{frequency: children[0].frequency + root.frequency, isLeaf: false, left: children[0], right: root}
	// 	return &parent
	// }
	return root
}

func createDictOLD(node *huffmanNode, ret []int) {
	if node == nil {
		return
	}
	if node.isLeaf {
		dict[node.value] = ret
		//fmt.Println(node.value, string(node.value), ret, node.frequency)
	} else {
		//createDict(node.left, ret+"0")
		//createDict(node.right, ret+"1")

		rl := make([]int, len(ret)+1)
		copy(rl, ret)
		rl = append(rl, 0)
		createDict(node.left, rl)
		rr := make([]int, len(ret)+1)
		copy(rr, ret)
		rr = append(rr, 1)
		createDict(node.right, rr)
	}
	return
}

func createDict(node *huffmanNode, ret []int) {
	if node.isLeaf {
		dict[node.value] = ret
		return
		//fmt.Println(node.value, string(node.value), ret, node.frequency)
	}
	//createDict(node.left, ret+"0")
	//createDict(node.right, ret+"1")

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
		//fmt.Println(b, "mod", int(math.Pow(2, ex)), " = ", temp)
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
		if i%1000000 == 0 {
			fmt.Println(i)
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
	// fmt.Println("---DICT---")
	// for i := range dict {
	// 	fmt.Println(string(i), dict[i])
	// }
	// fmt.Println("--------")
	// fmt.Println()

	// fmt.Println()
	// fmt.Println(intData)
	// fmt.Println()
	// //fmt.Println("Bits: ", len(intData))
	// fmt.Println()

	//Check that data read in from output file and converted with readBits() == intData
	// //fmt.Println("data ", data)
	// data, _ = os.ReadFile(outPath)
	// for _, x := range data {
	// 	fmt.Print(readBits(int(x)))
	// }
	// fmt.Println()
	// os.Exit(0)

	//Check leaf nodes of huffman tree
	// fmt.Println("Checking tree: ")
	// recCheck(hufT, []int{})
	// fmt.Println("  ")

	return hufT
}

func decode(outPath string, hufT *huffmanNode) (string, int) {
	fmt.Println("Reading encoded data...")
	fmt.Println()
	bytes, err := os.ReadFile(outPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Bytes: ", len(bytes))
	fmt.Println("Bits: ", len(bytes)*8)
	//intData := make([]int, len(bytes)*8)
	intData := []int{}
	for _, x := range bytes {
		intData = append(intData, readBits(int(x))...)
	}
	fmt.Println(intData[:30])
	//intData = intData[614 : len(intData)-2] //618
	//intData = intData[88 : len(intData)-2]
	//testList := []int{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	//intData = append(testList, intData...)

	// fmt.Println()
	// fmt.Println("intData decoder")
	// fmt.Println(intData)
	// fmt.Println()
	// fmt.Println(len(intData))
	//os.Exit(1)

	//test := []int{}
	//stringTest := []string{}

	data := []byte{}
	root := hufT
	for _, i := range intData {
		if root.isLeaf == true {
			data = append(data, root.value)
			//fmt.Println(root.isLeaf)
			//stringTest = append(stringTest, string(root.value))
			root = hufT
			//fmt.Println(test)
			//stringTestHelper := *root

			// for y, z := range dict {
			// 	test2 := ""
			// 	for _, q := range z {
			// 		test2 += string(rune(q))
			// 	}
			// 	if test == test2 {
			// 		fmt.Println(y, z, test)
			// 	}
			// }
			//test = []int{}
		}
		//test = append(test, i)
		if i == 0 {
			root = root.left
		} else if i == 1 {
			root = root.right
		}

	}
	//fmt.Println(stringTest)
	//fmt.Println(data)
	//fmt.Println(len(data))
	return string(data), len(bytes)
}

func testMin(minT *treeNode) {
	rt := minT
	for {
		if rt.left == nil {
			break
		}
		fmt.Println(rt, rt.left, rt.right)
		rt = rt.left
	}
	fmt.Println(" --- ")
	rt = minT
	for {
		if rt.right == nil {
			break
		}
		fmt.Println(rt, rt.left, rt.right)
		rt = rt.right
	}
	os.Exit(1)
}

func testHuf(hufT *huffmanNode) {
	rt := hufT
	for {
		if rt.isLeaf {
			fmt.Println("LEAF ", rt)
			break
		}
		fmt.Println(rt, rt.left, rt.right)
		rt = rt.left
	}
	fmt.Println(" --- ")
	rt = hufT
	for {
		if rt.isLeaf {
			fmt.Println("LEAF ", rt)
			break
		}
		fmt.Println(rt, rt.left, rt.right)
		rt = rt.right
	}
	os.Exit(1)
}

func main() {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	//bytes = bytes[:1000000]
	fmt.Println("Bytes: ", len(bytes))
	fmt.Println("Bits: ", len(bytes)*8)

	fmt.Println(" ")
	//fmt.Println("Input: ", string(bytes))
	fmt.Println(" ")
	//fmt.Println(bytes)
	fmt.Println(" ")
	fmt.Println(" ------- ")

	hufT := encode(bytes, outPath)
	_, deced := decode(outPath, hufT)
	fmt.Println(" ")
	fmt.Println("New file is ", 100*deced/len(bytes), "% of the size of origional file")
}
