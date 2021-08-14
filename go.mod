module github.com/kvdi/kvdi

go 1.16

require (
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/docker/docker v20.10.8+incompatible // indirect
	github.com/go-ldap/ldap/v3 v3.3.0
	github.com/go-logr/logr v0.4.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/vault v1.8.1
	github.com/hashicorp/vault/api v1.1.2-0.20210713235431-1fc8af4c041f
	github.com/jmespath/go-jmespath v0.4.0
	github.com/kennygrant/sanitize v1.2.4
	github.com/mattn/go-pointer v0.0.1
	github.com/mitchellh/mapstructure v1.4.1
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.49.0
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/tinyzimmer/go-glib v0.0.24
	github.com/tinyzimmer/go-gst v0.2.30
	github.com/xlzd/gotp v0.0.0-20181030022105-c8557ba2c119
	golang.org/x/crypto v0.0.0-20210813211128-0a44fdfbc16e
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d
	golang.org/x/oauth2 v0.0.0-20210810183815-faf39c7919d5
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.6
)

// replace github.com/tinyzimmer/go-gst => ../go-gst
// replace github.com/tinyzimmer/go-glib => ../go-glib
