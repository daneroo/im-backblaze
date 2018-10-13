package main

// Attempts to answer the question:
// - Which files are NOT backed up and why ?

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const baseDir = "./data/dirac/bzdata"

// const baseDir = "./data/fermat/bzdata"
// const baseDir = "/Library/Backblaze.bzpkg/bzdata"

// bzlogs/bzreports_lastfilestransmitted/13.log

func main() {
	parseTransmited()
	// xfrs := parseTransmited()
	// write("compare-fileids-sorted.dat", fileIds)
}

func parseTransmited() []string {

	files, err := filepath.Glob(baseDir + "/bzlogs/bzreports_lastfilestransmitted/*.log")
	if err != nil {
		log.Fatal(err)
	}

	lines := make([]string, 0)
	for _, file := range files {
		morelines := extractLastField(file)
		lines = append(lines, morelines...)
	}
	// lines = sortAndUniq(lines)
	return lines
}

func extractLastField(infilename string) []string {
	fmt.Fprintf(os.Stderr, "-= Parsing %s\n", infilename)
	infile, err := os.Open(infilename)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	scanner := bufio.NewScanner(infile)
	lines := make([]string, 0, 1000)
	skipped := 0
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " - ")

		stamp := fields[0]
		speed := ""
		size := ""
		if len(fields) == 6 {
			speed = fields[3]
			size = fields[4]
		}
		fname := fields[len(fields)-1]

		if strings.HasPrefix(fname, "Multiple small files batched in one request") {
			skipped++
			continue
		}

		// fmt.Fprintf(os.Stderr, "%02d %s|%20s|%20s|%s\n", len(fields), stamp, speed, size, fname)
		fmt.Printf("%02d %s|%20s|%20s|%s\n", len(fields), stamp, speed, size, fname)

		lines = append(lines, fname)
	}
	fmt.Fprintf(os.Stderr, "-= Parsed %d lines (%d skipped)\n", len(lines), skipped)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
}
