#!/bin/bash

# do a full update and install some utilities
yum -y update
yum -y install jq htop

# Install and then start/enable docker
amazon-linux-extras install -y docker
systemctl enable docker
systemctl start docker

# start a pihole to block ads and other nasties
docker run -d --name pihole \
  -p 53:53/tcp -p 53:53/udp \
  --dns 1.1.1.1 --dns 8.8.8.8 \
  --hostname pi.hole \
  -e ServerIP="127.0.0.1" \
  -e VIRTUAL_HOST="pi.hole" \
  -e PROXY_LOCATION="pi.hole" \
  --restart=unless-stopped \
  pihole/pihole

# write a resolvconf pointing to the pihole
tee /etc/k3s-resolv.conf <<EOF
nameserver `ifconfig eth0 | grep 'inet ' | awk '{print$2}'`
EOF

# K3s installation args
export INSTALL_K3S_SKIP_START="true"
export INSTALL_K3S_EXEC="server --disable traefik --resolv-conf /etc/k3s-resolv.conf"
export PROMETHEUS_OPERATOR_VERSION="v0.41.0"
export K3S_MANIFEST_DIR="/var/lib/rancher/k3s/server/manifests"

# Install K3s and symlink to global PATH
mkdir -p "$${K3S_MANIFEST_DIR}"
curl -sfL https://get.k3s.io | sh -
ln -s /usr/local/bin/k3s /usr/sbin/k3s

# Download the prometheus-operator manifest
curl -JL -o "$${K3S_MANIFEST_DIR}/prometheus-operator.yaml" \
    https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/$${PROMETHEUS_OPERATOR_VERSION}/bundle.yaml

%{ if use_lets_encrypt }
# If using Let's Encrypt, lay down a traefik v2 manifest
## file must NOT be named traefik.yaml or else k3s will delete it due to '--disable traefik'
tee "$${K3S_MANIFEST_DIR}/traefik-v2.yaml" << EOF
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: traefik
  namespace: kube-system
spec:
  chart: traefik
  repo: https://containous.github.io/traefik-helm-chart
  targetNamespace: kube-system
  chartVersion: 9.0.0
  valuesContent: |-
    securityContext:
      readOnlyRootFilesystem: false
    globalArguments: []
    additionalArguments:
      - --serverstransport.insecureskipverify
      - --certificatesresolvers.tls
      - --certificatesresolvers.tls.acme.email="${acme_email}"
      - --certificatesresolvers.tls.acme.storage=/data/acme.json
      - --certificatesresolvers.tls.acme.tlschallenge
EOF
%{ endif }

# Lay down the kVDI manifest
tee "$${K3S_MANIFEST_DIR}/kvdi.yaml" << EOF
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: kvdi
  namespace: kube-system
spec:
  chart: kvdi
  repo: https://tinyzimmer.github.io/kvdi/deploy/charts
  targetNamespace: default
  valuesContent: |-
    vdi:
      spec:
        app:
          auditLog: true
          replicas: 2
          serviceType: %{ if use_lets_encrypt }ClusterIP%{ else }LoadBalancer%{ endif }
        desktops:
          maxSessionLength: 5m
        auth:
          tokenDuration: 4h
        metrics:
          serviceMonitor:
            create: true
          prometheus:
            create: true
          grafana:
            enabled: true
      templates:
        - metadata:
            name: ubuntu-xfce4
          spec:
            image: quay.io/tinyzimmer/kvdi:ubuntu-xfce4-demo
            imagePullPolicy: IfNotPresent
            resources:
              requests:
                cpu: 500m
                memory: 512Mi
              limits:
                cpu: 1000m
                memory: 1024Mi
            config:
              allowRoot: false
              init: systemd
            tags:
              os: ubuntu
              desktop: xfce4
              applications: minimal
EOF

# Start k3s
systemctl start k3s

%{ if use_lets_encrypt }
# If using Let's Encrypt, configure traefik ingress using cert resolver
sleep 10
# Wait for traefik to come up before applying the ingress
while ! /usr/local/bin/k3s kubectl get pod -A | grep -v helm | grep traefik ; do sleep 2 ; done

# Wait for traefik to be ready to serve the ingress route
/usr/local/bin/k3s kubectl wait pod \
  --for=condition=Ready \
  -l "app.kubernetes.io/instance=traefik" \
  -n kube-system

# Create the ingress route
/usr/local/bin/k3s kubectl apply -f - << EOF
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: kvdi-ingress
  namespace: default
spec:
  entryPoints:
    - websecure
  routes:
  - match: Host(\`${kvdi_hostname}\`)
    kind: Rule
    services:
    - name: kvdi-app
      port: 443
  tls:
    certResolver: tls
EOF
%{ endif }

## To get the admin password from a booted instance
# sudo k3s kubectl get secret kvdi-admin-secret -o json | jq -r .data.password | base64 -d && echo