package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	enabledEncodings = []encoding{}
	bufferSize       int
	inputFile        string
	startByte        int
	lengthBytes      int
	enableColors     bool
	colorWidth       int
	enableOffsets    bool
	enablePosition   bool
)

func init() {
	for i, e := range encodings {
		flag.BoolVar(&encodings[i].Enabled, e.Name, false, e.Desc)
	}
	flag.StringVar(&inputFile, "file", "", "The file to read input from (stdin by default)")
	flag.IntVar(&startByte, "start", 0, "The byte to start reading at")
	flag.IntVar(&lengthBytes, "length", 0, "The number of bytes to read")
	flag.BoolVar(&enablePosition, "pos", false, "Show the position in bytes within the file")
	flag.BoolVar(&enableColors, "color", false, "Decorate output with rainbow colors")
	flag.IntVar(&colorWidth, "colorWidth", 2, "Width in bytes of each color")
	flag.BoolVar(&enableOffsets, "offsets", false, "Show multi-byte values for every offset")
	flag.IntVar(&bufferSize, "width", 8, "How many bytes to print per line")
	flag.Parse()

	for _, e := range encodings {
		if e.Enabled {
			enabledEncodings = append(enabledEncodings, e)
		}
	}
	if len(enabledEncodings) == 0 {
		for i, e := range encodings {
			if e.Name == "hex" || e.Name == "utf8" || e.Name == "u8" {
				encodings[i].Enabled = true
				enabledEncodings = append(enabledEncodings, e)
			}
		}
	}

}
func printHeader(enc []encoding) {
	sepWidth := 0
	if enablePosition {
		str := "position  "
		fmt.Fprint(os.Stdout, str)
		sepWidth += len(str)
	}
	for _, e := range enc {
		stri := fmt.Sprintf("%-*s  ", e.EncodingWidth(bufferSize), e.Name)
		sepWidth += len(stri)
		fmt.Fprint(os.Stdout, stri)
	}
	fmt.Fprint(os.Stdout, "\n")
	var sep string
	for i := 0; i < sepWidth; i++ {
		sep += "-"
	}
	fmt.Fprintln(os.Stdout, sep)
}
func processLine(chunk []byte, position int) {

	var encoded string
	for i := 0; i < len(enabledEncodings); i++ {
		encoded += enabledEncodings[i].Encode(chunk)
		if len(encoded) > 0 {
			encoded += "  "
		}
	}
	if len(encoded) > 0 {
		var ln string
		if enablePosition {
			ln += fmt.Sprintf("%8d  ", position - bufferSize)
		}
		ln += encoded
		if (enableColors) {
			ln += "\x1b[0m"
		}
		fmt.Fprintln(os.Stdout, ln)
	}

}
func main() {
	if bufferSize <= 0 {
		fmt.Fprintln(os.Stderr, "width must be >0")
		return
	}

	// Create a buffered reader
	reader := bufio.NewReader(os.Stdin);
	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Println("Error opening ", inputFile, ":", err)
			os.Exit(1)
		}
		reader = bufio.NewReader(file)
		defer file.Close()
	}

	// read full buffer

	printHeader(enabledEncodings)
	_, _ = io.CopyN(io.Discard, reader, int64(startByte))
	position := startByte

ReadLoop:
	for {
		chunk := make([]byte, bufferSize)
		n, err := io.ReadFull(reader, chunk)

		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				processLine(chunk[:n], position)
				processLine([]byte{}, position)
				break ReadLoop
			}
			fmt.Fprintln(os.Stderr, "error reading input:", err)
			return
		}

		// Only process the bytes that were actually read
		processLine(chunk[:n], position)

		position += n
		if lengthBytes > 0 && position > startByte + lengthBytes {
			break;
		}
	}
}
