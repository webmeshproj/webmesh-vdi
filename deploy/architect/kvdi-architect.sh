#!/bin/bash

# K3s/Prometheus constants
export INSTALL_K3S_SKIP_START="true"
export INSTALL_K3S_EXEC="server --disable traefik"
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
    systemctl restart k3s
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
trap "rm -rf ${dialogTmp}" EXIT

# Runs a dialog with built-in and arbitrary arguments
function do-dialog() {
    dialog \
        --backtitle "kVDI Architect" \
        --cancel-label "${cancel_label:-Quit}" \
        "${@}" 2> "${dialogTmp}/dialog.ans"
}

# Retrieves the answer from a dialog
function get-dialog-answer() {
    cat "${dialogTmp}/dialog.ans"
}

# Fetches the given helm chart version and returns a temporary path where
# it is extracted.
function fetch-helm-chart() {
    tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t 'kvdi')

    docker run --rm \
        -u $(id -u) \
        --net host \
        -e HOME=/workspace \
        -w /workspace \
        -v "${tmpdir}":/workspace \
            alpine/helm:3.2.4 \
            fetch --untar https://tinyzimmer.github.io/kvdi/deploy/charts/kvdi-${VERSION}.tgz 1> /dev/null

    echo "${tmpdir}"
}

function get-helm-chart() {
    # Figure out latest chart version if not supplied by user
    if [[ "${VERSION}" == "" ]] ; then
        dialog --backtitle "kVDI Architect" \
            --infobox "Fetching latest version of kVDI" 5 50
        export VERSION=$(curl https://tinyzimmer.github.io/kvdi/deploy/charts/index.yaml 2> /dev/null | head | grep appVersion | awk '{print$2}')
        sleep 1
    fi

    # Fetch the helm chart
    dialog --backtitle "kVDI Architect" \
        --infobox "Downloading kVDI Chart: ${VERSION}" 5 50 &

    export CHART_DIR=$(fetch-helm-chart)
}

# Pretty prints the given yaml file
function cat-yaml() {
    local file="${1}"

    # docker run --rm \
    #     -v "$(dirname "${file}")":/workdir \
    #     mikefarah/yq \
    
    yq -P -C read "$(basename "${file}")" 
}

function get-yq() {
    export PATH="${HOME}/.local/bin:${PATH}"

    if which yq &> /dev/null ; then return ; fi

    echo "[INFO] yq does not appear to be installed, installing now"
    echo "[INFO]  Downloading checksums for yq v3.3.2..."

    mkdir -p "${HOME}/.local/bin"
    tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t 'kvdi-manifests')
    trap "rm -rf ${tmpdir}" RETURN

    os=$(uname | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m | tr '[:upper:]' '[:lower:]')
    if [[ "${arch}" == "x86_64" ]] ; then arch="amd64" ; fi
    bin_name="yq_${os}_${arch}"

    curl -JL -o "${tmpdir}/checksums" "https://github.com/mikefarah/yq/releases/download/3.3.2/checksums" 2> /dev/null

    echo "[INFO]  Downloading yq v3.3.2..."

    curl -JL -o "${tmpdir}/yq" "https://github.com/mikefarah/yq/releases/download/3.3.2/${bin_name}" 2> /dev/null

    echo "[INFO]  Verifying checksum"
    binsum=$(sha256sum "${tmpdir}/yq" | awk '{print $1}')
    if ! cat "${tmpdir}/checksums" | grep ${bin_name} | grep ${binsum} ; then
        echo "Downloaded hash for yq did not match!"
        exit 1
    fi

    mv "${tmpdir}/yq" "${HOME}/.local/bin/yq"
    chmod +x "${HOME}/.local/bin/yq"
}

# Writes the given arguments to the kvdi values file.
function write-to-values() {
    local key=${1}
    local value=${2}

    dialog --sleep 1 --backtitle "kVDI Architect" \
        --infobox "Setting ${key}=${value} to kVDI Configuration" 5 100
  
    # docker run --rm \
    #     -v "${CHART_DIR}":/workdir \
    #     mikefarah/yq \

    yq write -i "${CHART_DIR}/kvdi/values.yaml" "${key}" "${value}"
  
    if [[ "${?}" != "0" ]] ; then
        echo "Error writing to values: $*"
        exit
    fi
}

