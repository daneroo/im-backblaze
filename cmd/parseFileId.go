package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

const fileids = "./data/dirac/bzdata/bzbackup/bzfileids.dat"

func main() {
	extractField(fileids, "compare-fileids-sorted-go.dat", 1, 2)
}

func extractField(infilename, outfilename string, fieldNo, expectedFields int) {
	fmt.Fprintf(os.Stderr, "-= Parsing %s\n", infilename)
	fmt.Fprintf(os.Stderr, "-= Sorting %s\n", outfilename)
	infile, err := os.Open(infilename)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	outfile, err := os.Create(outfilename)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	scanner := bufio.NewScanner(infile)
	lines := make([]string, 1000)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Println(line)
		fields := strings.Split(line, "\t")
		if len(fields) == expectedFields {
			// fmt.Printf("%s\n", fields[fieldNo])
			// fmt.Fprintf(outfile, "%s\n", fields[fieldNo])
			lines = append(lines, fields[fieldNo])
			count++
		} else {
			fmt.Fprintf(os.Stderr, "Err: %d!=%d %v\n", len(fields), expectedFields, fields)
		}
	}
	fmt.Fprintf(os.Stderr, "-= Parsed %d lines\n", count)
	fmt.Fprintf(os.Stderr, "-= line #12345 %s\n", lines[12345])

	sort.Strings(lines)
	fmt.Fprintf(os.Stderr, "-= Sorted %d lines\n", len(lines))
	fmt.Fprintf(os.Stderr, "-= line #12345 %s\n", lines[12345])

	// w := bufio.NewWriter(outfile)
	// n4, err := w.WriteString("buffered\n")
	// fmt.Printf("wrote %d bytes\n", n4)
	fmt.Fprintf(os.Stderr, "-= Writing %d lines\n", count)
	for i, field := range lines {
		fmt.Fprintf(outfile, "%s\n", field)
		if i == 12345 {
			fmt.Fprintf(os.Stderr, "-= line #%d %s\n", i, field)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}
