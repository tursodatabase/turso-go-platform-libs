set -euo pipefail
calc() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  elif command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$1" | awk '{print $2}'
  else
    echo "No SHA-256 tool available" >&2; exit 1
  fi
}
: > libs/MANIFEST-SHA256.txt
echo "SHA256 checksums:"
shopt -s nullglob
for f in $(find ./libs -type f -name "*.so" -o -name "*.a" -o -name "*.dylib" -o -name "*.dll"); do
  sum="$(calc "$f")"
  rel="${f#libs/}"
  printf "%s  %s\n" "$sum" "$rel" | tee -a libs/MANIFEST-SHA256.txt
  printf "%s\n" "$sum" > "${f}.sha256"
done
echo "Wrote libs/MANIFEST-SHA256.txt and per-file .sha256 sidecars."