# Reads a key from the values
function read-from-values() {
    local key=${1}

    # docker run --rm \
    #     -v "${CHART_DIR}":/workdir \
    #     mikefarah/yq \
    
    yq read "${CHART_DIR}/kvdi/values.yaml" "${key}"
    
    if [[ "${?}" != "0" ]] ; then
        echo -n
    fi
}

# Deletes the given object from the values
function delete-from-values() {
    local key=${1}

    dialog --sleep 1 --backtitle "kVDI Architect" \
        --infobox "Setting ${key}=null to kVDI Configuration" 5 100

    # docker run --rm \
    #     -v "${CHART_DIR}":/workdir \
    #     mikefarah/yq \

    yq delete -i "${CHART_DIR}/kvdi/values.yaml" "${key}"
}

# Prompts for configurations for LDAP auth and writes them to the values file
function set-ldap-config() {
    # Get the URL
    do-dialog --extra-button --extra-label "Back" \
        --inputbox \
        "What is the URL of the LDAP server?" \
        10 50 "ldaps://ldap.example.com:636"
    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac
    url="$(get-dialog-answer)"

    # If using TLS, ask if a CA is provided
    if [[ "${url}" =~ "ldaps" ]]  ; then
        while true ; do
            touch "${dialogTmp}/edit"
            do-dialog --extra-button --extra-label Reset \
                --title "Paste the CA certificate for the LDAP user (leave empty if not required or disabling verification)" \
                --editbox "${dialogTmp}/edit" 0 0
            case $? in
                0 ) break ;;
                1 ) echo "Aborting" && exit ;;
                3 ) continue ;;
            esac
        done
    fi

    ca=$(get-dialog-answer)

    # Determine the root DN for the ldap server from the URL
    # Used to autopopulate fields
    ldapHost=$(echo "${url}" | sed 's=.*://==' | sed 's=:.*==')
    ldapBase=$(echo "${ldapHost}" | awk 'BEGIN{FS=OFS="."}{for(i=1; i<=NF; i++) {printf "dc=%s,", $i}}')

    # Get the UserSearchBase
    do-dialog --extra-button --extra-label "Back" \
        --inputbox \
        "What search base should be used for users?" \
        10 80 "ou=users,${ldapBase%,}"
    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac
    userBase="$(get-dialog-answer)"

    # Get the initial admin group
    do-dialog --extra-button --extra-label "Back" \
        --inputbox \
        "What LDAP group should have initial admin access? (You can add more later)" \
        10 80 "ou=kvdi-admins,${ldapBase%,}"
    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac
    adminGroup=$(get-dialog-answer)

    # Get the bind user
    do-dialog --extra-button --extra-label "Back" \
        --inputbox \
        "What is the full DN of the LDAP user for kVDI?" \
        10 80 "cn=kvdi-user,ou=svcaccts,${ldapBase%,}"
    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac
    bindUser=$(get-dialog-answer)

    # Get the bind password
    do-dialog --extra-button --extra-label "Back" \
        --insecure --passwordbox \
        "What is the password for the LDAP bind user?" 10 50
    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac
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

    # If using TLS and no CA is provided, ask if disabling TLS verification
    if [[ "${url}" =~ "ldaps" ]] && [[ "${ca//[[:blank:]]/}" == "" ]] ; then
        do-dialog --defaultno  --yesno "Disable TLS Verification?" 0 0
        if [[ "${?}" == "0" ]] ; then
            write-to-values \
              "vdi.spec.auth.ldapAuth.tlsInsecureSkipVerify" true
        fi
    # CA is provided so write it to the values
    elif [[ "${ca//[[:blank:]]/}" != "" ]] ; then
        write-to-values \
          "vdi.spec.auth.ldapAuth.tlsCACert" "$(echo "${ca}" | base64 --wrap=0)"
    fi

    write-to-values "vdi.spec.auth.ldapAuth.URL" "${url}"
    write-to-values "vdi.spec.auth.ldapAuth.userSearchBase" "${userBase}"    
    write-to-values "vdi.spec.auth.ldapAuth.adminGroups[+]" "${adminGroup}"

    # Ensure other auth providers are not configured
    delete-from-values "vdi.spec.auth.localAuth"
    delete-from-values "vdi.spec.auth.oidcAuth"
}

