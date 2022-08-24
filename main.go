package main

import (
	backup "compression/backup"
	ctw "compression/ctw"
	Huffman "compression/huffman"
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
		//fmt.Println("DECODING NOW")
		//ctw.Decode(outPath, outOutPath)

	}
	if alg == "backup" {
		backup.Encode(filepath, outPath)
	}
}
