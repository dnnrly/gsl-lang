#!/usr/bin/env bash
# Regenerate the .svg diagram renders committed for README display.
#
# Renders every .mmd source referenced by an image in a README (root or
# flagships) with mermaid-cli, then rewrites the svg width="100%" to the
# diagram's natural viewBox size so GitHub renders it content-sized rather
# than stretched to the page width.
#
# Requires: mmdc (mermaid-cli), python3.
set -euo pipefail

here="$(cd "$(dirname "$0")/.." && pwd)"
config="$here/scripts/puppeteer-config.json"
tmpdir="${TMPDIR:-/tmp}/gsl-diagram-render.$$"
trap 'rm -rf "$tmpdir"' EXIT
mkdir -p "$tmpdir"

command -v mmdc >/dev/null || { echo "mmdc (mermaid-cli) not found" >&2; exit 1; }
command -v python3 >/dev/null || { echo "python3 not found" >&2; exit 1; }

rendered=0
skipped=0
for readme in "$here/README.md" "$here"/examples/flagships/*/README.md; do
  dir="$(dirname "$readme")"
  # every image reference of the form (path/to/x.svg)
  while IFS= read -r svg; do
    [ -n "$svg" ] || continue
    src="${svg%.svg}.mmd"
    dst="$dir/$svg"
    [ -f "$dst" ] && [ "$dst" -nt "$dir/$src" ] && { skipped=$((skipped + 1)); continue; }
    [ -f "$dir/$src" ] || { echo "missing .mmd source for $dst" >&2; exit 1; }
    mmdc -p "$config" -i "$dir/$src" -o "$tmpdir/out.svg"
    python3 - "$tmpdir/out.svg" "$dst" <<'PY'
import re, sys
src, dst = sys.argv[1], sys.argv[2]
with open(src) as f:
    content = f.read()
m = re.search(r'<svg[^>]*viewBox="([-\d.]+) ([-\d.]+) ([-\d.]+) ([-\d.]+)"', content)
if not m:
    sys.exit(f"no viewBox in {src}")
w, h = m.group(3), m.group(4)
out = re.sub(r'width="100%"', f'width="{w}" height="{h}"', content, count=1)
with open(dst, 'w') as f:
    f.write(out)
PY
    rendered=$((rendered + 1))
  done < <(grep -oE '\([^)]*\.svg\)' "$readme" | tr -d '()' | sort -u)
done

echo "rendered $rendered diagram(s), skipped $skipped up-to-date"