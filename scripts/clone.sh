
# remove -n
# explore log rotation

EXTRA="--delete --exclude *.future"

for h in galois davinci; do
  echo "Cloning ${h}"
  mkdir -p ./data/${h}
  rsync -azv --progress ${EXTRA} ${h}:/Library/Backblaze.bzpkg/bzdata ./data/${h}
done



