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
		xfrs := parseTransmited(file)
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

type transmited struct {
	Stamp     string `json:"stamp"`
	Speed     int    `json:"-"`
	SpeedUnit string `json:"-"`
	Size      int    `json:"size"`
	SizeUnit  string `json:"-"`
	Chunk     int    `json:"chunk"`
	FName     string `json:"fname"`
}

func parent(path string) string {
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-1]
	}
	dir, _ := filepath.Split(path)
	return dir
}

func summarize(xfrs []transmited) []transmited {
	fmt.Fprintf(os.Stderr, "-= Writing %d entries\n", len(xfrs))
	tree := make(map[string]*transmited)
	for _, tx := range xfrs {
		// walk up the current path
		// fmt.Println("fname", tx.FName)
		dir := parent(tx.FName)
		for len(dir) > 0 {
			// fmt.Println("dir", dir)
			// add to current dir
			_, ok := tree[dir]
			if !ok {
				tree[dir] = &transmited{}
				tree[dir].FName = dir
				tree[dir].Stamp = tx.Stamp[0:10]
			}
			tree[dir].Size += tx.Size

			// walk up
			dir = parent(dir)
		}

	}
	// return tree
	list := make([]transmited, 0, len(tree))
	for _, tx := range tree {
		list = append(list, *tx)
	}
	return list
}

func writeTree(tree map[string]*transmited) {
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

/*
Four examples of what we are parsing:
Normal:
2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg

Deduped (not sent):
2018-10-10 01:40:42 -  small  - throttle x           -           dedup - 0 bytes - /Users/daniel/Library/Containers/com.evernote.Evernote/Data/Library/Application Support/com.evernote.Evernote/puppetmaster/OutputsCache.json

Combined (Multiple files combined into one transmission):
2018-10-01 15:25:14 -  large  - throttle manual   11 -  3822 kBits/sec - 10469477 bytes - Multiple small files batched in one request, the 3 files are listed below:
2018-10-01 15:25:14 -                                                                   - /Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG
2018-10-01 15:25:14 -                                                                   - /Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG
2018-10-01 15:25:14 -                                                                   - /Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG

Chunked (1 file split into multiple chunks)
2018-10-11 10:49:34 -  large  - throttle auto     11 -  1643 kBits/sec -   410714 bytes - Chunk 00519 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-11 10:49:35 -  large  - throttle auto     11 -  1973 kBits/sec -   634794 bytes - Chunk 0052a of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-11 10:49:37 -  large  - throttle auto     11 -  2604 kBits/sec -   834682 bytes - Chunk 00545 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
*/

func parseTransmited(infilename string) []transmited {
	fmt.Fprintf(os.Stderr, "-= Parsing %s\n", infilename)
	infile, err := os.Open(infilename)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	scanner := bufio.NewScanner(infile)
	list := make([]transmited, 0, 1000)
	skipped := 0

	lastCombined := transmited{}

	for scanner.Scan() {
		line := scanner.Text()
		// Filename may have ' - 's
		fields := strings.SplitN(line, " - ", 6)
		// should be 3 or six fields, if second field is blank, must be 3.
		if len(fields[1]) == 65 && len(strings.Trim(fields[1], " ")) == 0 && len(fields) > 3 {
			fields = strings.SplitN(line, " - ", 3)
			// fmt.Fprintf(os.Stderr, "=reformat %d %q\n", len(fields), fields)
		}

		tx := transmited{}
		tx.Stamp = fields[0]
		if len(fields) == 6 {
			// ignore errors, default struct values are OK
			fmt.Sscanf(strings.Trim(fields[3], " "), "%d %s", &tx.Speed, &tx.SpeedUnit)
			fmt.Sscanf(strings.Trim(fields[4], " "), "%d %s", &tx.Size, &tx.SizeUnit)
			// skip deduped
			if tx.Speed == 0 && tx.Size == 0 {
				skipped++
				continue
			}
		}

		tx.FName = fields[len(fields)-1]
		// if Chunked, will replace fname, and set chunk, on error, no action
		fmt.Sscanf(tx.FName, "Chunk %x of %s", &tx.Chunk, &tx.FName)

		if strings.HasPrefix(tx.FName, "Multiple small files batched in one request") {
			lastCombined = tx
			_, err := fmt.Sscanf(tx.FName, "Multiple small files batched in one request, the %d files are listed below:", &lastCombined.Chunk)
			if err != nil {
				log.Fatal(err)
			}
			// now spread the size into tx.chunk parts!
			lastCombined.Size = lastCombined.Size / lastCombined.Chunk
			lastCombined.SizeUnit = "bytes*" //estimated
			skipped++
			// fmt.Printf("------- %d - %#v\n", tx.Size, lastCombined)
			continue
		}

		if len(fields) == 3 {
			tx.Chunk = -lastCombined.Chunk
			tx.Size = lastCombined.Size
			tx.SizeUnit = lastCombined.SizeUnit
			tx.Speed = lastCombined.Speed
			tx.SpeedUnit = lastCombined.SpeedUnit

			lastCombined.Chunk-- // combined chunks are numbered -7,-6,..,-1
		}

		if len(fields) == 6 && tx.Speed != 0 {
			// fmt.Printf("%02d | %#v\n", len(fields), tx)
		}
		if len(fields) == 3 {
			// fmt.Printf("%02d | %#v\n", len(fields), tx)
		}
		// fmt.Printf("%02d | %#v\n", len(fields), tx)

		list = append(list, tx)
	}
	fmt.Fprintf(os.Stderr, "-= Parsed %d lines (%d skipped)\n", len(list), skipped)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return list
}

func writeJSON(xfrs []transmited, perLine bool) {
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
