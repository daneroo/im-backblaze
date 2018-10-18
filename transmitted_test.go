package backblaze

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func ss(line string) (string, string, string) {
	var s1, s2, s3 string
	s1 = line[0:19]
	if line[65:70] == "dedup" {
		s2 = line[22:80]
		s3 = line[83 : len(line)-1]
	} else {
		s2 = line[22:87]
		s3 = line[90 : len(line)-1]
	}
	return s1, s2, s3
}

func BenchmarkIndex(b *testing.B) {
	line := "2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		strings.Index(line[80:len(line)], " -  /") // 64 ns
		// strings.Index(line, " -  /") // 120 ns
		// strings.LastIndex(line, " -  /") // 304 ns
		// strings.LastIndex(line[80:len(line)], " -  /") // 160 ns
	}
}
func BenchmarkLastIndex(b *testing.B) {
	line := "2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		/*fields :=*/
		strings.LastIndex(line, " -  /")
	}
}

func BenchmarkSplitSlice(b *testing.B) {
	line := "2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"
	b.ResetTimer()
	// sum := 0
	for i := 0; i < b.N; i++ {
		ss(line)
		// s1, s2, s3 := ss(line)
		// sum += len(s1) + len(s2) + len(s3)
	}
	//fmt.Printf("sum=%d", sum)
}

func BenchmarkSplitN6(b *testing.B) {
	line := "2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		/*fields :=*/ strings.SplitN(line, " - ", 6)
	}
}
func BenchmarkSplitN3(b *testing.B) {
	line := "2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		/*fields :=*/ strings.SplitN(line, " - ", 3)
	}
}

func BenchmarkSpliFields(b *testing.B) {
	lastCombined := Transmitted{}
	line := "2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		splitFields(line, &lastCombined)
	}
}
func BenchmarkSpliFieldsFast(b *testing.B) {
	lastCombined := Transmitted{}
	line := "2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		splitFieldsFast(line, &lastCombined)
	}
}

