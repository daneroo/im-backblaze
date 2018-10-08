
# remove -n
# explore log rotation

# EXTRA=""
# EXTRA="--exclude *.future"
# EXTRA="-n --delete --exclude *.future"
EXTRA="-n --exclude *.future"

for h in dirac fermat; do
  echo "Cloning ${h}"
  mkdir -p ./data/${h}
  rsync -azv --progress ${EXTRA} ${h}:/Library/Backblaze.bzpkg/bzdata ./data/${h}
done



