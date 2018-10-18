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

	lastCombined := Transmitted{}

	for scanner.Scan() {
		line := scanner.Text()

		// skip blank lines
		if 0 == len(strings.TrimSpace(line)) {
			skipped++
			continue
		}

		fields := strings.SplitN(line, " - ", 6)

		// should be 3 or six fields, if second field is blank, must be 3.
		// BUT: Filename may have ' - 's
		if len(fields[1]) == 65 && len(strings.TrimSpace(fields[1])) == 0 && len(fields) > 3 {
			fields = strings.SplitN(line, " - ", 3)
			// fmt.Fprintf(os.Stderr, "=reformat %d %q\n", len(fields), fields)
		}

		tx := Transmitted{}
		tx.Stamp = fields[0]
		if len(fields) == 6 {
			// ignore errors, default struct values are OK
			fmt.Sscanf(strings.TrimSpace(fields[3]), "%d %s", &tx.Speed, &tx.SpeedUnit)
			fmt.Sscanf(strings.TrimSpace(fields[4]), "%d %s", &tx.Size, &tx.SizeUnit)
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
	// fmt.Printf("%#v\n", list)

	return list
}
