package main

// Attempts to answer the question:
// - Which files are NOT backed up and why ?

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// const baseDir = "./data/dirac/bzdata"
// const baseDir = "./data/fermat/bzdata"
const baseDir = "/Library/Backblaze.bzpkg/bzdata"

func main() {
	fileIds := parseFileIds()
	// write("compare-fileids-sorted.dat", fileIds)

	fileLists := parseFileLists()
	// write("compare-filelists-sorted.dat", fileLists)

	missingOnDisk, notBackedUp := diff(fileIds, fileLists)
	reportMissingOnDisk(missingOnDisk)
	reportNotBackedUp(notBackedUp)
}

func reportMissingOnDisk(missingOnDisk []string) {
	fmt.Fprintf(os.Stderr, "aNotInB (Missing on Disk): %d\n", len(missingOnDisk))
}

// see /Library/Backblaze.bzpkg/bzdata/bzexcluderules_mandatory.xml
// see /Library/Backblaze.bzpkg/bzdata/bzexcluderules_editable.xml
// wab~,vmc,vhd,vhdx,vdi,vo1,vo2,vsv,vud,iso,dmg,sparseimage,sys,cab,
// exe,msi,dll,dl_,wim,ost,o,qtch,log,ithmb,vmdk,vmem,vmsd,vmsn,vmss,vmx,vmxf,
// menudata,appicon,appinfo,pva,pvs,pvi,pvm,fdd,hds,drk,mem,nvram,hdd
func reportNotBackedUp(notBackedUp []string) {
	ignoredSuffix := map[string]int{
		".lockn":     0,
		".ds_store":  0,
		".localized": 0,
		".log":       0,
		".exe":       0,
		".dmg":       0,
		".iso":       0,
		".sys":       0,
		".o":         0,
		".ithmb":     0,
		".dll":       0,
		".vmdk":      0,
		".vdi":       0,
		".msi":       0,
	}
	// regexp.MustCompile("(gopher){2}")
	ignoredRules := map[string]int{
		"(?i)^/(volumes/[[:alnum:]]+/)?.bzvol/(README.txt|bzvol_id.xml)$": 0,
		"(?i)^/.file":           0,
		"(?i)^/users/.*/.trash": 0,
		"(?i)^/users/.*/library/.*saved application state":               0,
		"(?i)^/users/.*/library/logs/":                                   0,
		"(?i)^/users/.*/iphoto library/thumbnails/":                      0,
		"(?i)^/users/.*/iphoto library.migratedphotolibrary/thumbnails/": 0,
		"(?i)^/users/.*/photos library.photoslibrary/thumbnails/":        0,
		"(?i)^/users/.*/movies/.*render files/":                          0,
		// <excludefname_rule plat="mac" osVers="*"  ruleIsOptional="f" skipFirstCharThenStartsWith="users/" contains_1="/itunes/" contains_2="*" doesNotContain="*" endsWith="*" hasFileExtension="ipsw" />                             <!-- iPod software updates -->
		"(?i)^/users/.*/itunes/.*\\.ipsw$":             0,
		"(?i)^/users/.*/itunes/itunes music/podcasts/": 0,
		"(?i)^/users/.*/itunes/itunes media/podcasts/": 0,

		"(?i)^/users/.*/library/mail/v2/maildata/envelope index(-shm|-wal)?$": 0,
		"(?i)^/users/.*/library/mail/v2/maildata/availablefeeds(-shm|-wal)?$": 0,
		"(?i)^/users/.*/library/mail/v2/maildata/defaultcounts$":              0,
		"(?i)^/users/.*/library/mail/v2/maildata/lsmmap2$":                    0,

		"(?i)^/users/.*/library/safari/(icons|historyindex\\.sk|webpageicons\\.db)": 0,

		"(?i)^/users/shared/blizzard": 0,

		"(?i)^/users/.*/library/application support/syncservices/local/":                      0,
		"(?i)^/users/.*/library/application support/google/chrome/.*safe browsing":            0,
		"(?i)^/users/.*/library/application support/google/chrome/default/.*history":          0,
		"(?i)^/users/.*/library/application support/google/chrome/default/.*thumbnails":       0,
		"(?i)^/users/.*/library/application support/google/chrome/default/.*archived history": 0,
		// minus the bookmarks
		"(?i)^/users/.*/library/application support/firefox": 0,
		// minus the bookmarks
		"(?i)^/users/.*/library/cache/firefox/profiles":               0,
		"(?i)^/users/.*/library/caches/":                              0,
		"(?i)^/developer":                                             0,
		"(?i)^/users/.*/library/developer/.*shared/documentation/":    0,
		"(?i)^/users/.*/library/developer/.*xcode/ios devicesupport/": 0,
	}
	fmt.Fprintf(os.Stderr, "bNotInA (Not Backed Up): %d\n", len(notBackedUp))
	unaccounted := 0
	ignoredRulesREs := make(map[string](*regexp.Regexp))
	for k := range ignoredRules {
		ignoredRulesREs[k] = regexp.MustCompile(k)
	}
	for _, line := range notBackedUp {
		accountedFor := false
		for k := range ignoredRules {
			re := ignoredRulesREs[k]
			if re.MatchString(line) {
				// fmt.Fprintf(os.Stderr, "Matched: %s (%s)\n", line, k)
				ignoredRules[k]++
				accountedFor = true
			}
		}
		for k := range ignoredSuffix {
			if strings.HasSuffix(strings.ToLower(line), k) {
				ignoredSuffix[k]++
				accountedFor = true
			}
		}
		if !accountedFor {
			fmt.Fprintf(os.Stderr, "NotBackedUp: %s\n", line)
			unaccounted++
		}
	}
	fmt.Fprintf(os.Stderr, "NotBackedUp: total: %d\n", len(notBackedUp))
	fmt.Fprintf(os.Stderr, "NotBackedUp: unaccounted: %d\n", unaccounted)
	// fmt.Fprintf(os.Stderr, "NotBackedUp: %v\n", ignoredSuffix)
	fmt.Fprintf(os.Stderr, "NotBackedUp: Ignored by Suffix\n")
	for k := range ignoredSuffix {
		fmt.Fprintf(os.Stderr, " %9d : %s\n", ignoredSuffix[k], k)
	}
	// fmt.Fprintf(os.Stderr, "NotBackedUp: %v\n", ignoredRules)
	fmt.Fprintf(os.Stderr, "NotBackedUp: Ignored by Rule\n")
	for k := range ignoredRules {
		fmt.Fprintf(os.Stderr, " %9d : %s\n", ignoredRules[k], k)
	}
}