function set-oidc-config() {
    # Get the IssuerURL
    do-dialog --extra-button --extra-label "Back" \
        --inputbox \
        "What is the Issuer URL of the authentication provider?" \
        10 50 "https://auth.example.com/authorize"

    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac

    url="$(get-dialog-answer)"

    # Get the RedirectURL
    do-dialog --extra-button --extra-label "Back" --inputbox \
        "What is the Redirect URL for this oauth client? \n(This should be the URI of this server followed by /api/login)" \
        10 60 "https://kvdi.example.com/api/login"

    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac

    redirectURL="$(get-dialog-answer)"

    # If using HTTPS, get the CA if required
    if [[ "${url}" =~ "https" ]]  ; then
        while true ; do
            touch "${dialogTmp}/edit"
            do-dialog --extra-button --extra-label Reset \
                --title "Paste the CA certificate for the authentication provider (leave empty if not required or disabling verification)" \
                --editbox "${dialogTmp}/edit" 0 0
            case $? in
                0 ) break ;;
                1 ) clear ;
                    echo "Exiting" ;
                    exit ;;
            esac
        done
    fi

    ca=$(get-dialog-answer)

    # Get the client id
    do-dialog --extra-button --extra-label "Back" --inputbox \
        "What is the ClientID for the authentication provider?" \
        10 80 ""
    
    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac
    
    clientID=$(get-dialog-answer)

    # Get the client secret
    do-dialog --extra-button --extra-label "Back" --insecure --passwordbox \
        "What is the ClientSecret for the authentication provider?" 10 50

    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac

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
    do-dialog --extra-button --extra-label "Back" --checklist \
      "Select the authentication scopes to request\n(If you need to provide custom values say 'yes' when prompted for additional changes)" \
      0 90 5 \
      "openid" "" "on" \
      "email" "" "on" \
      "profile" "" "on" \
      "groups" "" "on"

    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac

    scopes=$(get-dialog-answer)

    # Get the group scope
    do-dialog --extra-button --extra-label "Back" --inputbox \
        "What is the 'groups' scope for the authentication provider?\n(Answer yes to 'Allow all authenticated' if you don't have one)" \
        10 80 "groups"

    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac

    groupScope=$(get-dialog-answer)

    # Get the initial admin group
    do-dialog --extra-button --extra-label "Back" --inputbox \
        "What openid group should have initial admin access? (You can add more later)" \
        10 80 "kvdi-admins"

    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  set-auth-config ;;
    esac

    adminGroup=$(get-dialog-answer)

    # Check if allowing all authenticated users
    do-dialog \
        --defaultno  --yesno "Allow all authenticated users?" 0 0
    
    case ${?} in
        0)  write-to-values "vdi.spec.auth.oidcAuth.allowNonGroupedReadOnly" true ;;
        1)  clear ;
            echo "Exiting" ;
            exit ;;
    esac

    # Write the rest of the configurations

    write-to-values "vdi.spec.auth.oidcAuth.issuerURL" "${url}"
    write-to-values "vdi.spec.auth.oidcAuth.redirectURL" "${redirectURL}"   

    # If using HTTPS and no CA is provided, ask if disabling TLS verification
    # otherwise write the CA to the values.
    if [[ "${url}" =~ "https" ]] && [[ "${ca//[[:blank:]]/}" == "" ]] ; then
        do-dialog --defaultno  --yesno "Disable TLS Verification?" 0 0
        if [[ "${?}" == "0" ]] ; then
            write-to-values "vdi.spec.auth.oidcAuth.tlsInsecureSkipVerify" true
        fi
    elif [[ "${ca//[[:blank:]]/}" != "" ]] ; then
        write-to-values "vdi.spec.auth.oidcAuth.tlsCACert" "$(echo "${ca}" | base64 --wrap=0)"
    fi
    for scope in "${scopes[@]}" ; do 
        write-to-values "vdi.spec.auth.oidcAuth.scopes[+]" "${scope}"
    done

    write-to-values "vdi.spec.auth.oidcAuth.groupScope" "${groupScope}"
    write-to-values "vdi.spec.auth.oidcAuth.adminGroups[+]" "${adminGroup}"

    # Bump the access token duration to 12 hours since oauth refresh is not implemented
    write-to-values "vdi.spec.auth.tokenDuration" "12h"
    
    # Ensure other auth providers are not configured
    delete-from-values "vdi.spec.auth.localAuth"
    delete-from-values "vdi.spec.auth.ldapAuth"
}

