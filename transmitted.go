package backblaze

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Transmitted represent a line in the transmitted logs
type Transmitted struct {
	Type      txRecordType `json:"type"`
	Stamp     string       `json:"stamp"`
	Speed     int          `json:"-"`
	SpeedUnit string       `json:"-"`
	Size      int          `json:"size"`
	SizeUnit  string       `json:"-"`
	Chunk     int          `json:"chunk"`
	FName     string       `json:"fname"`
}

/*
Examples of what we are parsing:

Normal:
2018-10-02 02:39:30 -  large  - throttle manual   11 -  3450 kBits/sec -  7827914 bytes - /Volumes/Space/archive/media/mp3/creative/Binye (Respect)/08-Seourouba.mp3
2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg
# this line is special, has different field lengths
2018-10-17 18:39:45 -  small  - throttle auto     11 -     8 kBits/sec - 1 bytes - /Volumes/Space/fake_filename_to_refresh_volume_dashboard.txt

Dedup (not sent):
2018-10-10 01:40:42 -  small  - throttle x           -           dedup - 0 bytes - /Users/daniel/Library/Containers/com.evernote.Evernote/Data/Library/Application Support/com.evernote.Evernote/puppetmaster/OutputsCache.json
2018-10-01 03:35:31 -  small  - throttle x           -           dedup - 0 bytes - /Users/daniel/.bash_sessions/34D616D0-93F6-4AF2-AD60-9A5D4B83C76A.historynew
2018-10-01 03:35:48 -  small  - throttle x           -           dedup - 0 bytes - /Volumes/Space/archive/media/photo/dadSulbalcon/200308/Catherine35Ans2003/130-3052_IMG.JPG

DedupChunked (not sent):
2018-10-13 10:55:33 -  small  - throttle x           -           dedup - 0 bytes - Chunk 00505 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-13 10:55:33 -  small  - throttle x           -           dedup - 0 bytes - Chunk 00506 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-13 10:55:33 -  small  - throttle x           -           dedup - 0 bytes - Chunk 00507 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2

Combined (Multiple files combined into one transmission):
-CombinedHeader
2018-10-01 15:25:14 -  large  - throttle manual   11 -  3822 kBits/sec - 10469477 bytes - Multiple small files batched in one request, the 3 files are listed below:
-CombinedContinued
2018-10-01 15:25:14 -                                                                   - /Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG
2018-10-01 15:25:14 -                                                                   - /Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG
2018-10-01 15:25:14 -                                                                   - /Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG

Chunked (1 file split into multiple chunks)
2018-10-02 13:30:58 -  small  - throttle manual   11 -    28 kBits/sec -     7290 bytes - Chunk 00003 of /Volumes/Space/archive/media/ebooks/ebook-1100/The UNIX CD Bookshelf, v3.0 (2003).zip
2018-10-02 13:31:15 -  large  - throttle manual   11 -  4143 kBits/sec - 10486490 bytes - Chunk 00000 of /Volumes/Space/archive/media/ebooks/ebook-1100/The UNIX CD Bookshelf, v3.0 (2003).zip
2018-10-11 10:49:34 -  large  - throttle auto     11 -  1643 kBits/sec -   410714 bytes - Chunk 00519 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-11 10:49:35 -  large  - throttle auto     11 -  1973 kBits/sec -   634794 bytes - Chunk 0052a of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-11 10:49:37 -  large  - throttle auto     11 -  2604 kBits/sec -   834682 bytes - Chunk 00545 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2

*/

// ParseTransmited parses transmitted logs
func ParseTransmited(r io.Reader) []Transmitted {

	scanner := bufio.NewScanner(r)
	list := make([]Transmitted, 0, 1000)
	skipped := 0

	compare := false
	var tx2, tx3 Transmitted
	lastCombined2 := Transmitted{}
	lastCombined3 := Transmitted{}

	for scanner.Scan() {
		line := scanner.Text()

		// Orig 8.1s : Fast 5.3 : both 9.6

		if compare {
			tx2 = splitFields(line, &lastCombined2)
		}
		// tx3, txtyp3 = splitFields(line, &lastCombined3)
		tx3 = splitFieldsFast(line, &lastCombined3)
		countType(tx3.Type, line)

		if tx3.Type == empty {
			skipped++
			continue
		}
		if tx2.Type == combinedHeader {
			lastCombined2 = tx2
		}
		if tx3.Type == combinedHeader {
			lastCombined3 = tx3
		}

		if compare && (tx2 != tx3) {
			fmt.Fprintf(os.Stderr, "UnMatched-2,3\n%#v\n%#v\n%s\n", tx2, tx3, line)
		}
		if tx3.Type == dedup || tx3.Type == dedupChunked || tx3.Type == combinedHeader {
			skipped++
			continue
		}

		list = append(list, tx3)
	}
	// fmt.Fprintf(os.Stderr, "-= Parsed %d lines (%d skipped)\n", len(list), skipped)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Print Counts, and optionally reset
	// fmt.Fprintf(os.Stderr, "|countTypes|=%d %#v\n", len(countTypes), countTypes)
	// countTypes = make(map[txRecordType]int)

	return list
}

