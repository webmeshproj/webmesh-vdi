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
mkdir -p "${GOPATH}/src/github.com/kvdi"
gitdir="${GOPATH}/src/github.com/kvdi/kvdi"
cp -r "${REPO_ROOT}" "${gitdir}"
cd "$gitdir"
echo

# Generate appv1 Docs

echo "+++ Generating appv1 documentation"

"${REPO_ROOT}/bin/refdocs" \
  --config "${REPO_ROOT}/doc/refdocs.json" \
  --template-dir "${REPO_ROOT}/doc/template" \
  --api-dir "github.com/kvdi/kvdi/apis/app/v1" \
  --out-file "${GOPATH}/out.html"

pandoc --from html --to markdown_strict "${GOPATH}/out.html" -o "${REPO_ROOT}/doc/appv1.md"
sed -i 's/#app.kvdi\.io\/v1\./#/g' "${REPO_ROOT}/doc/appv1.md"
sed -i 's/#%23app.kvdi\.io%2fv1\./#/g' "${REPO_ROOT}/doc/appv1.md"
sed -i 's:#\*github\.com/kvdi/kvdi/apis/app/v1\.:#:g' "${REPO_ROOT}/doc/appv1.md"
sed -i 's:\[\]github\.com/kvdi/kvdi/apis/rbac/v1\.Rule:\<a href\=\"rbacv1\.md#Rule\"\>\[\]rbacv1\.Rule\</a\>:g' "${REPO_ROOT}/doc/appv1.md"

echo

# Generate rbacv1 Docs

echo "+++ Generating rbacv1 documentation"

"${REPO_ROOT}/bin/refdocs" \
  --config "${REPO_ROOT}/doc/refdocs.json" \
  --template-dir "${REPO_ROOT}/doc/template" \
  --api-dir "github.com/kvdi/kvdi/apis/rbac/v1" \
  --out-file "${GOPATH}/out.html"

pandoc --from html --to markdown_strict "${GOPATH}/out.html" -o "${REPO_ROOT}/doc/rbacv1.md"
sed -i 's/#rbac.kvdi\.io\/v1\./#/g' "${REPO_ROOT}/doc/rbacv1.md"
sed -i 's/#%rbac.kvdi\.io%2fv1\./#/g' "${REPO_ROOT}/doc/rbacv1.md"
sed -i 's:#\*github\.com/kvdi/kvdi/apis/rbac/v1\.:#:g' "${REPO_ROOT}/doc/rbacv1.md"

echo

# Generate desktopsv1 Docs

echo "+++ Generating desktopsv1 documentation"

"${REPO_ROOT}/bin/refdocs" \
  --config "${REPO_ROOT}/doc/refdocs.json" \
  --template-dir "${REPO_ROOT}/doc/template" \
  --api-dir "github.com/kvdi/kvdi/apis/desktops/v1" \
  --out-file "${GOPATH}/out.html"

pandoc --from html --to markdown_strict "${GOPATH}/out.html" -o "${REPO_ROOT}/doc/desktopsv1.md"
sed -i 's/#desktops.kvdi\.io\/v1\./#/g' "${REPO_ROOT}/doc/desktopsv1.md"
sed -i 's/#%desktops.kvdi\.io%2fv1\./#/g' "${REPO_ROOT}/doc/desktopsv1.md"
sed -i 's:#\*github\.com/kvdi/kvdi/apis/desktops/v1\.:#:g' "${REPO_ROOT}/doc/desktopsv1.md"

echo