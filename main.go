package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// formatWithCommas formats a number with comma separators
func formatWithCommas(num float64, precision int) string {
	str := fmt.Sprintf("%."+strconv.Itoa(precision)+"f", num)
	parts := strings.Split(str, ".")

	// Add commas to the integer part
	intPart := parts[0]
	if len(intPart) > 3 {
		var result strings.Builder
		for i, digit := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		intPart = result.String()
	}

	if len(parts) > 1 {
		return intPart + "." + parts[1]
	}
	return intPart
}

// formatIntWithCommas formats an integer with comma separators
func formatIntWithCommas(num int) string {
	str := strconv.Itoa(num)
	if len(str) > 3 {
		var result strings.Builder
		for i, digit := range str {
			if i > 0 && (len(str)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		return result.String()
	}
	return str
}

func usage() {
	me := filepath.Base(os.Args[0])
	msg := fmt.Sprintf(`usage: %s [ FILE ]

Displays the entropy (in total bits and bits/byte) of the data in FILE
(or from stdin if FILE is omitted).
`, me)
	fmt.Fprint(os.Stderr, msg)
	os.Exit(1)
}

func main() {
	me := filepath.Base(os.Args[0])

	// Parse command line and create a file object to read the data.
	if len(os.Args) > 2 {
		usage()
	}

	var input io.Reader

	if len(os.Args) > 1 {
		if os.Args[1][0] == '-' {
			usage()
		}

		// Check if file exists
		if _, err := os.Stat(os.Args[1]); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s: File not found: '%s'\n", me, os.Args[1])
			os.Exit(2)
		}

		file, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: Error opening file: %v\n", me, err)
			os.Exit(2)
		}
		defer file.Close()
		input = file
	} else {
		// Read from stdin
		input = os.Stdin
	}

	// Create the frequency distribution of each byte value from 0 to 255.
	freqCounters := make([]int, 256)
	readSize := 100 * 1024
	byteCount := 0

	buffer := make([]byte, readSize)
	reader := bufio.NewReader(input)

	// Read first chunk to check if there's any data
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "%s: Error reading data: %v\n", me, err)
		os.Exit(1)
	}

	if n == 0 {
		fmt.Printf("%s: No data found!\n", me)
		os.Exit(1)
	}

	// Process the first chunk
	for i := 0; i < n; i++ {
		freqCounters[buffer[i]]++
		byteCount++
	}

	// Continue reading remaining data
	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%s: Error reading data: %v\n", me, err)
			os.Exit(1)
		}
		if n == 0 {
			break
		}

		for i := 0; i < n; i++ {
			freqCounters[buffer[i]]++
			byteCount++
		}
	}

	// Compute the entropy.

	// Compute the probability of each byte value.
	probabilities := make([]float64, 256)
	for v := 0; v < 256; v++ {
		probabilities[v] = float64(freqCounters[v]) / float64(byteCount)
	}

	// Compute the base-2 log of each probability.  We need to handle zero probabilities
	// to avoid computing math.Log(0) which would result in -Inf.
	log2probabilities := make([]float64, 256)
	for v := 0; v < 256; v++ {
		if probabilities[v] > 0 {
			log2probabilities[v] = math.Log2(probabilities[v])
		} else {
			log2probabilities[v] = 0
		}
	}

	// Compute the entropy in bits per byte.  This is Claude Shannon's entropy function.
	// See https://en.wikipedia.org/wiki/Entropy_(information_theory)#Definition
	entropy := 0.0
	for v := 0; v < 256; v++ {
		entropy -= probabilities[v] * log2probabilities[v]
	}

	// Output the results in the same format as the Python script
	fmt.Printf("%s bits (%s bytes) = %.4f%% of %s bytes (%.4f bits/byte)\n",
		formatWithCommas(entropy*float64(byteCount), 2),
		formatWithCommas(entropy*float64(byteCount)/8, 2),
		entropy/8*100,
		formatIntWithCommas(byteCount),
		entropy)
}
