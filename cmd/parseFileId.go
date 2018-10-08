package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const fileids = "./data/dirac/bzdata/bzbackup/bzfileids.dat"

func main() {
	extractField(fileids, "compare-fileids-unsorted-go.dat", 1, 2)
}

func extractField(infilename, outfilename string, fieldNo, expectedFields int) {
	fmt.Fprintf(os.Stderr, "-= Parsing %s\n", fileids)
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
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Println(line)
		fields := strings.Split(line, "\t")
		if len(fields) == expectedFields {
			// fmt.Printf("%s\n", fields[fieldNo])
			fmt.Fprintf(outfile, "%s\n", fields[fieldNo])
			count++
		} else {
			fmt.Fprintf(os.Stderr, "Err: %d!=%d %v\n", len(fields), expectedFields, fields)
		}

	}
	fmt.Fprintf(os.Stderr, "-= Parsed %d lines\n", count)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}
