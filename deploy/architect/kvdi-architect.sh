#!/bin/bash

# K3s/Prometheus constants
export INSTALL_K3S_VERSION="v1.18.6+k3s1"
export INSTALL_K3S_SKIP_START="true"
export INSTALL_K3S_EXEC="server --no-deploy traefik"
export PROMETHEUS_OPERATOR_VERSION="v0.41.0"
export K3S_MANIFEST_DIR="/var/lib/rancher/k3s/server/manifests"

# Installs K3s using the official installer. Configurations are passed
# via the above variables.
function install-k3s() {
    echo "[INFO]  K3s will be installed to your system"
    curl -sfL https://get.k3s.io | sh -
}

# Starts K3s and waits for 5 seconds. The wait gives a little extra time for the 
# API server to be able to serve requests.
function start-k3s() {
    dialog --backtitle "kVDI Architect" \
        --infobox "Starting k3s service" 5 50
    systemctl start k3s
    sleep 5
}

# Installs the prometheus-operator manifest to the k3s directory. It will be auto-applied
# by k3s at start.
function install-prometheus() {
    # Lays down a prometheus-operator manifest to be loaded into k3s
    # This can be made optional
    echo "[INFO]  Installing prometheus-operator manifests"
    curl -JL -q \
        -o "${K3S_MANIFEST_DIR}/prometheus-operator.yaml" \
        https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/${PROMETHEUS_OPERATOR_VERSION}/bundle.yaml 2> /dev/null
}

export dialogTmp=$(mktemp -d 2>/dev/null || mktemp -d -t 'kvdi-dialog')
trap "rm -rf '${dialogTmp}'" EXIT

# Runs a dialog with built-in and arbitrary arguments
function do-dialog() {
    dialog \
        --backtitle "kVDI Architect" \
        --cancel-label "Quit" \
        "${@}" 2> "${dialogTmp}/dialog.ans"
}

# Retrieves the answer from a dialog
function get-dialog-answer() {
    cat "${dialogTmp}/dialog.ans"
}

# Fetches the given helm chart version and returns a temporary path where
# it is extracted.
function fetch-helm-chart() {
    local version=${1}

    tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t 'kvdi')

    docker run --rm \
        -u $(id -u) \
        --net host \
        -e HOME=/workspace \
        -w /workspace \
        -v "${tmpdir}":/workspace \
            alpine/helm:3.2.4 \
            fetch --untar https://tinyzimmer.github.io/kvdi/deploy/charts/kvdi-${version}.tgz 1> /dev/null

    echo "${tmpdir}"
}

# Writes the given arguments to the kvdi values file.
function write-to-values() {
    local chart_dir=${1}
    local key=${2}
    local value=${3}

    dialog --sleep 1 --backtitle "kVDI Architect" \
        --infobox "Setting ${key}=${value} to kVDI Configuration" 5 100
  
    docker run --rm \
        -v "${chart_dir}":/workdir \
        mikefarah/yq \
        yq write -i kvdi/values.yaml "${key}" "${value}"
  
    if [[ "${?}" != "0" ]] ; then
        echo "Error writing to values: $*"
        exit
    fi
}

# Deletes the given object from the values
function delete-from-values() {
    local chart_dir=${1}
    local key=${2}

    dialog --sleep 1 --backtitle "kVDI Architect" \
        --infobox "Setting ${key}=null to kVDI Configuration" 5 100

    docker run --rm \
        -v "${chart_dir}":/workdir \
        mikefarah/yq \
        yq delete -i kvdi/values.yaml "${key}"
}