// How about some tests (assume sorted?)
func diff(as, bs []string) (aNotInB, bNotInA []string) {
	aNotInB = make([]string, 0)
	bNotInA = make([]string, 0)
	eq := 0
	a, b := 0, 0
	for a < len(as) && b < len(bs) {
		cmp := strings.Compare(as[a], bs[b])
		if cmp == 0 {
			// fmt.Fprintf(os.Stderr, "Equal: a[%d] = b[%d]=%s,%s \n", a, b, as[a], bs[b])
			eq++
			a++
			b++
		} else if cmp < 0 { // a < b
			// fmt.Fprintf(os.Stderr, "Missing in b (Missing on Disk): a[%d] = %s \n", a, as[a])
			aNotInB = append(aNotInB, as[a])
			a++
		} else if cmp > 0 { // a > b
			// fmt.Fprintf(os.Stderr, "Missing in a (Not Backed Up): b[%d] = %s \n", b, bs[b])
			bNotInA = append(bNotInA, bs[b])
			b++
		}
		// fmt.Fprintf(os.Stderr, "Compare: %d %d \n", a, b)
	}
	fmt.Fprintf(os.Stderr, "Equal: %d\n", eq)
	return aNotInB, bNotInA
}
func parseFileLists() []string {

	files, err := filepath.Glob(baseDir + "/bzfilelists/v*filelist.dat")
	if err != nil {
		log.Fatal(err)
	}

	lines := make([]string, 0)
	for _, file := range files {
		morelines := extractField(file, 3, filterFilelist)
		lines = append(lines, morelines...)
	}
	lines = sortAndUniq(lines)
	return lines
}
func parseFileIds() []string {
	const fileids = baseDir + "/bzbackup/bzfileids.dat"
	lines := extractField(fileids, 1, filterFileIds)
	lines = sortAndUniq(lines)
	return lines
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