type txRecordType string

const (
	empty             txRecordType = "Empty"
	normal            txRecordType = "Normal"
	dedup             txRecordType = "Dedup"
	dedupChunked      txRecordType = "DedupChunked"
	combinedHeader    txRecordType = "CombinedHeader"
	combinedContinued txRecordType = "CombinedContinued"
	chunked           txRecordType = "Chunked"
)

func splitFields(line string, lastCombined *Transmitted) Transmitted {
	tx := Transmitted{}

	if 0 == len(strings.TrimSpace(line)) {
		tx.Type = empty
		return tx
	}

	tx.Type = normal
	fields := strings.SplitN(line, " - ", 6)

	// should be 3 or six fields, if second field is blank, must be 3.
	// BUT: Filename may have ' - 's
	if len(fields[1]) == 65 && len(strings.TrimSpace(fields[1])) == 0 && len(fields) > 3 {
		fields = strings.SplitN(line, " - ", 3)
	}

	tx.Stamp = fields[0]
	if len(fields) == 6 {
		// ignore errors, default struct values are OK
		fmt.Sscanf(strings.TrimSpace(fields[3]), "%d %s", &tx.Speed, &tx.SpeedUnit)
		fmt.Sscanf(strings.TrimSpace(fields[4]), "%d %s", &tx.Size, &tx.SizeUnit)
		// skip deduped
		if tx.Speed == 0 && tx.Size == 0 {
			tx.Type = dedup
		}
	}

	tx.FName = fields[len(fields)-1]
	// if Chunked, will replace fname, and set chunk, on error, no action
	if strings.HasPrefix(tx.FName, "Chunk") {
		_, err := fmt.Sscanf(tx.FName, "Chunk %x of", &tx.Chunk)
		tx.FName = tx.FName[15:len(tx.FName)]
		if tx.Type == dedup {
			tx.Type = dedupChunked
		} else {
			tx.Type = chunked
		}
		if err != nil {
			log.Printf("Unable to parse chunked record:\n%s", line)
			log.Fatal(err)
		}
	}

	if strings.HasPrefix(tx.FName, "Multiple small files batched in one request") {
		_, err := fmt.Sscanf(tx.FName, "Multiple small files batched in one request, the %d files are listed below:", &tx.Chunk)
		if err != nil {
			log.Fatal(err)
		}
		// now spread the size into tx.chunk parts!
		tx.Size = tx.Size / tx.Chunk
		tx.SizeUnit = "bytes*" //estimated
		tx.Type = combinedHeader
	}

	if len(fields) == 3 {
		tx.Chunk = -lastCombined.Chunk
		tx.Size = lastCombined.Size
		tx.SizeUnit = lastCombined.SizeUnit
		tx.Speed = lastCombined.Speed
		tx.SpeedUnit = lastCombined.SpeedUnit

		lastCombined.Chunk-- // combined chunks are numbered -7,-6,..,-1
		tx.Type = combinedContinued
	}

	return tx
}

var countTypes map[txRecordType]int

func countType(typ txRecordType, line string) {
	if countTypes == nil {
		countTypes = make(map[txRecordType]int)
	}
	countTypes[typ] = countTypes[typ] + 1
	// if countTypes[typ] < 3 {
	// 	fmt.Printf("--%s:%d--\n%s\n", typ, countTypes[typ], line)
	// }
}