# Prompts for configurations for LDAP auth and writes them to the values file
function set-ldap-config() {
    local chart_dir=${1}

    # Get the URL
    do-dialog --inputbox \
        "What is the URL of the LDAP server?" \
        10 50 "ldaps://ldap.example.com:636"
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi

    url="$(get-dialog-answer)"
    write-to-values "${chart_dir}" "vdi.spec.auth.ldapAuth.URL" "${url}"

    # If using TLS, ask if a CA is provided
    if [[ "${url}" =~ "ldaps" ]]  ; then
        while true ; do
            do-dialog --extra-button --extra-label Reset \
                --title "Paste the CA certificate for the LDAP user (leave empty if not required or disabling verification)" \
                --editbox /tmp/test 0 0
            case $? in
                0 ) break ;;
                1 ) echo "Aborting" && exit ;;
                3 ) continue ;;
            esac
        done
    fi

    ca=$(get-dialog-answer)

    # If using TLS and no CA is provided, ask if disabling TLS verification
    if [[ "${url}" =~ "ldaps" ]] && [[ "${ca//[[:blank:]]/}" == "" ]] ; then
        do-dialog --defaultno  --yesno "Disable TLS Verification?" 0 0
        if [[ "${?}" == "0" ]] ; then
            write-to-values "${chart_dir}" \
              "vdi.spec.auth.ldapAuth.tlsInsecureSkipVerify" true
        fi
    # CA is provided so write it to the values
    elif [[ "${ca//[[:blank:]]/}" != "" ]] ; then
        write-to-values "${chart_dir}" \
          "vdi.spec.auth.ldapAuth.tlsCACert" "$(echo "${ca}" | base64 --wrap=0)"
    fi

    # Determine the root DN for the ldap server from the URL
    # Used to autopopulate fields
    ldapHost=$(echo "${url}" | sed 's=.*://==' | sed 's=:.*==')
    ldapBase=$(echo "${ldapHost}" | awk 'BEGIN{FS=OFS="."}{for(i=1; i<=NF; i++) {printf "dc=%s,", $i}}')

    # Get the UserSearchBase
    do-dialog --inputbox \
        "What search base should be used for users?" \
        10 80 "ou=users,${ldapBase%,}"
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    write-to-values "${chart_dir}" "vdi.spec.auth.ldapAuth.userSearchBase" "$(get-dialog-answer)"    

    # Get the initial admin group
    do-dialog --inputbox \
        "What LDAP group should have initial admin access? (You can add more later)" \
        10 80 "ou=kvdi-admins,${ldapBase%,}"
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    write-to-values "${chart_dir}" "vdi.spec.auth.ldapAuth.adminGroups[+]" "$(get-dialog-answer)"

    # Get the bind user
    do-dialog --inputbox \
        "What is the full DN of the LDAP user for kVDI?" \
        10 80 "cn=kvdi-user,ou=svcaccts,${ldapBase%,}"
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    bindUser=$(get-dialog-answer)

    # Get the bind password
    do-dialog --insecure --passwordbox \
        "What is the password for the LDAP bind user?" 10 50
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    bindPassw=$(get-dialog-answer)

    # Write credentials to a K8s secret to be loaded into k3s
    tee "${K3S_MANIFEST_DIR}/kvdi-secrets.yaml" 1> /dev/null << EOF
apiVersion: v1
kind: Secret
metadata:
  name: kvdi-app-secrets
  namespace: default
data:
  ldap-userdn: $(echo "${bindUser}" | base64 --wrap=0)
  ldap-password: $(echo "${bindPassw}" | base64 --wrap=0)

EOF

    # Ensure other auth providers are not configured
    delete-from-values "${chart_dir}" "vdi.spec.auth.localAuth"
    delete-from-values "${chart_dir}" "vdi.spec.auth.oidcAuth"
}

function set-oidc-config() {
    local chart_dir=${1}

    # Get the IssuerURL
    do-dialog --inputbox \
        "What is the Issuer URL of the authentication provider?" \
        10 50 "https://auth.example.com/authorize"
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi

    url="$(get-dialog-answer)"
    write-to-values "${chart_dir}" "vdi.spec.auth.oidcAuth.issuerURL" "${url}"

    # Get the RedirectURL
    do-dialog --inputbox \
        "What is the Redirect URL for this oauth client? \n(This should be the URI of this server followed by /api/login)" \
        10 60 "https://kvdi.example.com/api/login"
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi

    url="$(get-dialog-answer)"
    write-to-values "${chart_dir}" "vdi.spec.auth.oidcAuth.redirectURL" "${url}"   

    # If using HTTPS, get the CA if required
    if [[ "${url}" =~ "https" ]]  ; then
        while true ; do
            do-dialog --extra-button --extra-label Reset \
                --title "Paste the CA certificate for the authentication provider (leave empty if not required or disabling verification)" \
                --editbox /tmp/test 0 0
            case $? in
                0 ) break ;;
                1 ) echo "Aborting" && exit ;;
                3 ) continue ;;
            esac
        done
    fi

    ca=$(get-dialog-answer)

    # If using HTTPS and no CA is provided, ask if disabling TLS verification
    # otherwise write the CA to the values.
    if [[ "${url}" =~ "https" ]] && [[ "${ca//[[:blank:]]/}" == "" ]] ; then
        do-dialog --defaultno  --yesno "Disable TLS Verification?" 0 0
        if [[ "${?}" == "0" ]] ; then
            write-to-values "${chart_dir}" "vdi.spec.auth.oidcAuth.tlsInsecureSkipVerify" true
        fi
    elif [[ "${ca//[[:blank:]]/}" != "" ]] ; then
        write-to-values "${chart_dir}" "vdi.spec.auth.oidcAuth.tlsCACert" "$(echo "${ca}" | base64 --wrap=0)"
    fi

    # Get the client id
    do-dialog --inputbox \
        "What is the ClientID for the authentication provider?" \
        10 80 ""
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    clientID=$(get-dialog-answer)

    # Get the client secret
    do-dialog --insecure --passwordbox \
        "What is the ClientSecret for the authentication provider?" 10 50
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    clientSecret=$(get-dialog-answer)

    # Write credentials to a K8s secret to be loaded into k3s
    tee "${K3S_MANIFEST_DIR}/kvdi-secrets.yaml" 1> /dev/null << EOF
