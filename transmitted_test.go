package backblaze

import (
	"fmt"
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

func TestParseTransmited_Parts(t *testing.T) {
	var data = []struct {
		name string
		in   string
		out  []Transmitted
	}{
		{
			name: "Empty lines",
			in:   "\n",
			out:  []Transmitted{},
		},
		{
			name: "Normal",
			in:   "2018-10-02 13:27:18 -  large  - throttle manual   11 -  3112 kBits/sec - 30460266 bytes - /Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg",
			out: []Transmitted{
				Transmitted{Stamp: "2018-10-02 13:27:18", Speed: 3112, SpeedUnit: "kBits/sec", Size: 30460266, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"},
			},
		},
		{
			name: "Deduped (not sent)",
			in:   "2018-10-10 01:40:42 -  small  - throttle x           -           dedup - 0 bytes - /Users/daniel/Library/Containers/com.evernote.Evernote/Data/Library/Application Support/com.evernote.Evernote/puppetmaster/OutputsCache.json",
			out:  []Transmitted{},
		},
		{
			name: "Combined (Multiple files combined into one transmission)",
			in: `2018-10-01 15:25:14 -  large  - throttle manual   11 -  3822 kBits/sec - 10469477 bytes - Multiple small files batched in one request, the 3 files are listed below:
2018-10-01 15:25:14 -                                                                   - /Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG
2018-10-01 15:25:14 -                                                                   - /Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG
2018-10-01 15:25:14 -                                                                   - /Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG`,
			out: []Transmitted{
				Transmitted{Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -3, FName: "/Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG"},
				Transmitted{Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -2, FName: "/Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG"},
				Transmitted{Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -1, FName: "/Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG"},
			},
		},
		{
			name: "Chunked (1 file split into multiple chunks)",
			in: `2018-10-11 10:49:34 -  large  - throttle auto     11 -  1643 kBits/sec -   410714 bytes - Chunk 00519 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-11 10:49:35 -  large  - throttle auto     11 -  1973 kBits/sec -   634794 bytes - Chunk 0052a of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2
2018-10-11 10:49:37 -  large  - throttle auto     11 -  2604 kBits/sec -   834682 bytes - Chunk 00545 of /Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2`,
			out: []Transmitted{
				Transmitted{Stamp: "2018-10-11 10:49:34", Speed: 1643, SpeedUnit: "kBits/sec", Size: 410714, SizeUnit: "bytes", Chunk: 1305, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				Transmitted{Stamp: "2018-10-11 10:49:35", Speed: 1973, SpeedUnit: "kBits/sec", Size: 634794, SizeUnit: "bytes", Chunk: 1322, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
				Transmitted{Stamp: "2018-10-11 10:49:37", Speed: 2604, SpeedUnit: "kBits/sec", Size: 834682, SizeUnit: "bytes", Chunk: 1349, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
			},
		},
	}
	for _, tt := range data {
		got := parseTransmited(strings.NewReader(tt.in))
		if !reflect.DeepEqual(tt.out, got) {
			t.Errorf("Test:%s: input:\n%s\nexpected:\n%sgot:\n %s", tt.name, tt.in, vslice(tt.out), vslice(got))
		}
	}
}
func TestParseTransmited(t *testing.T) {
	got := ParseTransmited("./test/data/transmitted.log")
	expected := []Transmitted{
		Transmitted{Stamp: "2018-10-02 13:27:18", Speed: 3112, SpeedUnit: "kBits/sec", Size: 30460266, SizeUnit: "bytes", Chunk: 0, FName: "/Volumes/Space/archive/media/video/PMB/12-23-2008(1)/20081219122438.mpg"},
		Transmitted{Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -3, FName: "/Volumes/Space/archive/media/photo/catou/2005_11_02-R/IMG_0927.JPG"},
		Transmitted{Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -2, FName: "/Volumes/Space/archive/media/photo/catou/2007-07-04-lesours/IMG_4941.JPG"},
		Transmitted{Stamp: "2018-10-01 15:25:14", Speed: 3822, SpeedUnit: "kBits/sec", Size: 3489825, SizeUnit: "bytes*", Chunk: -1, FName: "/Users/daniel/GoogleDrive/Google Photos/2013/12/IMG_1490.JPG"},
		Transmitted{Stamp: "2018-10-11 10:49:34", Speed: 1643, SpeedUnit: "kBits/sec", Size: 410714, SizeUnit: "bytes", Chunk: 1305, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
		Transmitted{Stamp: "2018-10-11 10:49:35", Speed: 1973, SpeedUnit: "kBits/sec", Size: 634794, SizeUnit: "bytes", Chunk: 1322, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
		Transmitted{Stamp: "2018-10-11 10:49:37", Speed: 2604, SpeedUnit: "kBits/sec", Size: 834682, SizeUnit: "bytes", Chunk: 1349, FName: "/Users/daniel/Library/Containers/com.docker.docker/Data/vms/0/Docker.qcow2"},
	}
	if len(got) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(got))
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Test: expected:\n %s got:\n %s", vslice(expected), vslice(got))
	}

	// fmt.Printf("%#v\n", got)
}

func vslice(s []Transmitted) string {
	var str string
	for _, i := range s {
		str += fmt.Sprintf("%#v\n", i)
	}
	return str
}
