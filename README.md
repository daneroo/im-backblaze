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

## References
- [Lees Byron's _Stacked Graphs – Geometry & Aesthetics_](https://leebyron.com/streamgraph/)
- [Interactive streamgraph](http://bl.ocks.org/WillTurman/4631136)
- [StreamGraph Demos from Flowning Data](https://flowingdata.com/tag/streamgraph/)
- [D3 Zoomable Sunburst](https://bl.ocks.org/mbostock/4348373)
- [D3 Stramgraph](https://beta.observablehq.com/@mbostock/streamgraph-transitions)
- [My BBWorld Twitter StreamGraph](https://github.com/daneroo/socialbuzz)
- [D3 Sunburst tutorial](https://bl.ocks.org/denjn5/e1cdbbe586ac31747b4a304f8f86efa5)
- [Disk And Memory Space Visualization App built with Electron & d3.js](https://github.com/zz85/space-radar)
- [Python diskover app](https://github.com/shirosaidev/diskover)
- [Single Layer](https://github.com/kratsg/uct3_diskspace)


## bzFlow
Attempts to answer the question:

- Which files are being transmited, on an ongoing basis?
```
time go run cmd/bzFlow/bzFlow.go
```

## bzWhyIgnored
Attempts to answer the question:

- Which files are NOT backed up and why ?

```
time go run cmd/bzWhyIgnored/bzWhyIgnored.go

go build cmd/bzWhyIgnored/bzWhyIgnored.go
time bzWhyIgnored
scp -p bzWhyIgnored fermat:Downloads
ssh fermat time Downloads/bzWhyIgnored
```

## Monitor progress during inital upload
```
tail -f /Library/Backblaze.bzpkg/bzdata/bzlogs/bzreports_lastfilestransmitted/$(date +%d).log
sudo /usr/local/sbin/iftop -i en1
```

## Manual exploration

### Temporary copy, in case some files disppear
```
./scripts/clone.sh
```

### Counting things
FILELIST=v0009a98724006e621c1646e011f_root_filelist.dat
```
./scripts/count.sh
```

## Find _skipped / excluded_ files Manually
Use `cut` and `sort` to compare with `diff` `cmp`
```
wc -l ./data/dirac/bzdata/bzbackup/bzfileids.dat
# 1988792 ./data/dirac/bzdata/bzbackup/bzfileids.dat
grep '^f' ./data/dirac/bzdata/bzfilelists/v00*dat|wc -l
# 1993454

cat ./data/dirac/bzdata/bzbackup/bzfileids.dat | cut -f 2 | sort |uniq > compare-fileids-sorted.dat
grep -h '^f' ./data/dirac/bzdata/bzfilelists/v00*dat | cut -f 4 | sort > compare-filelists-sorted.dat

sha1sum compare-file*
bf871d7e8f16540e730bc7233208e9035b6f2a30  compare-fileids-sorted.dat
bf9bedc8f29f8cc1aa0e889afe20aa0d12ca69d1  compare-filelists-sorted.dat

diff -W 240 --suppress-common-lines --side-by-side compare-file*dat
```

### Finding Excluded files explicitly

Compare listed with actual:
```
```