func TestSplitFieldsByType(t *testing.T) {
	var data = []struct {
		name string
		in   string
		out  []Transmitted
	}{
		{
			name: "Empty",
			in:   "\n  \n \t \n",
			out: []Transmitted{
				Transmitted{Type: "Empty", Stamp: "", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "", Chunk: 0, FName: ""},
				Transmitted{Type: "Empty", Stamp: "", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "", Chunk: 0, FName: ""},
				Transmitted{Type: "Empty", Stamp: "", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "", Chunk: 0, FName: ""},
				Transmitted{Type: "Empty", Stamp: "", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "", Chunk: 0, FName: ""},
			},
		},
		{
			name: "Normal",
			in: `2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg
2018-10-17 18:39:45 -  small  - throttle auto     11 -     8 kBits/sec - 1 bytes - /Volumes/Space/fake_filename_to_refresh_volume_dashboard.txt
2018-10-02 02:39:30 -  large  - throttle manual   11 -  3450 kBits/sec -  7827914 bytes - /Volumes/Space/archive/media/mp3/creative/Binye (Respect)/08-Seourouba.mp3
2018-10-02 02:39:36 -  large  - throttle manual   11 -  4972 kBits/sec -  7832042 bytes - /Volumes/Space/archive/media/mp3/peered/Brazil-Rodrigo/Cantoria 1 - Elomar, Geraldo Azevedo, Vital Faria e Xangai - 1984/09 Cantiga do Estradar.mp3`,
			out: []Transmitted{
				Transmitted{Type: "Normal", Stamp: "2018-10-02 13:27:18", Speed: 3112, SpeedUnit: "kBits/sec", Size: 30460266, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"},
				Transmitted{Type: "Normal", Stamp: "2018-10-17 18:39:45", Speed: 8, SpeedUnit: "kBits/sec", Size: 1, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/fake_filename_to_refresh_volume_dashboard.txt"},
				Transmitted{Type: "Normal", Stamp: "2018-10-02 02:39:30", Speed: 3450, SpeedUnit: "kBits/sec", Size: 7827914, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/mp3/creative/Binye (Respect)/08-Seourouba.mp3"},
				Transmitted{Type: "Normal", Stamp: "2018-10-02 02:39:36", Speed: 4972, SpeedUnit: "kBits/sec", Size: 7832042, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/mp3/peered/Brazil-Rodrigo/Cantoria 1 - Elomar, Geraldo Azevedo, Vital Faria e Xangai - 1984/09 Cantiga do Estradar.mp3"},
			},
		},
		{
			name: "Dedup",
			in: `2018-10-01 03:35:31 -  small  - throttle x           -           dedup - 0 bytes - /Users/daniel/.bash_sessions/34D616D0-93F6-4AF2-AD60-9A5D4B83C76A.historynew
2018-10-01 03:35:48 -  small  - throttle x           -           dedup - 0 bytes - /Volumes/Space/archive/media/photo/dadSulbalcon/200308/Catherine35Ans2003/130-3052_IMG.JPG`,
			out: []Transmitted{
				Transmitted{Type: "Dedup", Stamp: "2018-10-01 03:35:31", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 0, FName: "/Users/daniel/.bash_sessions/34D616D0-93F6-4AF2-AD60-9A5D4B83C76A.historynew"},
				Transmitted{Type: "Dedup", Stamp: "2018-10-01 03:35:48", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/photo/dadSulbalcon/200308/Catherine35Ans2003/130-3052_IMG.JPG"},
			},
		},
		{
			name: "DedupChunked",
			in: `2018-10-02 13:32:57 -  small  - throttle x           -           dedup - 0 bytes - Chunk 00000 of /Users/daniel/GoogleDrive/Jobs/Sologlobe/Sologlobe  Mar 08,2013  03 40 PM.QBB
2018-10-02 13:32:57 -  small  - throttle x           -           dedup - 0 bytes - Chunk 00001 of /Users/daniel/GoogleDrive/Jobs/Sologlobe/Sologlobe  Mar 08,2013  03 40 PM.QBB`,
			out: []Transmitted{
				Transmitted{Type: "DedupChunked", Stamp: "2018-10-02 13:32:57", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 0, FName: "/Users/daniel/GoogleDrive/Jobs/Sologlobe/Sologlobe  Mar 08,2013  03 40 PM.QBB"},
				Transmitted{Type: "DedupChunked", Stamp: "2018-10-02 13:32:57", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 1, FName: "/Users/daniel/GoogleDrive/Jobs/Sologlobe/Sologlobe  Mar 08,2013  03 40 PM.QBB"},
			},
		},
		{
			name: "Combined-Header-Continued",
			in: `2018-10-01 15:25:14 -  large  - throttle manual   11 -  3822 kBits/sec - 10469477 bytes - Multiple small files batched in one request, the 3 files are listed below:
2018-10-01 15:25:14 -                                                                   - /Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG
2018-10-01 15:25:14 -                                                                   - /Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG
2018-10-01 15:25:14 -                                                                   - /Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG`,
			out: []Transmitted{
				Transmitted{Type: "CombinedHeader", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: 3, FName: "Multiple small files batched in one request, the 3 files are listed below:"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -3, FName: "/Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -2, FName: "/Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -1, FName: "/Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG"},
			},
		},
		{
			name: "Chunked",
			in: `2018-10-11 10:49:34 -  large  - throttle auto     11 -  1643 kBits/sec -   410714 bytes - Chunk 00519 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-11 10:49:35 -  large  - throttle auto     11 -  1973 kBits/sec -   634794 bytes - Chunk 0052a of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-11 10:49:37 -  large  - throttle auto     11 -  2604 kBits/sec -   834682 bytes - Chunk 00545 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2`,
			out: []Transmitted{
				Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:34", Speed: 1643, SpeedUnit: "kBits/sec", Size: 410714, SizeUnit: "bytes", Chunk: 1305, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:35", Speed: 1973, SpeedUnit: "kBits/sec", Size: 634794, SizeUnit: "bytes", Chunk: 1322, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:37", Speed: 2604, SpeedUnit: "kBits/sec", Size: 834682, SizeUnit: "bytes", Chunk: 1349, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
			},
		},
	}
	for _, tt := range data {
		lines := strings.Split(tt.in, "\n")
		list := make([]Transmitted, 0)
		lastCombined := Transmitted{}

		for _, line := range lines {
			tx, txtyp := splitFields(line, &lastCombined)
			// tx, txtyp := splitFieldsFast(line, &lastCombined)
			tx.Type = txtyp
			list = append(list, tx)
			if tx.Type == combinedHeader {
				lastCombined = tx
			}
		}
		got := list
		if !reflect.DeepEqual(tt.out, got) {
			t.Errorf("Test:%s: input:\n%s\nexpected:\n%sgot:\n %s", tt.name, tt.in, vslice(tt.out), vslice(got))
			// t.Errorf("ZZ %#v", got)
		}
	}
}

func TestSplitFields(t *testing.T) {
	var data = []struct {
		filename string
		out      []Transmitted
	}{
		{
			filename: "./test/data/transmitted.log",
			out: []Transmitted{
				// Transmitted{Type: "Normal", Stamp: "2018-10-02 13:27:18", Speed: 3112, SpeedUnit: "kBits/sec", Size: 30460266, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"},
				// Transmitted{Type: "Dedup", Stamp: "2018-10-10 01:40:42", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 0, FName: "/Users/daniel/Library/Containers/com.evernote.Evernote/Data/Library/Application Support/com.evernote.Evernote/puppetmaster/OutputsCache.json"},
				// Transmitted{Type: "CombinedHeader", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: 3, FName: "Multiple small files batched in one request, the 3 files are listed below:"},
				// Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "", Chunk: 0, FName: "/Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG"},
				// Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "", Chunk: 1, FName: "/Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG"},
				// Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "", Chunk: 2, FName: "/Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG"},
				// Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:34", Speed: 1643, SpeedUnit: "kBits/sec", Size: 410714, SizeUnit: "bytes", Chunk: 1305, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				// Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:35", Speed: 1973, SpeedUnit: "kBits/sec", Size: 634794, SizeUnit: "bytes", Chunk: 1322, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				// Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:37", Speed: 2604, SpeedUnit: "kBits/sec", Size: 834682, SizeUnit: "bytes", Chunk: 1349, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},

				Transmitted{Type: "Normal", Stamp: "2018-10-02 13:27:18", Speed: 3112, SpeedUnit: "kBits/sec", Size: 30460266, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"},
				Transmitted{Type: "Dedup", Stamp: "2018-10-10 01:40:42", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 0, FName: "/Users/daniel/Library/Containers/com.evernote.Evernote/Data/Library/Application Support/com.evernote.Evernote/puppetmaster/OutputsCache.json"},
				Transmitted{Type: "CombinedHeader", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: 3, FName: "Multiple small files batched in one request, the 3 files are listed below:"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -3, FName: "/Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -2, FName: "/Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -1, FName: "/Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:34", Speed: 1643, SpeedUnit: "kBits/sec", Size: 410714, SizeUnit: "bytes", Chunk: 1305, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:35", Speed: 1973, SpeedUnit: "kBits/sec", Size: 634794, SizeUnit: "bytes", Chunk: 1322, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:37", Speed: 2604, SpeedUnit: "kBits/sec", Size: 834682, SizeUnit: "bytes", Chunk: 1349, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
			},
		},
		{
			filename: "./test/data/transmitted-sample.log",
			out: []Transmitted{

				Transmitted{Type: "Empty", Stamp: "", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "", Chunk: 0, FName: ""},
				// special case with different width: /Volumes/Space/fake_filename_to_refresh_volume_dashboard.txt
				Transmitted{Type: "Normal", Stamp: "2018-10-17 18:39:45", Speed: 8, SpeedUnit: "kBits/sec", Size: 1, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/fake_filename_to_refresh_volume_dashboard.txt"},
				Transmitted{Type: "Normal", Stamp: "2018-10-02 02:39:30", Speed: 3450, SpeedUnit: "kBits/sec", Size: 7827914, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/mp3/creative/Binye (Respect)/08-Seourouba.mp3"},
				Transmitted{Type: "Normal", Stamp: "2018-10-02 02:39:36", Speed: 4972, SpeedUnit: "kBits/sec", Size: 7832042, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/mp3/peered/Brazil-Rodrigo/Cantoria 1 - Elomar, Geraldo Azevedo, Vital Faria e Xangai - 1984/09 Cantiga do Estradar.mp3"},
				Transmitted{Type: "Dedup", Stamp: "2018-10-01 03:35:31", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 0, FName: "/Users/daniel/.bash_sessions/34D616D0-93F6-4AF2-AD60-9A5D4B83C76A.historynew"},
				Transmitted{Type: "Dedup", Stamp: "2018-10-01 03:35:48", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/photo/dadSulbalcon/200308/Catherine35Ans2003/130-3052_IMG.JPG"},
				Transmitted{Type: "DedupChunked", Stamp: "2018-10-02 13:32:57", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 0, FName: "/Users/daniel/GoogleDrive/Jobs/Sologlobe/Sologlobe  Mar 08,2013  03 40 PM.QBB"},
				Transmitted{Type: "DedupChunked", Stamp: "2018-10-02 13:32:57", Speed: 0, SpeedUnit: "", Size: 0, SizeUnit: "bytes", Chunk: 1, FName: "/Users/daniel/GoogleDrive/Jobs/Sologlobe/Sologlobe  Mar 08,2013  03 40 PM.QBB"},
				Transmitted{Type: "CombinedHeader", Stamp: "2018-10-01 00:00:18", Speed: 3429, SpeedUnit: "kBits/sec", Size: 1616376, SizeUnit: "bytes*", Chunk: 7, FName: "Multiple small files batched in one request, the 7 files are listed below:"},
				Transmitted{Type: "CombinedHeader", Stamp: "2018-10-01 00:00:22", Speed: 2632, SpeedUnit: "kBits/sec", Size: 1645021, SizeUnit: "bytes*", Chunk: 7, FName: "Multiple small files batched in one request, the 7 files are listed below:"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 00:00:18", Speed: 2632, SpeedUnit: "kBits/sec", Size: 1645021, SizeUnit: "bytes*", Chunk: -7, FName: "/Volumes/Space/archive/media/photo/dad/2003/2003_08_23/129-2919_IMG.JPG"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 00:00:18", Speed: 2632, SpeedUnit: "kBits/sec", Size: 1645021, SizeUnit: "bytes*", Chunk: -6, FName: "/Volumes/Space/archive/media/photo/dad/2003/2003_07_06/125-2583_IMG.JPG"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-02 13:30:58", Speed: 28, SpeedUnit: "kBits/sec", Size: 7290, SizeUnit: "bytes", Chunk: 3, FName: "/Volumes/Space/archive/media/ebooks/ebook-1100/Over 1100 General Computer Ebooks/The UNIX CD Bookshelf, v3.0 (2003).zip"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-02 13:31:15", Speed: 4143, SpeedUnit: "kBits/sec", Size: 10486490, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/ebooks/ebook-1100/Over 1100 General Computer Ebooks/The UNIX CD Bookshelf, v3.0 (2003).zip"},
			},
		},
	}

	for _, tt := range data {
		infile, err := os.Open(tt.filename)
		if err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(infile)
		list := make([]Transmitted, 0)
		lastCombined := Transmitted{}
		for scanner.Scan() {
			line := scanner.Text()
			tx, txtyp := splitFields(line, &lastCombined)
			// tx, txtyp := splitFieldsFast(line, &lastCombined)
			tx.Type = txtyp
			list = append(list, tx)
			if tx.Type == combinedHeader {
				lastCombined = tx
			}

		}
		got := list
		if !reflect.DeepEqual(tt.out, got) {
			// t.Errorf("Test:%s\nexpected:\n%sgot:\n %s", tt.filename, vslice(tt.out), vslice(got))
			t.Errorf("ZZ %#v", got)
		}
	}

}

func TestParseTransmited(t *testing.T) {
	var data = []struct {
		filename string
		out      []Transmitted
	}{
		{
			filename: "./test/data/transmitted.log",
			out: []Transmitted{
				Transmitted{Type: "Normal", Stamp: "2018-10-02 13:27:18", Speed: 3112, SpeedUnit: "kBits/sec", Size: 30460266, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -3, FName: "/Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -2, FName: "/Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -1, FName: "/Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:34", Speed: 1643, SpeedUnit: "kBits/sec", Size: 410714, SizeUnit: "bytes", Chunk: 1305, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:35", Speed: 1973, SpeedUnit: "kBits/sec", Size: 634794, SizeUnit: "bytes", Chunk: 1322, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-11 10:49:37", Speed: 2604, SpeedUnit: "kBits/sec", Size: 834682, SizeUnit: "bytes", Chunk: 1349, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
			},
		},
		{
			filename: "./test/data/transmitted-sample.log",
			out: []Transmitted{

				Transmitted{Type: "Normal", Stamp: "2018-10-17 18:39:45", Speed: 8, SpeedUnit: "kBits/sec", Size: 1, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/fake_filename_to_refresh_volume_dashboard.txt"},
				Transmitted{Type: "Normal", Stamp: "2018-10-02 02:39:30", Speed: 3450, SpeedUnit: "kBits/sec", Size: 7827914, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/mp3/creative/Binye (Respect)/08-Seourouba.mp3"},
				Transmitted{Type: "Normal", Stamp: "2018-10-02 02:39:36", Speed: 4972, SpeedUnit: "kBits/sec", Size: 7832042, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/mp3/peered/Brazil-Rodrigo/Cantoria 1 - Elomar, Geraldo Azevedo, Vital Faria e Xangai - 1984/09 Cantiga do Estradar.mp3"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 00:00:18", Speed: 2632, SpeedUnit: "kBits/sec", Size: 1645021, SizeUnit: "bytes*", Chunk: -7, FName: "/Volumes/Space/archive/media/photo/dad/2003/2003_08_23/129-2919_IMG.JPG"},
				Transmitted{Type: "CombinedContinued", Stamp: "2018-10-01 00:00:18", Speed: 2632, SpeedUnit: "kBits/sec", Size: 1645021, SizeUnit: "bytes*", Chunk: -6, FName: "/Volumes/Space/archive/media/photo/dad/2003/2003_07_06/125-2583_IMG.JPG"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-02 13:30:58", Speed: 28, SpeedUnit: "kBits/sec", Size: 7290, SizeUnit: "bytes", Chunk: 3, FName: "/Volumes/Space/archive/media/ebooks/ebook-1100/Over 1100 General Computer Ebooks/The UNIX CD Bookshelf, v3.0 (2003).zip"},
				Transmitted{Type: "Chunked", Stamp: "2018-10-02 13:31:15", Speed: 4143, SpeedUnit: "kBits/sec", Size: 10486490, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/ebooks/ebook-1100/Over 1100 General Computer Ebooks/The UNIX CD Bookshelf, v3.0 (2003).zip"}},
		},
	}
	for _, tt := range data {
		got := ParseTransmited(tt.filename)
		if !reflect.DeepEqual(tt.out, got) {
			t.Errorf("Test:%s\nexpected:\n%sgot:\n %s", tt.filename, vslice(tt.out), vslice(got))
			// t.Errorf("ZZ %#v", got)
		}
	}
}

func vslice(s []Transmitted) string {
	var str string
	for _, i := range s {
		str += fmt.Sprintf("%#v\n", i)
	}
	return str
}
