
# or loop over /Library/Backblaze.bzpkg/bzdata/bzfilelists/v*.dat


for f in ./data/*/bzdata/bzfilelists/v*.dat; do
  echo "-=-= Analyzing ${f} =-=-"

  echo "-= Total lines =-"
  cat $f |wc -l

  echo "-=-= Directories =-=-"
  cat $f | grep '^# Dir:' | wc -l

  echo "-=-= Non Directories by type =-=-"
  # only found f(iles) s(ymbolic link)
  cat $f | grep -v '^#' | cut -f 1 | sort | uniq -c

  echo 
done

echo "-=-= Search for an iso file =-=-"
time grep -r /Users/daniel/Downloads/iso .

echo 


