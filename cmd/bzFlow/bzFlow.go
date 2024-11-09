package main

// Attempts to answer the question:
// - Which files are NOT backed up and why ?

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/daneroo/backblaze"
)

var hosts = []string{"galois", "davinci"}

// bzlogs/bzreports_lastfilestransmitted/13.log
const doSummary = false

// Should this be UTC? I think the filenames are localtime
const daysAgo = 20

var minStamp = time.Now().AddDate(0, 0, -daysAgo).Format("2006-01-02")

const maxStamp = "2040-12-31"

func main() {
	for _, host := range hosts {
		fmt.Fprintf(os.Stderr, "Processing host: %s\n", host)
		// bzdata on localhost!
		// baseDir = "/Library/Backblaze.bzpkg/bzdata"
		baseDir := fmt.Sprintf("./data/%s/bzdata", host)

		files, err := filepath.Glob(baseDir + "/bzlogs/bzreports_lastfilestransmitted/*.log")
		if err != nil {
			log.Printf("Error for host %s: %v", host, err)
			continue
		}

		allxfrs := make([]backblaze.Transmitted, 0)
		fmt.Fprintf(os.Stderr, " -- Date range: [%s,%s)\n", minStamp, maxStamp)
		for _, file := range files {
			xfrs := parse(file)

			fmt.Fprintf(os.Stderr, " -- Considering: %s %d\n", file, len(xfrs))
			if len(xfrs) > 0 {
				firstDate := xfrs[0].Stamp[0:10]
				inRange := firstDate >= minStamp && firstDate < maxStamp
				if len(xfrs) > 0 && !(inRange) {
					fmt.Fprintf(os.Stderr, " -- Skipping: %s %s\n", firstDate, file)
					continue
				} else {
					fmt.Fprintf(os.Stderr, " -- Keeping: %s %s\n", firstDate, file)
				}
			}

			if doSummary {
				summary := summarize(xfrs)
				sortBySizeThenName(summary)
				writeJSON(summary, "", true)
			}
			allxfrs = append(allxfrs, xfrs...)
		}
		fmt.Fprintf(os.Stderr, "-= Accumulated %d entries\n", len(allxfrs))
		writeJSON(allxfrs, fmt.Sprintf("%sFlow", host), false)
	}
}

// Sorts the passed in slice, in place
//
//	Sort is guaranteed Stable
func sortBySizeThenName(list []backblaze.Transmitted) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Size == list[j].Size {
			return list[i].FName < list[j].FName // FName lexicographical ascending
		}
		return list[i].Size > list[j].Size // Size descending
	})

}

func parse(file string) []backblaze.Transmitted {
	infile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	return backblaze.ParseTransmited(infile)

}

func parent(path string) string {
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-1]
	}
	dir, _ := filepath.Split(path)
	return dir
}

func summarize(xfrs []backblaze.Transmitted) []backblaze.Transmitted {
	fmt.Fprintf(os.Stderr, "-= Writing %d entries\n", len(xfrs))
	tree := make(map[string]*backblaze.Transmitted)
	for _, tx := range xfrs {
		// walk up the current path
		// fmt.Println("fname", tx.FName)
		dir := parent(tx.FName)
		for len(dir) > 0 {
			// fmt.Println("dir", dir)
			// add to current dir
			_, ok := tree[dir]
			if !ok {
				tree[dir] = &backblaze.Transmitted{}
				tree[dir].FName = dir
				tree[dir].Stamp = tx.Stamp[0:10]
			}
			tree[dir].Size += tx.Size

			// walk up
			dir = parent(dir)
		}

	}
	// return tree
	list := make([]backblaze.Transmitted, 0, len(tree))
	for _, tx := range tree {
		list = append(list, *tx)
	}
	return list
}

func writeTree(tree map[string]*backblaze.Transmitted) {
	if _, ok := tree["/"]; !ok {
		return // rootless tree
	}
	outfilename := fmt.Sprintf("tree-%s.json", tree["/"].Stamp)
	fmt.Fprintf(os.Stderr, "-= Writing %s (%d entries)\n", outfilename, len(tree))

	outfile, err := os.Create(outfilename)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	enc := json.NewEncoder(outfile)
	err = enc.Encode(tree)
	if err != nil {
		log.Fatal(err)
	}

}

func writeJSON(xfrs []backblaze.Transmitted, filename string, perLine bool) {
	if len(xfrs) == 0 {
		fmt.Fprintf(os.Stderr, "-= Writing %d entries - skipped\n", len(xfrs))
		return
	}
	ext := "json"
	if perLine {
		ext = "jsonl"
	}
	outfilename := fmt.Sprintf("raw-tx-%s.%s", xfrs[0].Stamp[0:10], ext)
	if 0 != len(filename) {
		outfilename = fmt.Sprintf("%s.%s", filename, ext)

	}
	fmt.Fprintf(os.Stderr, "-= Writing %s (%d entries)\n", outfilename, len(xfrs))

	outfile, err := os.Create(outfilename)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	bw := bufio.NewWriter(outfile)

	if !perLine {
		enc := json.NewEncoder(bw)
		enc.Encode(xfrs)
	} else {
		for _, tx := range xfrs {
			// json per line
			txJ, _ := json.Marshal(tx)
			fmt.Fprintf(bw, "%s\n", txJ)
		}
	}

	err = bw.Flush()
	if err != nil {
		log.Fatal(err)
	}
}
