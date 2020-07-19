set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"

if ! command -v go &>/dev/null; then
    echo "Ensure go command is installed"
    exit 1
fi

tmpdir="$(mktemp -d)"
cleanup() {
	export GO111MODULE="auto"
	echo "+++ Cleaning up temporary GOPATH"
	go clean -modcache
	rm -rf "${tmpdir}"
}
trap cleanup EXIT

# Create fake GOPATH
echo "+++ Creating temporary GOPATH"
export GOPATH="${tmpdir}/go"
echo "+++ Using temporary GOPATH ${GOPATH}"
export GO111MODULE="on"
GOROOT="$(go env GOROOT)"
export GOROOT
mkdir -p "${GOPATH}/src/github.com/tinyzimmer"
gitdir="${GOPATH}/src/github.com/tinyzimmer/kvdi"
cp -r "${REPO_ROOT}" "${gitdir}"
cd "$gitdir"

"${REPO_ROOT}/_bin/refdocs" \
  --config "${REPO_ROOT}/doc/refdocs.json" \
  --template-dir "${REPO_ROOT}/doc/template" \
  --api-dir "github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1" \
  --out-file "${GOPATH}/out.html"


pandoc --from html --to markdown_strict "${GOPATH}/out.html" -o "${REPO_ROOT}/doc/crds.md"
sed -i 's/#kvdi\.io\/v1alpha1\./#/g' ${REPO_ROOT}/doc/crds.md
sed -i 's/#%23kvdi\.io%2fv1alpha1\./#/g' "${REPO_ROOT}/doc/crds.md"

echo "Generated reference documentation"