apiVersion: v1
kind: Secret
metadata:
  name: kvdi-app-secrets
  namespace: default
data:
  oidc-clientid: $(echo "${clientID}" | base64 --wrap=0)
  oidc-clientsecret: $(echo "${clientSecret}" | base64 --wrap=0)

EOF

    # Get the scopes
    do-dialog --checklist \
      "                    Select the authentication scopes to request\n(If you need to provide custom values say 'yes' when prompted for additional changes)" \
      0 90 5 \
      "openid" "" "on" \
      "email" "" "on" \
      "profile" "" "on" \
      "groups" "" "on"
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    scopes=$(get-dialog-answer)
    for scope in "${scopes[@]}" ; do 
        write-to-values "${chart_dir}" \
            "vdi.spec.auth.oidcAuth.scopes[+]" "${scope}"
    done

    # Get the group scope
    do-dialog --inputbox \
        "What is the 'groups' scope for the authentication provider?\n(Answer yes to 'Allow all authenticated' if you don't have one)" \
        10 80 "groups"
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    groupScope=$(get-dialog-answer)
    write-to-values "${chart_dir}" "vdi.spec.auth.oidcAuth.groupScope" "${groupScope}"

    # Get the initial admin group
    do-dialog --inputbox \
        "What openid group should have initial admin access? (You can add more later)" \
        10 80 "kvdi-admins"
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    write-to-values "${chart_dir}" "vdi.spec.auth.oidcAuth.adminGroups[+]" "$(get-dialog-answer)"

    # Check if allowing all authenticated users
    do-dialog --defaultno  --yesno "Allow all authenticated users?" 0 0
    if [[ "${?}" == "0" ]] ; then
        write-to-values "${chart_dir}" "vdi.spec.auth.oidcAuth.allowNonGroupedReadOnly" true
    fi

    # Ensure other auth providers are not configured
    delete-from-values "${chart_dir}" "vdi.spec.auth.localAuth"
    delete-from-values "${chart_dir}" "vdi.spec.auth.ldapAuth"
}

# Configures authentication values for kVDI
function set-auth-config() {
    local chart_dir=${1}

    # Figure out which auth backend we are using
    do-dialog --radiolist \
        "Select the authentication method" 15 60 10 \
        "Local" "Use local built-in authentication" "on" \
        "LDAP" "Use LDAP for authentication" "" \
        "OIDC" "Use OpenID/OAuth for authentication" ""
    if [[ "${?}" != "0" ]] ; then echo "Aborting" && exit ; fi
    
    backend=$(get-dialog-answer)

    # If LDAP/OIDC prompt for configuration options
    case ${backend} in
        "LDAP" )  set-ldap-config "${chart_dir}" ;;
        "OIDC" )  set-oidc-config "${chart_dir}" ;;
        "Local")  export AUTH_METHOD="Local" ; \
                  delete-from-values "${chart_dir}" "vdi.spec.auth.ldapAuth" ; \
                  delete-from-values "${chart_dir}" "vdi.spec.auth.oidcAuth" ;;
    esac
}

# Sets custom server certificate options
function set-tls-config() {
    local chart_dir="${1}"

    # Check if using custom certificate
    do-dialog --defaultno --yesno "Use a pre-existing TLS server certificate?" 5 50
    if [[ "${?}" != "0" ]] ; then return 0 ; fi

    # Get the certificate
    while true ; do
        do-dialog --extra-button --extra-label Reset \
            --title "Paste the PEM-encoded TLS Server Certificate" \
            --editbox /tmp/test 0 0
        case $? in
            0 ) break ;;
            1 ) echo "Aborting" && exit ;;
            3 ) continue ;;
        esac
    done
    serverCert=$(get-dialog-answer)

    # Get the private key
    while true ; do
        do-dialog --extra-button --extra-label Reset \
            --title "Paste the PEM-encoded, unencrypted TLS Server Private Key" \
            --editbox /tmp/test 0 0
        case $? in
            0 ) break ;;
            1 ) echo "Aborting" && exit ;;
            3 ) continue ;;
        esac
    done
    serverKey=$(get-dialog-answer)

    # Write a TLS secret
    tee "${K3S_MANIFEST_DIR}/kvdi-tls.yaml" 1> /dev/null << EOF