# Configures authentication values for kVDI
function set-auth-config() {
    # Figure out which auth backend we are using
    do-dialog --extra-button --extra-label "Back" \
        --menu \
        "Select the authentication method" 15 60 10 \
        "Local" "Use local built-in authentication" \
        "LDAP" "Use LDAP for authentication" \
        "OIDC" "Use OpenID/OAuth for authentication"

    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  main-menu ;;
    esac

    backend=$(get-dialog-answer)

    # If LDAP/OIDC prompt for configuration options
    case ${backend} in
        "LDAP" )  set-ldap-config "${CHART_DIR}" ;;
        "OIDC" )  set-oidc-config "${CHART_DIR}" ;;
        "Local")  delete-from-values "vdi.spec.auth.ldapAuth" ;
                  delete-from-values "vdi.spec.auth.oidcAuth" ;
    esac
    
    export AUTH_METHOD="${backend}"
}

# Sets custom server certificate options
function set-tls-config() {
    # Check if using custom certificate
    do-dialog --defaultno --yesno "Use a pre-existing TLS server certificate?" 5 50
    if [[ "${?}" == "0" ]] ; then

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
        write-to-values "vdi.spec.app.tls.serverSecret" "kvdi-app-external-tls"
    fi
}

# Fetches the kvdi helm chart, reads in the values, and calls the appropriate
# functions for prompting for configurations.
# Once all configurations are gathered, a HelmChart manifest is written to the K3s
# load directory.
function install-kvdi() {
    # Enable metrics by default for now
    write-to-values "vdi.spec.metrics.serviceMonitor.create" true
    write-to-values "vdi.spec.metrics.prometheus.create" true
    write-to-values "vdi.spec.metrics.grafana.enabled" true

    # Lay down the HelmChart for kVDI
    if [[ "${DRY_RUN}" == "false" ]] ; then
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
$(sed 's/^/    /g' "${CHART_DIR}/kvdi/values.yaml")

EOF

    else
        clear
        echo
        echo "# DRY RUN OUTPUT #"
        echo
        echo "# values.yaml"
        echo
        cat-yaml "${CHART_DIR}/kvdi/values.yaml" 
        echo
        if [[ -f "${K3S_MANIFEST_DIR}/kvdi-secrets.yaml" ]] ; then
            echo
            echo "# kvdi-secrets.yaml"
            cat-yaml "${K3S_MANIFEST_DIR}/kvdi-secrets.yaml"
            echo
        fi
        if [[ -f "${K3S_MANIFEST_DIR}/kvdi-tls.yaml" ]] ; then
            echo
            echo "# kvdi-tls.yaml"
            cat-yaml "${K3S_MANIFEST_DIR}/kvdi-tls.yaml"
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
    local dots="."
    dialog --backtitle "kVDI Architect" \
        --infobox "Waiting for kVDI to start${dots}" 5 50
    while ! k3s kubectl get pod 2> /dev/null | grep kvdi-app 1> /dev/null ; do
        sleep 2 
        if [[ $(echo ${dots} | wc -c) == 4 ]] ; then
            dots="."
        else
            dots="${dots}."
        fi
        dialog --backtitle "kVDI Architect" \
            --infobox "Waiting for kVDI to start${dots}" 5 50
    done

    if [[ $(read-from-values vdi.spec.metrics.grafana.enabled) == "true" ]] ; then
        ready="2/2"
    else
        ready="1/1"
    fi

    while ! k3s kubectl get pod -l vdiComponent=app | grep ${ready} 1> /dev/null ; do
        sleep 2 
        if [[ $(echo ${dots} | wc -c) == 4 ]] ; then
            dots="."
        else
            dots="${dots}."
        fi
        dialog --backtitle "kVDI Architect" \
            --infobox "Waiting for kVDI to start${dots}" 5 50
    done
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
    set -e
    set -o pipefail

    if [[ "${DRY_RUN}" == "false" ]] ; then 
        install-base
    fi

    install-kvdi

    if [[ "${DRY_RUN}" == "false" ]] ; then
        start-k3s
        wait-for-kvdi
        echo "# `pwd`/kvdi-out.log" > kvdi-out.log
        print-instructions >> kvdi-out.log
        dialog --clear --backtitle "kVDI Architect" --cancel-label "Quit" \
            --textbox kvdi-out.log 0 0
        clear
    fi

    exit 0
}

