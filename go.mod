module github.com/tinyzimmer/kvdi

go 1.16

require (
	github.com/containerd/containerd v1.4.3 // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v20.10.3+incompatible // indirect
	github.com/go-ldap/ldap/v3 v3.1.10
	github.com/go-logr/logr v0.3.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/vault v1.6.2
	github.com/hashicorp/vault/api v1.0.5-0.20201001211907-38d91b749c77
	github.com/kennygrant/sanitize v1.2.4
	github.com/mattn/go-pointer v0.0.1
	github.com/mitchellh/mapstructure v1.3.3
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.45.0
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/tinyzimmer/go-glib v0.0.23
	github.com/tinyzimmer/go-gst v0.2.23
	github.com/xlzd/gotp v0.0.0-20181030022105-c8557ba2c119
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.7.0
)

// replace github.com/tinyzimmer/go-gst => ../go-gst
// replace github.com/tinyzimmer/go-glib => ../go-glib