func splitFieldsFast(line string, lastCombined *Transmitted) Transmitted {
	tx := Transmitted{}

	if 0 == len(strings.TrimSpace(line)) {
		tx.Type = empty
		return tx
	}

	tx.Type = normal
	tx.Stamp = line[0:19]
	if line[65:70] == "dedup" {
		tx.FName = line[83:len(line)]
		tx.SizeUnit = "bytes" // just to conform, but 0 is 0!
		tx.Type = dedup
		//  No other (non-default) fields required
		if strings.HasPrefix(tx.FName, "Chunk") {
			_, err := fmt.Sscanf(tx.FName, "Chunk %x of", &tx.Chunk)
			tx.FName = tx.FName[15:len(tx.FName)]
			if err != nil {
				log.Printf("Unable to parse dedup-chunked record:\n%s", line)
				log.Fatal(err)
			}
			tx.Type = dedupChunked
		}
		// fmt.Printf("|%s|%s|%s|\n", tx.Stamp, mid, tx.FName)
	} else {
		preBeginOfPath := strings.Index(line, " - /")
		if preBeginOfPath == -1 {
			preBeginOfPath = strings.Index(line, " - Chunk")
		}
		if preBeginOfPath == -1 {
			preBeginOfPath = strings.Index(line, " - Multiple")
		}
		if preBeginOfPath != -1 && preBeginOfPath != 87 && preBeginOfPath != 80 {
			fmt.Fprintf(os.Stderr, "-= Unexpected line structure (might be ok)\n")
			fmt.Fprintf(os.Stderr, "-begin: %d |%s|\n", preBeginOfPath, line)
			fmt.Fprintf(os.Stderr, "+begin: %d |%s|\n", preBeginOfPath, line[preBeginOfPath+3:len(line)])
		}
		mid := line[22:preBeginOfPath]
		tx.FName = line[preBeginOfPath+3 : len(line)]

		if len(strings.TrimSpace(mid)) == 0 {
			// combinedContinued
			tx.Type = combinedContinued
			//  No other (non-default) fields required
			tx.Chunk = -lastCombined.Chunk
			tx.Size = lastCombined.Size
			tx.SizeUnit = lastCombined.SizeUnit
			tx.Speed = lastCombined.Speed
			tx.SpeedUnit = lastCombined.SpeedUnit

			lastCombined.Chunk-- // combined chunks are numbered -7,-6,..,-1
		} else if strings.HasPrefix(tx.FName, "Multiple small files batched in one request") {
			// combinedHeader
			tx.Type = combinedHeader
			_, err := fmt.Sscanf(tx.FName, "Multiple small files batched in one request, the %d files are listed below:", &tx.Chunk)
			if err != nil {
				log.Printf("Unable to parse combinedHeader record:\n%s", line)
				log.Fatal(err)
			}
			fields := strings.SplitN(mid, " - ", 4)
			// ignore errors, default struct values are OK
			fmt.Sscanf(strings.TrimSpace(fields[2]), "%d %s", &tx.Speed, &tx.SpeedUnit)
			fmt.Sscanf(strings.TrimSpace(fields[3]), "%d %s", &tx.Size, &tx.SizeUnit)
			// now spread the size into tx.chunk parts!
			tx.Size = tx.Size / tx.Chunk
			tx.SizeUnit = "bytes*" //estimated
		} else if strings.HasPrefix(tx.FName, "Chunk") {
			// chunked
			tx.Type = chunked
			_, err := fmt.Sscanf(tx.FName, "Chunk %x of", &tx.Chunk)
			tx.FName = tx.FName[15:len(tx.FName)]
			if err != nil {
				log.Printf("Unable to parse chunked record:\n%s", line)
				log.Fatal(err)
			}
			fields := strings.SplitN(mid, " - ", 4)
			// ignore errors, default struct values are OK
			fmt.Sscanf(strings.TrimSpace(fields[2]), "%d %s", &tx.Speed, &tx.SpeedUnit)
			fmt.Sscanf(strings.TrimSpace(fields[3]), "%d %s", &tx.Size, &tx.SizeUnit)
			// fmt.Printf("chunked:%d: %s\n", txtyp, line)
			// fmt.Printf("|%s|%s|%s|\n", tx.Stamp, mid, tx.FName)
			// fmt.Printf("chunked:%d: %#v\n", txtyp, tx)
		} else {
			// normal
			tx.Type = normal
			fields := strings.SplitN(mid, " - ", 4)
			// ignore errors, default struct values are OK
			fmt.Sscanf(strings.TrimSpace(fields[2]), "%d %s", &tx.Speed, &tx.SpeedUnit)
			fmt.Sscanf(strings.TrimSpace(fields[3]), "%d %s", &tx.Size, &tx.SizeUnit)
			// fmt.Printf("normal:%d: %s\n", txtyp, line)
			// fmt.Printf("|%s|%s|%s|\n", tx.Stamp, mid, tx.FName)
			// fmt.Printf("normal:%d: %#v\n", txtyp, tx)
		}

	}
	return tx
}
