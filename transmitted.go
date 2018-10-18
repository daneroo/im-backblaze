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
	Stamp     string `json:"stamp"`
	Speed     int    `json:"-"`
	SpeedUnit string `json:"-"`
	Size      int    `json:"size"`
	SizeUnit  string `json:"-"`
	Chunk     int    `json:"chunk"`
	FName     string `json:"fname"`
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

// ParseTransmited parses transmitteds logs
func ParseTransmited(infilename string) []Transmitted {
	fmt.Fprintf(os.Stderr, "-= Parsing %s\n", infilename)
	infile, err := os.Open(infilename)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()
	return parseTransmited(infile)
}

func parseTransmited(r io.Reader) []Transmitted {

	scanner := bufio.NewScanner(r)
	list := make([]Transmitted, 0, 1000)
	skipped := 0

	lastCombined2 := Transmitted{}
	lastCombined3 := Transmitted{}

	for scanner.Scan() {
		line := scanner.Text()

		tx2, txtyp2 := splitFields(line, &lastCombined2)
		tx3, txtyp3 := splitFieldsFast(line, &lastCombined3)
		if txtyp2 == empty {
			skipped++
			continue
		}
		if txtyp2 == combinedHeader {
			lastCombined2 = tx2
		}
		if txtyp3 == combinedHeader {
			lastCombined3 = tx3
		}

		if txtyp2 != txtyp3 || tx2 != tx3 {
			fmt.Printf("UnMatched-2,3 %d-%d\n%#v\n%#v\n", txtyp2, txtyp3, tx2, tx3)
		}

		if txtyp3 == dedup || txtyp3 == combinedHeader {
			skipped++
			continue
		}
		list = append(list, tx3)
	}
	fmt.Fprintf(os.Stderr, "-= Parsed %d lines (%d skipped)\n", len(list), skipped)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return list
}

type txRecordType int

const (
	empty txRecordType = iota
	normal
	dedup
	combinedHeader
	combinedContinued
	chunked
)

func splitFields(line string, lastCombined *Transmitted) (Transmitted, txRecordType) {
	tx := Transmitted{}
	txtyp := normal

	if 0 == len(strings.TrimSpace(line)) {
		return tx, empty
	}

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
			txtyp = dedup
		}
	}

	tx.FName = fields[len(fields)-1]
	// if Chunked, will replace fname, and set chunk, on error, no action
	if strings.HasPrefix(tx.FName, "Chunk") {
		_, err := fmt.Sscanf(tx.FName, "Chunk %x of", &tx.Chunk)
		tx.FName = tx.FName[15:len(tx.FName)]
		if txtyp != dedup {
			txtyp = chunked
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
		txtyp = combinedHeader
	}

	if len(fields) == 3 {
		tx.Chunk = -lastCombined.Chunk
		tx.Size = lastCombined.Size
		tx.SizeUnit = lastCombined.SizeUnit
		tx.Speed = lastCombined.Speed
		tx.SpeedUnit = lastCombined.SpeedUnit

		lastCombined.Chunk-- // combined chunks are numbered -7,-6,..,-1
		txtyp = combinedContinued
	}

	return tx, txtyp
}

func splitFieldsFast(line string, lastCombined *Transmitted) (Transmitted, txRecordType) {
	tx := Transmitted{}
	txtyp := normal

	/*
		2018-10-11 10:49:37 -  large  - throttle auto     11 -  2604 kBits/sec -   834682 bytes - Chunk 00545 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
	*/
	if 0 == len(strings.TrimSpace(line)) {
		return tx, empty
	}

	// + empty txRecordType = iota
	// normal
	// + dedup
	// + combinedHeader
	// + combinedContinued
	// chunked

	// n, err := fmt.Sscanf(line, "%19s - %65s - %s", &s1, &s2, &s3)
	tx.Stamp = line[0:19]
	if line[65:70] == "dedup" {
		tx.FName = line[83:len(line)]
		tx.SizeUnit = "bytes" // just to conform, but 0 is 0!
		txtyp = dedup
		//  No other (non-default) fields required
		if strings.HasPrefix(tx.FName, "Chunk") {
			_, err := fmt.Sscanf(tx.FName, "Chunk %x of", &tx.Chunk)
			tx.FName = tx.FName[15:len(tx.FName)]
			if err != nil {
				log.Printf("Unable to parse dedup-chunked record:\n%s", line)
				log.Fatal(err)
			}
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
			fmt.Printf("-begin: %d |%s|\n", preBeginOfPath, line)
			fmt.Printf("+begin: %d |%s|\n", preBeginOfPath, line[preBeginOfPath+3:len(line)])
		}
		mid := line[22:preBeginOfPath]
		tx.FName = line[preBeginOfPath+3 : len(line)]

		// fmt.Printf("|%s|%s|%s|\n", tx.Stamp, mid, tx.FName)

		if len(strings.TrimSpace(mid)) == 0 {
			// combinedContinued
			txtyp = combinedContinued
			//  No other (non-default) fields required
			tx.Chunk = -lastCombined.Chunk
			tx.Size = lastCombined.Size
			tx.SizeUnit = lastCombined.SizeUnit
			tx.Speed = lastCombined.Speed
			tx.SpeedUnit = lastCombined.SpeedUnit

			lastCombined.Chunk-- // combined chunks are numbered -7,-6,..,-1
		} else if strings.HasPrefix(tx.FName, "Multiple small files batched in one request") {
			// combinedHeader
			txtyp = combinedHeader
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
			txtyp = chunked
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
			txtyp = normal
			fields := strings.SplitN(mid, " - ", 4)
			// ignore errors, default struct values are OK
			fmt.Sscanf(strings.TrimSpace(fields[2]), "%d %s", &tx.Speed, &tx.SpeedUnit)
			fmt.Sscanf(strings.TrimSpace(fields[3]), "%d %s", &tx.Size, &tx.SizeUnit)
			// fmt.Printf("normal:%d: %s\n", txtyp, line)
			// fmt.Printf("|%s|%s|%s|\n", tx.Stamp, mid, tx.FName)
			// fmt.Printf("normal:%d: %#v\n", txtyp, tx)
		}

	}

	return tx, txtyp
}
