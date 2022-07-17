package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/spf13/pflag"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const bufferSize = 4096
const createFilePerm = 0700

var ErrKeyTooShort = errors.New("key is too short")
var ErrNoKeyFile = errors.New("key file path not provided")
var ErrNoOutputFile = errors.New("output file path not provided")
var ErrMissingPlaceholder = errors.New("missing {i} placeholder in file path")

// Command-line Flags
var verbose bool
var generate, keyPath, inputPath, outputPath string
var genCount, genOffset int

func init() {
	// Generic
	pflag.StringVarP(&outputPath, "output", "o", "", "the file to write the output to")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	// Key File Generation
	pflag.StringVarP(&generate, "generate", "g", "", "the size of the key file to generate")
	pflag.IntVarP(&genCount, "count", "c", 1, "the number of key files to generate")
	pflag.IntVarP(&genOffset, "offset", "O", 0, "the index offset (start) for the file index")
	// Encoding and Decoding
	pflag.StringVarP(&keyPath, "key", "k", "", "the file containing the key")
	pflag.StringVarP(&inputPath, "input", "i", "", "the file containing the input, else stdin")
	pflag.Parse()
}

func main() {
	err := program()
	if err != nil {
		log.Fatal(err)
	}
}

func program() error {
	vprintln("step: program")
	if generate != "" {
		return generateKeys()
	}
	return xorFiles(keyPath, inputPath, outputPath)
}

func generateKeys() error {
	vprintln("step: generateKeys")
	if outputPath == "" {
		return ErrNoOutputFile
	}

	genSize, err := parseSizeBytes(generate)
	if err != nil {
		return err
	}
	if genCount == 1 {
		return generateKey(outputPath, genSize)
	}
	if !strings.Contains(outputPath, "{i}") {
		return ErrMissingPlaceholder
	}
	for i := genOffset; i < genCount+genOffset; i++ {
		path := strings.ReplaceAll(outputPath, "{i}", strconv.Itoa(i))
		err := generateKey(path, genSize)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateKey(filePath string, genSize int64) error {
	vprintln("step: generateKey")
	file, err := os.OpenFile(filePath, os.O_CREATE, createFilePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	vprintln("generating key of size", genSize, "bytes and writing it to", filePath)
	_, err = io.CopyN(file, rand.Reader, genSize)
	return err
}

func xorFiles(keyFilePath, inputFilePath, outputFilePath string) error {
	vprintln("step: xorFiles")
	if keyFilePath == "" {
		return ErrNoKeyFile
	}
	vprintln("opening key file at", keyFilePath)
	keyFile, err := os.Open(keyFilePath)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	inputReader := os.Stdin
	if inputFilePath != "" {
		vprintln("opening input file at", inputFilePath)
		inputFile, err := os.Open(inputFilePath)
		if err != nil {
			return err
		}
		defer inputFile.Close()
		inputReader = inputFile
	}

	outputWriter := os.Stdout
	if outputFilePath != "" {
		vprintln("opening input file at", outputFilePath)
		outputFile, err := os.OpenFile(outputFilePath, os.O_CREATE, createFilePerm)
		if err != nil {
			return err
		}
		defer outputFile.Close()
		outputWriter = outputFile
	}

	_, err = xorStreams(keyFile, inputReader, outputWriter)
	return err
}

func xorStreams(key, input io.Reader, output io.Writer) (int, error) {
	vprintln("step: xorStreams")
	written := 0
	msgBuf := make([]byte, bufferSize)
	keyBuf := make([]byte, bufferSize)
	for {
		n, msgErr := input.Read(msgBuf)
		if msgErr != nil && !errors.Is(msgErr, io.EOF) {
			return written, msgErr
		}
		_, keyErr := io.ReadAtLeast(key, keyBuf, n)
		if keyErr != nil {
			if errors.Is(keyErr, io.EOF) || errors.Is(keyErr, io.ErrUnexpectedEOF) {
				return written, ErrKeyTooShort
			}
			return written, keyErr
		}
		vprintln(fmt.Sprintf("rdbuf bytes=%d", n))

		for i := 0; i < n; i++ {
			msgBuf[i] ^= keyBuf[i]
		}

		w, err := output.Write(msgBuf[:n])
		if err != nil {
			return written, err
		}
		written += w

		if n < len(msgBuf) {
			return written, nil
		}
	}
}

func vprintln(a ...any) {
	if verbose {
		fmt.Println(a...)
	}
}

const (
	// 1000-based
	kilobyte = 1000
	megabyte = 1000 * kilobyte
	gigabyte = 1000 * megabyte
	terabyte = 1000 * gigabyte
	// 1024-based
	kibibyte = 1024
	mebibyte = 1024 * kibibyte
	gibibyte = 1024 * mebibyte
	tebibyte = 1024 * gibibyte
)

var ErrInvalidUnit = errors.New("invalid byte size unit")

func parseSizeBytes(str string) (int64, error) {
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ToLower(str)

	index := -1
	for i, char := range str {
		if !unicode.IsNumber(char) {
			index = i
			break
		}
	}
	if index == -1 {
		return strconv.ParseInt(str, 10, 64)
	}

	size := str[index] // B, K, M, G, T
	unitGap := false   // 1000 vs 1024
	if index+1 < len(str) {
		// 'i' in 'KiB'
		unitGap = str[index+1] == 'i'
	}

	scale := int64(1)
	if unitGap {
		switch size {
		case 't':
			scale = tebibyte
		case 'g':
			scale = gibibyte
		case 'm':
			scale = mebibyte
		case 'k':
			scale = kibibyte
		case 'b':
		default:
			return 0, ErrInvalidUnit
		}
	} else {
		switch size {
		case 't':
			scale = terabyte
		case 'g':
			scale = gigabyte
		case 'm':
			scale = megabyte
		case 'k':
			scale = kilobyte
		case 'b':
		default:
			return 0, ErrInvalidUnit
		}
	}

	value, err := strconv.ParseInt(str[:index], 10, 64)
	if err != nil {
		return 0, err
	}

	return value * scale, nil
}