apiVersion: v1
kind: Secret
metadata:
  name: kvdi-app-external-tls
  namespace: default
data:
  tls.crt: $(echo ${serverCert} | base64 --wrap=0)
  tls.key: $(echo ${serverKey} | base64 --wrap=0)
EOF

    # Point kVDI at the external TLS secret
    write-to-values "${chart_dir}" "vdi.spec.app.tls.serverSecret" "kvdi-app-external-tls"
}

# Fetches the kvdi helm chart, reads in the values, and calls the appropriate
# functions for prompting for configurations.
# Once all configurations are gathered, a HelmChart manifest is written to the K3s
# load directory.
function install-kvdi() {
    local version=${1}
    local dry_run=${2}

    # Figure out latest chart version if not supplied by user
    if [[ "${version}" == "" ]] ; then
        dialog --backtitle "kVDI Architect" \
            --infobox "Fetching latest version of kVDI" 5 50
        version=$(curl https://tinyzimmer.github.io/kvdi/deploy/charts/index.yaml 2> /dev/null | head | grep appVersion | awk '{print$2}')
        sleep 1
    fi

    # Fetch the helm chart
    dialog --backtitle "kVDI Architect" \
        --infobox "Downloading kVDI Chart: ${version}" 5 50 &
    dpid=${!}
    chart_dir=$(fetch-helm-chart ${version})
    trap "rm -rf '${tmpdir}'" EXIT

    # Enable metrics by default for now
    write-to-values "${chart_dir}" "vdi.spec.metrics.serviceMonitor.create" true
    write-to-values "${chart_dir}" "vdi.spec.metrics.prometheus.create" true
    write-to-values "${chart_dir}" "vdi.spec.metrics.grafana.enabled" true

    kill ${dpid}

    # Check if user wants to supply custom TLS
    set-tls-config "${chart_dir}"

    # Set auth configurations
    set-auth-config "${chart_dir}"

    ## need to implement more vault auth methods to have it make sense
    ## in a standalone setup
    # set-secrets-backend "${install_mode}" "${chart_dir}"

    do-dialog --defaultno --yesno "Do you want to open the kVDI values for additional changes?" 5 70
    if [[ "${?}" == "0" ]] ; then
        ${EDITOR:-vi} "${chart_dir}/kvdi/values.yaml"
    fi

    # Lay down the HelmChart for kVDI
    if [[ "${dry_run}" == "false" ]] ; then
      tee "${K3S_MANIFEST_DIR}/kvdi.yaml" 1> /dev/null << EOF
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
$(cat "${chart_dir}/kvdi/values.yaml" | sed 's/^/    /g')

EOF

    else
        echo
        echo "# DRY RUN OUTPUT #"
        echo
        echo "# values.yaml"
        echo
        cat "${chart_dir}/kvdi/values.yaml" 
        echo
        if [[ -f "${K3S_MANIFEST_DIR}/kvdi-secrets.yaml" ]] ; then
            echo
            echo "# kvdi-secrets.yaml"
            cat "${K3S_MANIFEST_DIR}/kvdi-secrets.yaml"
            echo
        fi
        if [[ -f "${K3S_MANIFEST_DIR}/kvdi-tls.yaml" ]] ; then
            echo
            echo "# kvdi-tls.yaml"
            cat "${K3S_MANIFEST_DIR}/kvdi-tls.yaml"
            echo
        fi
    fi
}

# Installs the core requirements for kVDI
function install-base() {
    touch "${dialogTmp}/install-base.log"

    dialog --clear --no-cancel --backtitle "kVDI Architect" \
        --tailbox "${dialogTmp}/install-base.log" 30 140 &
    dpid=${!}

    sleep 1

    install-k3s &> "${dialogTmp}/install-base.log"    
    install-prometheus &> "${dialogTmp}/install-base.log"

    sleep 1

    kill ${dpid}
}

# Waits for kVDI app instance to be present and ready
function wait-for-kvdi() {
    dialog --backtitle "kVDI Architect" --cancel-label "Quit" \
        --infobox "Waiting for kVDI to start..." 5 50
    while ! k3s kubectl get pod 2> /dev/null | grep kvdi-app 1> /dev/null ; do sleep 2 ; done
    k3s kubectl wait pod --for condition=Ready -l vdiComponent=app --timeout=300s > /dev/null
}

# Prints initial instructions to the user after a successful install
function print-instructions() {
    echo
    echo "kVDI is installed and listening on https://0.0.0.0:443"
    if [[ "${AUTH_METHOD}" == "Local" ]] ; then
        adminpassword=$(k3s kubectl get secret kvdi-admin-secret -o yaml | grep password | head -n1 | awk '{print$2}' | base64 -d)
        echo
        echo "You can login with the following credentials:"
        echo
        echo "    username: admin"
        echo "    password: ${adminpassword}"
    fi
    echo
    echo "To install the example DesktopTemplates, run:"
    echo "    sudo k3s kubectl apply -f https://raw.githubusercontent.com/tinyzimmer/kvdi/main/deploy/examples/example-desktop-templates.yaml"
    echo
    echo "To uninstall kVDI you can run:"
    echo "    ${0} --uninstall"
    echo
    echo "Thanks for installing kVDI :)"
}

function run-install() {
    local version="${1}"
    local dry_run="${2}"

    if ! which docker &> /dev/null ; then
        echo "You must install 'docker' first to use this script."
        exit 1
    fi

    if ! which dialog &> /dev/null ; then
        echo "You must install the 'dialog' package to use this script."
        exit 1
    fi

    set -e
    set -o pipefail

    # Initialize the manifest dir
    mkdir -p "${K3S_MANIFEST_DIR}"

    if [[ "${dry_run}" == "false" ]] ; then install-base ; fi

    # Allow bad exit codes from dialogs
    set +e
    install-kvdi "${version}" "${dry_run}"
    set -e

    if [[ "${dry_run}" == "false" ]] ; then
        start-k3s
        wait-for-kvdi
        echo "# `pwd`/kvdi-out.log" > kvdi-out.log
        print-instructions >> kvdi-out.log
        dialog --clear --backtitle "kVDI Architect" --cancel-label "Quit" \
            --textbox kvdi-out.log 0 0
        clear
    fi
}

function run-uninstall() {
    if ! which k3s-uninstall.sh &> /dev/null ; then
        echo "kVDI does not appear to be installed"
        exit 0
    fi
    echo "kVDI will be uninstalled from the system"
    read -p "[INFO]  Press any key to continue. Ctrl-C to abort. " -n1 -s -u1
    k3s-uninstall.sh
    exit 0
}

function die() { echo "$*" >&2; usage ; }  # complain to STDERR and exit with error
function needs_arg() { if [ -z "$OPTARG" ]; then die "ERROR: --$OPT option requires an argument"; fi; }

function usage() {
  echo "Usage: ${0} [-h|--help] [-u|--uninstall] [-v|--version=<VERSION>] [--dry-run]"
  echo
  echo "Command Line Arguments:"
  echo
  echo "    --dry-run                ) Print helm values and generated secrets from prompts and exit."
  echo "    -u | --uninstall         ) Uninstalls K3s and kVDI from the system."
  echo "    -v | --version=<VERSION> ) Set the version of kVDI to install. Defaults to latest."
  echo "    -h | --help              ) Print this message and exit."
  echo
  exit -1
}

# Main entrypoint
{ 
    # Make sure we are root
    if [[ "$(id -u)" != "0" ]] ; then
        echo "Not running as root, elevating with sudo"
        exec sudo /bin/bash ${0} $*
        exit
    fi

    dry_run="false"
    while getopts hucv:-: OPT; do
        # support long options: https://stackoverflow.com/a/28466267/519360
        if [ "$OPT" = "-" ]; then   # long option: reformulate OPT and OPTARG
            OPT="${OPTARG%%=*}"       # extract long option name
            OPTARG="${OPTARG#$OPT}"   # extract long option argument (may be empty)
            OPTARG="${OPTARG#=}"      # if long option argument, remove assigning `=`
        fi
        case "$OPT" in
            u | uninstall ) run-uninstall ;;
            h | help )      usage ;;
            v | version )   needs_arg; version="$OPTARG" ;;
            dry-run )       dry_run="true" ;;
            ??* )           die "Illegal option --$OPT" ;;  # bad long option
            \? )            exit 2 ;;  # bad short option (error reported via getopts)
        esac
    done

    shift $((OPTIND-1)) # remove parsed options and args from $@ list

    if [[ "${dry_run}" == "true" ]] ; then
        export K3S_MANIFEST_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'kvdi-manifests')
        trap "rm -rf '${K3S_MANIFEST_DIR}'" EXIT
    fi

    run-install "${version}" "${dry_run}"
}