function userdata-menu() {
    local userdata_enabled="false"

    current_config="$(read-from-values "vdi.spec.userdataSpec")"
    if [[ "${current_config}" != "{}" ]] && [[ "${current_config}" != "" ]] ; then
        userdata_enabled="true"
    fi

    if [[ "${userdata_enabled}" == "false" ]] ; then
        do-dialog --extra-button --extra-label "Back" \
            --menu "Userdata Persistence Options" 0 0 0 \
            "Enable Persistence" "Enable userdata persistence"
        case ${?} in
            1)  clear ;
                echo "Exiting" ;
                exit ;;
            3)  main-menu ;;
        esac

        write-to-values "vdi.spec.userdataSpec.accessModes[+]" "ReadWriteOnce"
        write-to-values "vdi.spec.userdataSpec.resources.requests.storage" "1Gi"
    fi

    do-dialog --extra-button --extra-label "Back" \
        --menu "Userdata Persistence Options" 0 0 0 \
        "Disable Persistence" "Disable userdata persistence" \
        "Disk Quota" "Configure the amount of storage allotted to each user" \
        "Storage Class" "Configure the storage class for user volumes" \
        "AccessMode" "Set the access mode for volumes. (EXPERIMENTAL: ReadWriteMany volumes are not thoroughly tested)" \

    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  main-menu ;;
    esac


    case $(get-dialog-answer) in 

        "Disable Persistence" ) delete-from-values "vdi.spec.userdataSpec" ;;

        "Disk Quota"          ) cancel_label="Cancel" do-dialog --inputbox "How much storage should be allocated to each user?" 0 0 \
                                    "$(read-from-values "vdi.spec.userdataSpec.resources.requests.storage")" ;
                                if [[ ${?} == 0 ]] && [[ "$(get-dialog-answer)" != "" ]] ; then
                                    write-to-values "vdi.spec.userdataSpec.resources.requests.storage" "$(get-dialog-answer)"
                                fi ;;
        "Storage Class"       ) cancel_label="Cancel" do-dialog --inputbox "Which StorageClass should be used to provision volumes?\n(Leave empty to use local directories)" 0 0 \
                                    "$(read-from-values "vdi.spec.userdataSpec.storageClassName")" ;
                                if [[ ${?} == 0 ]] && [[ "$(get-dialog-answer)" != "" ]] ; then
                                    write-to-values "vdi.spec.userdataSpec.storageClassName" "$(get-dialog-answer)"
                                fi ;;

        "AccessMode"          ) cancel_label="Cancel" do-dialog --inputbox "Set the AccessMode for volume claims" 0 0 \
                                    "$(read-from-values "vdi.spec.userdataSpec.accessModes" | sed 's/\[//' | sed 's/\]//')"
                                if [[ ${?} == 0 ]] && [[ "$(get-dialog-answer)" != "" ]] ; then
                                    delete-from-values "vdi.spec.userdataSpec.accessModes"
                                    write-to-values "vdi.spec.userdataSpec.accessModes[+]" "$(get-dialog-answer)"
                                fi ;;
    esac

    userdata-menu
}

