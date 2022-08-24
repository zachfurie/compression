package main

import (
	ctw "compression/ctw"
	Huffman "compression/huffman"
	"fmt"
)

var filepath = "enwik8"
var outPath = "enwik8_encoded"
var outOutPath = "enwik8_decoded"

// var alg = "huffman"
var alg = "ctw"

func main() {
	if alg == "huffman" {
		Huffman.HuffMain(filepath, outPath)
	}
	if alg == "ctw" {
		ctw.Encode(filepath, outPath)
		fmt.Println("DECODING NOW")
		ctw.Decode(outPath, outOutPath)
	}
}
