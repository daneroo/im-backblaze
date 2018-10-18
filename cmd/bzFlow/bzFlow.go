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

	"github.com/daneroo/backblaze"
)

// const baseDir = "./data/dirac/bzdata"
// const baseDir = "./data/fermat/bzdata"
const baseDir = "/Library/Backblaze.bzpkg/bzdata"

// bzlogs/bzreports_lastfilestransmitted/13.log

func main() {
	files, err := filepath.Glob(baseDir + "/bzlogs/bzreports_lastfilestransmitted/*.log")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		xfrs := backblaze.ParseTransmited(file)
		// writeJSON(xfrs, false) // one json array '.json'
		// writeJSON(xfrs, true)  //perLine '.jsonl'

		// writeTree(tree)

		summary := summarize(xfrs)
		// include files themselves
		// summary = append(summary, xfrs...)

		sort.Slice(summary, func(i, j int) bool {
			if summary[i].Size == summary[j].Size {
				return summary[i].FName < summary[j].FName // FName lexicographical ascending
			}
			return summary[i].Size > summary[j].Size // Size descending
		})
		writeJSON(summary, true) //perLine '.jsonl'
	}

}

func parent(path string) string {
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-1]
	}
	dir, _ := filepath.Split(path)
	return dir
}

func makeTree() {
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

func writeJSON(xfrs []backblaze.Transmitted, perLine bool) {
	if len(xfrs) == 0 {
		fmt.Fprintf(os.Stderr, "-= Writing %d entries - skipped\n", len(xfrs))
		return
	}
	ext := "json"
	if perLine {
		ext = "jsonl"
	}
	outfilename := fmt.Sprintf("raw-tx-%s.%s", xfrs[0].Stamp[0:10], ext)
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
