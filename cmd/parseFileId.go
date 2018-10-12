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
const filelists0 = "./data/dirac/bzdata/bzfilelists/v0009a98724006e621c1646e011f_root_filelist.dat"
const filelists1 = "./data/dirac/bzdata/bzfilelists/v0019ae8724006e621c1646e011f______filelist.dat"

func main() {
	var lines []string
	lines = extractField(fileids, 1, filterFileIds)
	lines = sortAndUniq(lines)
	write("compare-fileids-sorted.dat", lines)

	lines0 := extractField(filelists0, 3, filterFilelist)
	lines1 := extractField(filelists1, 3, filterFilelist)
	lines = append(lines0, lines1...)
	lines = sortAndUniq(lines)
	write("compare-filelists-sorted.dat", lines)
}

func filterFileIds(fields []string) bool {
	const expectedFields = 2
	if len(fields) != expectedFields {
		fmt.Fprintf(os.Stderr, "Err: %d!=%d %v\n", len(fields), expectedFields, fields)
		return false
	}
	return true
}
func filterFilelist(fields []string) bool {
	if strings.HasPrefix(fields[0], "#") {
		return false
	}
	const expectedFields = 4
	if len(fields) != expectedFields {
		fmt.Fprintf(os.Stderr, "Err: %d!=%d %v\n", len(fields), expectedFields, fields)
		return false
	}
	// only take 'f':files, not 's':symbolic links
	if !strings.HasPrefix(fields[0], "f") {
		// check for other than f,s
		if "s" != fields[0] {
			fmt.Fprintf(os.Stderr, "Unexpected: filetype (%s) not in [f,s]\n", fields[0])
		}
		return false
	}
	return true
}

func extractField(infilename string, fieldNo int, filter func(fields []string) bool) []string {
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
		fields := strings.Split(line, "\t")
		if !filter(fields) {
			skipped++
			continue
		}
		if len(fields) > fieldNo {
			lines = append(lines, fields[fieldNo])
		} else {
			// fmt.Fprintf(os.Stderr, "Err: %d <= %d %v\n", len(fields), fieldNo)

		}
	}
	fmt.Fprintf(os.Stderr, "-= Parsed %d lines (%d skipped)\n", len(lines), skipped)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
}

func sortAndUniq(lines []string) []string {
	sort.Strings(lines)
	fmt.Fprintf(os.Stderr, "-= Sorted %d lines\n", len(lines))

	uniqed := make([]string, 0, 1000)
	deduped := 0
	var prev string
	for _, line := range lines {
		if prev == line {
			// fmt.Fprintf(os.Stderr, "Dedup'd line(%d) %s\n", i, line)
			deduped++
			continue
		}
		uniqed = append(uniqed, line)
		prev = line
	}
	fmt.Fprintf(os.Stderr, "-= Uniq'd %d lines, dedup'd %d (total=%d)\n", len(uniqed), deduped, len(lines))
	return uniqed
}
func write(outfilename string, lines []string) {
	fmt.Fprintf(os.Stderr, "-= Writing %s\n", outfilename)
	outfile, err := os.Create(outfilename)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	bw := bufio.NewWriter(outfile)
	fmt.Fprintf(os.Stderr, "-= Writing %d lines\n", len(lines))
	for i, field := range lines {
		if len(field) == 0 {
			fmt.Fprintf(os.Stderr, "Empty field on line:  %d\n", i)
			continue
		}
		// fmt.Fprintf(outfile, "%s\n", field)
		fmt.Fprintf(bw, "%s\n", field)
		// n4, err := w.WriteString(fmt.Sprintf("%s\n"))
		// bw.WriteString(fmt.Sprintf("%s\n", field))
	}
	err = bw.Flush()
	if err != nil {
		log.Fatal(err)
	}
}

func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
