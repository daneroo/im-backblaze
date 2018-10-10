# Backblaze monitoring

In the course of it's operation Backblaze leaves these log files, and we wish to answer some questions using them; such as:

- Explicitly which files (and directories) are excluded/skipped
- Which files (and sizes) are uploaded on an ongoing basis
- Vizualize the operation of continuous backup

## Backblaze log files: (on dirac)
- `/Library/Backblaze.bzpkg/bzdata/bzlogs/bzfilelist/bzfilelist$(date +%d).log`: filelist process report
- `/Library/Backblaze.bzpkg/bzdata/bzinfo.xml`: All info on backup. Which volumes, schedule, excluded dirs
- `/Library/Backblaze.bzpkg/bzdata`
- `/.bzvol/bzvol_id.xml` Volume identifier
- `/Volumes/Space/.bzvol/bzvol_id.xml` Volume identifier

Lists are in: `/Library/Backblaze.bzpkg/bzdata/bzfilelists`

- Golang; See [encoding/csv](https://www.socketloop.com/tutorials/golang-read-tab-delimited-file-with-encoding-csv-package) for parsing tsv

## Monitor progress
```
tail -f /Library/Backblaze.bzpkg/bzdata/bzlogs/bzreports_lastfilestransmitted/$(date +%d).log
```

## Find _skipped / excluded_ files
Use `cut` and `sort` to compare with `diff` `cmp`
```
wc -l ./data/dirac/bzdata/bzbackup/bzfileids.dat
# 1988792 ./data/dirac/bzdata/bzbackup/bzfileids.dat
grep '^f' ./data/dirac/bzdata/bzfilelists/v00*dat|wc -l
# 1993454

cat ./data/dirac/bzdata/bzbackup/bzfileids.dat | cut -f 2 | sort > compare-fileids-sorted.dat
grep -h '^f' ./data/dirac/bzdata/bzfilelists/v00*dat | cut -f 4 | sort > compare-filelists-sorted.dat

diff -W 240 --suppress-common-lines --side-by-side compare-file*dat
```

### With golang
```
go run cmd/parseFileId.go >compare-fileids-unsorted-go.dat
go run cmd/parseFileId.go | sort >compare-fileids-sorted-go.dat
```

### Temporary copy, in case some files disppear
```
./clone.sh
```

### Counting things
FILELIST=v0009a98724006e621c1646e011f_root_filelist.dat
```
./count.sh
```

### Finding Excluded files explicitly

Compare listed with actual:
```
```