function misc-menu() {
    do-dialog --extra-button --extra-label "Back" \
        --menu "Miscellaneous Options" 0 0 0 \
        "Token TTL" "The access token duration for the UI" \
        "Audit Log" "Configure the api access audit log" \
        "Replicas" "Configure the number of app pods running" \
        "Session Length" "Configure the maximum desktop session length" \
        "Anonymous" "Allow anonymous users to use kVDI"

    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  main-menu ;;
    esac

    case "$(get-dialog-answer)" in

        "Token TTL" ) cancel_label="Cancel" do-dialog --inputbox "How long should issued access tokens be valid?" 0 0 \
                          "$(read-from-values "vdi.spec.auth.tokenDuration")" ;
                      [[ ${?} == 0 ]] \
                          && write-to-values "vdi.spec.auth.tokenDuration" "$(get-dialog-answer)" ;;

        "Audit Log" ) cancel_label="Cancel" do-dialog $([[ "$(read-from-values "vdi.spec.app.auditLog")" == "false" ]] && echo --defaultno) \
                        --yesno "Enable the API audit log?" 0 0 ;
                      [[ ${?} == 0 ]] \
                          && write-to-values "vdi.spec.app.auditLog" true \
                          || write-to-values "vdi.spec.app.auditLog" false ;;

        "Replicas"  ) cancel_label="Cancel" do-dialog --inputbox "How many app server pods should be running?" 0 0 \
                          "$(read-from-values "vdi.spec.app.replicas")" ;
                      [[ ${?} == 0 ]] \
                          && write-to-values "vdi.spec.app.replicas" "$(get-dialog-answer)" ;;

        "Session Length" )  cancel_label="Cancel" do-dialog --inputbox "How long should desktop sessions live for (e.g. 2h)?" 0 0 \
                                "$(read-from-values "vdi.spec.desktops.maxSessionLength")" ;
                            [[ ${?} == 0 ]] && [[ "$(get-dialog-answer)" != "" ]] \
                                && write-to-values "vdi.spec.desktops.maxSessionLength" "$(get-dialog-answer)" ;;

        "Anonymous" ) cancel_label="Cancel" do-dialog $([[ "$(read-from-values "vdi.spec.auth.allowAnonymous")" == "false" ]] && echo --defaultno) \
                        --yesno "Allow unauthenticated users to use kVDI?" 0 0 ;
                      [[ ${?} == 0 ]] \
                          && write-to-values "vdi.spec.auth.allowAnonymous" true \
                          || write-to-values "vdi.spec.auth.allowAnonymous" false ;;
    esac

    # Return to misc menu after setting something. Selecting Back at anytime
    # will take you to the main menu.
    misc-menu
}

