#!/bin/bash

set -e
set -o pipefail

export INSTALL_K3S_VERSION="v1.18.6+k3s1"
export INSTALL_K3S_SKIP_START="true"
export INSTALL_K3S_EXEC="server --no-deploy traefik"
export PROMETHEUS_OPERATOR_VERSION="v0.41.0"
export K3S_MANIFEST_DIR="/var/lib/rancher/k3s/server/manifests"

function install_k3s() {
    curl -sfL https://get.k3s.io | sh -
}

function install_prometheus() {
    sudo mkdir -p "${K3S_MANIFEST_DIR}"
    # Lays down a prometheus-operator manifest to be loaded into k3s
    # This can be optional
    sudo curl -JL -q \
        -o "${K3S_MANIFEST_DIR}/prometheus-operator.yaml" \
        https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/${PROMETHEUS_OPERATOR_VERSION}/bundle.yaml
}

function install_kvdi() {
  echo "[INFO]  Fetching latest version of kVDI"
  latest=$(curl https://tinyzimmer.github.io/kvdi/deploy/charts/index.yaml 2> /dev/null | head | grep appVersion | awk '{print$2}')
  tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t 'kvdi')
  trap 'rm -f "$tmpdir"' EXIT

  echo "[INFO]  Downloading kVDI Chart"
  user=$(id -u)
  sudo docker run --rm \
    -u ${user} \
    --net host \
    -e HOME=/workspace \
    -w /workspace \
    -v "${tmpdir}":/workspace \
      alpine/helm:3.2.4 fetch --untar https://tinyzimmer.github.io/kvdi/deploy/charts/kvdi-${latest}.tgz

  echo "[INFO]  You will now have the option to edit the values passed to kVDI"
  read -p "[INFO]  Press any key to continue... " -n1 -s -u1
  ${EDITOR:-vi} "${tmpdir}/kvdi/values.yaml"
  echo

  values=$(cat "${tmpdir}/kvdi/values.yaml" | sed 's/^/    /g')

  # Lay down the HelmChart for kVDI
  cat << EOF | sudo tee "${K3S_MANIFEST_DIR}/kvdi.yaml" 1> /dev/null
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: kvdi
  namespace: kube-system
spec:
  chart: kvdi
  repo: https://tinyzimmer.github.io/kvdi/deploy/charts
  targetNamespace: default
  set:
    vdi.spec.metrics.serviceMonitor.create: "true"
    vdi.spec.metrics.prometheus.create: "true"
    vdi.spec.metrics.grafana.enabled: "true"
  valuesContent: |-
${values}
EOF

  rm -rf "${tmpdir}"
}

function main () {
    if ! which docker &> /dev/null ; then
        echo "You must install docker first!"
        exit 1
    fi
    
    echo "[INFO]  K3s will be installed to your system, you may be asked for your password"
    install_k3s
    
    echo "[INFO]  Installing kVDI manifests"
    install_prometheus
    install_kvdi
    
    echo "[INFO]  Starting k3s service"
    sudo systemctl start k3s
    sleep 5
    
    echo "[INFO]  Waiting for kVDI to start..."
    while ! sudo k3s kubectl get pod 2> /dev/null | grep kvdi-app 1> /dev/null ; do sleep 2 ; done
    sudo k3s kubectl wait pod --for condition=Ready -l vdiComponent=app --timeout=300s

    adminpassword=$(sudo k3s kubectl get secret kvdi-admin-secret -o yaml | grep password | head -n1 | awk '{print$2}' | base64 -d)
    echo
    echo "kVDI is installed and listening on https://0.0.0.0:443. You can login with the following credentials:"
    echo
    echo "    username: admin"
    echo "    password: ${adminpassword}"
    echo
    echo "To install the example DesktopTemplates, run:"
    echo "    sudo k3s kubectl apply -f https://raw.githubusercontent.com/tinyzimmer/kvdi/main/deploy/examples/example-desktop-templates.yaml"
    echo
    echo "To uninstall kVDI you can run:"
    echo "    sudo k3s-uninstall.sh"
    echo
    echo "Thanks for installing kVDI :)"
}

{ 
  main 
}