function advanced-menu() {
    # Figure out the default editor
    if which vim &> /dev/null ; then
        editor_default="vim"
    else
        editor_default="vi"
    fi

    do-dialog --extra-button --extra-label "Back" \
        --menu "Advanced Options" 0 0 0 \
        "Edit Values" "Edit the kVDI values.yaml directly" \
        "Start K3s" "Pre-start K3s to be able to configure additional resources via the shell" \
        "Add File" "Add raw kubernetes manifests to the k3s installation" \
        "Shell" "Drop to a shell to run arbitrary commands"
    
    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  main-menu ;;
    esac

    case "$(get-dialog-answer)" in

        "Edit Values" ) ${EDITOR:-${editor_default}} "${CHART_DIR}/kvdi/values.yaml" ;;

        "Start K3s"   ) if [[ "${DRY_RUN}" == "false" ]] ; then 
                            install-base
                            start-k3s
                        else
                            dialog --sleep 3 --backtitle "kVDI Architect" --infobox "Cannot start K3s when --dry-run is enabled" 5 50
                        fi ;;

        "Add File"    ) do-dialog --fselect `pwd` 0 100 ;
                        if [[ "${?}" == 0 ]] ; then
                            cp "$(get-dialog-answer)" "${K3S_MANIFEST_DIR}/"
                        fi ;;

        "Shell"       ) clear ;
                        echo "# Launching a shell process" ;
                        echo ;
                        echo "# To interact with a running K3s environment:" ;
                        echo "#     k3s kubectl <args...>" ;
                        echo ;
                        echo "# Useful environment variables:" ;
                        echo "#     VERSION=${VERSION}"
                        echo "#     CHART_DIR=${CHART_DIR}" ;
                        echo "#     K3S_MANIFEST_DIR=${K3S_MANIFEST_DIR}" ;
                        echo
                        ${SHELL:-/bin/sh} ;;
    esac

    # Return to advanced menu after changes. Selecting Back at anytime
    # will take you to the main menu.
    advanced-menu
}

function main-menu() {
    do-dialog --extra-button --extra-label "Install" \
        --menu "Configuration Options" 0 0 0 \
        "Auth" "Authentication configurations" \
        "TLS"  "TLS configurations" \
        "Userdata" "Configure user desktop persistence" \
        "Misc" "Miscellaneous configurations" \
        "Advanced" "Advanced configurations" \
        "Reset" "Reset all configurations to their defaults"
    
    case ${?} in
        1)  clear ;
            echo "Exiting" ;
            exit ;;
        3)  run-install ;;
    esac

    case $(get-dialog-answer) in
        "Auth"     ) set-auth-config ;;
        "TLS"      ) set-tls-config ;;
        "Userdata" ) userdata-menu ;;
        "Misc"     ) misc-menu ;;
        "Advanced" ) advanced-menu ;;
        "Reset"    ) rm -f "${K3S_MANIFEST_DIR}/kvdi-secrets.yaml" ;
                     rm -f "${K3S_MANIFEST_DIR}/kvdi-tls.yaml" ;
                     rm -rf "${CHART_DIR}" ;
                     get-helm-chart ;
                     trap "rm -rf '${CHART_DIR}'" EXIT ;
                     main-menu ;;
    esac

    main-menu
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
    # return immediately if being sourced
    [[ "${BASH_SOURCE[0]}" != "${0}" ]] && return

    # Make sure dependencies are installed
    if ! which docker &> /dev/null ; then
        echo "You must install 'docker' first to use this scri  valuesContent: |-pt."
        exit 1
    fi

    if ! which dialog &> /dev/null ; then
        echo "You must install the 'dialog' package to use this script."
        exit 1
    fi

    # Make sure we are root
    if [[ "$(id -u)" != "0" ]] ; then
        echo "Not running as root, elevating with sudo"
        exec sudo /bin/bash ${0} $*
        exit
    fi

    export VERSION=""
    export DRY_RUN="false"
    export AUTH_METHOD="Local"

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
            v | version )   needs_arg;
                            export VERSION="$OPTARG" ;;
            dry-run )       export DRY_RUN="true" ;;
            ??* )           die "Illegal option --$OPT" ;;  # bad long option
            \? )            exit 2 ;;  # bad short option (error reported via getopts)
        esac
    done

    shift $((OPTIND-1)) # remove parsed options and args from $@ list

    get-yq

    if [[ "${DRY_RUN}" == "true" ]] ; then
        export K3S_MANIFEST_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'kvdi-manifests')
        trap "rm -rf '${K3S_MANIFEST_DIR}'" EXIT
    fi

    # Initialize the manifest dir
    mkdir -p "${K3S_MANIFEST_DIR}"

    # Pull down the helm chart
    get-helm-chart
    trap "rm -rf '${CHART_DIR}'" EXIT

    # Go to the main menu
    main-menu 
}