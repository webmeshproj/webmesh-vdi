#!/bin/bash

set -e
set -o pipefail

echo "** Writing CRDs to deploy/charts/kvdi/crds"
rm -rf deploy/charts/kvdi/crds && mkdir -p deploy/charts/kvdi/crds
cp -r config/crd/bases/* deploy/charts/kvdi/crds/

echo "** Writing Roles to deploy/charts/kvdi/templates/roles.yaml"
cat deploy/bundle.yaml | \
    yq -y '. | select(.kind | test("^Role$"))' | \
    sed '/kvdi-system$/a \ \ labels:\n    {{- include "kvdi.labels" . | nindent 4 }}' | \
    sed '/kvdi-system/d' | \
    sed '0,/kvdi/s/kvdi/{{ include "kvdi.fullname" . }}/g' | \
    sed 's/default/{{ include "kvdi.fullname" . }}-manager/g' \
    > deploy/charts/kvdi/templates/roles.yaml

echo "** Writing RoleBindings to deploy/charts/kvdi/templates/role_bindings.yaml"
cat deploy/bundle.yaml | \
    yq -y '. | select(.kind | test("^RoleBinding$"))' | \
    sed '/kvdi-system/d' | \
    sed 's/kvdi/{{ include "kvdi.fullname" . }}/g' | \
    sed 's/default/{{ include "kvdi.serviceAccountName" . }}/g' | \
    sed '/rolebinding$/a \ \ labels:\n    {{- include "kvdi.labels" . | nindent 4 }}' \
    > deploy/charts/kvdi/templates/role_bindings.yaml

echo "** Writing ClusterRoles to deploy/charts/kvdi/templates/cluster_roles.yaml"
cat deploy/bundle.yaml | \
    yq -y '. | select(.kind | test("^ClusterRole$"))' | \
    sed 's/kvdi-system/{{ .Release.Namespace }}/g' | \
    sed '0,/kvdi/s/kvdi/{{ include "kvdi.fullname" . }}/' | \
    sed '/proxy-role$/a \ \ labels:\n    {{- include "kvdi.labels" . | nindent 4 }}' | \
    sed '/metrics-reader$/a \ \ labels:\n    {{- include "kvdi.labels" . | nindent 4 }}' | \
    sed '/manager-role$/a \ \ labels:\n    {{- include "kvdi.labels" . | nindent 4 }}' | \
    sed 's/kvdi-metrics/{{ include "kvdi.fullname" . }}/g' | \
    sed 's/kvdi-proxy/{{ include "kvdi.fullname" . }}-proxy/g' | \
    sed '/creationTimestamp: null/d' \
    > deploy/charts/kvdi/templates/cluster_roles.yaml

echo "** Writing ClusterRoleBindings to deploy/charts/kvdi/templates/cluster_role_bindings.yaml"
cat deploy/bundle.yaml | \
    yq -y '. | select(.kind | test("^ClusterRoleBinding$"))' | \
    sed 's/kvdi-system/{{ .Release.Namespace }}/g' | \
    sed 's/kvdi/{{ include "kvdi.fullname" . }}/g' | \
    sed 's/default/{{ include "kvdi.serviceAccountName" . }}/g' | \
    sed '/rolebinding$/a \ \ labels:\n    {{- include "kvdi.labels" . | nindent 4 }}' \
    > deploy/charts/kvdi/templates/cluster_role_bindings.yaml

echo "** Writing ConfigMaps to deploy/charts/kvdi/templates/configmaps.yaml"
cat deploy/bundle.yaml | \
    yq -y '. | select(.kind | test("^ConfigMap$"))' | \
    sed '/kvdi-system$/a \ \ labels:\n    {{- include "kvdi.labels" . | nindent 4 }}' | \
    sed '/kvdi-system/d' | \
    sed 's/kvdi-/{{ include "kvdi.fullname" . }}-/' \
    > deploy/charts/kvdi/templates/configmaps.yaml

echo "** Writing Services to deploy/charts/kvdi/templates/services.yaml"
cat deploy/bundle.yaml | \
    yq -y '. | select(.kind | test("^Service$"))' | \
    sed 's/control-plane: controller-manager/{{- include "kvdi.labels" . | nindent 4 }}/g' | \
    sed '/kvdi-system/d' | \
    sed 's/kvdi-controller/{{ include "kvdi.fullname" . }}/' \
    > deploy/charts/kvdi/templates/services.yaml


echo "** Writing Deployments to deploy/charts/kvdi/templates/deployments.yaml"
cat deploy/bundle.yaml | \
    yq -y '. | select(.kind | test("^Deployment$"))' | \
    sed '/kvdi-system/d' | \
    sed '0,/control-plane: controller-manager/s/control-plane: controller-manager/{{- include "kvdi.labels" . | nindent 4 }}/g' | \
    sed '0,/control-plane: controller-manager/s/control-plane: controller-manager/{{- include "kvdi.selectorLabels" . | nindent 6 }}/g' | \
    sed '0,/control-plane: controller-manager/s/control-plane: controller-manager/{{- include "kvdi.labels" . | nindent 8 }}/g' | \
    sed 's/kvdi-controller/{{ include "kvdi.fullname" . }}/' | \
    sed 's/replicas: 1/replicas: {{ .Values.manager.replicaCount }}/' | \
    sed 's,image: gcr.io/kubebuilder/kube-rbac-proxy:v0.5.0,image: {{ .Values.rbac.proxy.repository }}:{{ .Values.rbac.proxy.tag }},' | \
    sed "s,image: ghcr.io/kvdi/manager:.*$,image: {{ .Values.manager.image.repository }}:{{ default .Values.manager.image.tag .Chart.AppVersion }}\n          imagePullPolicy: {{ .Values.manager.image.pullPolicy }}," | \
    sed 's/}}latest/}}/' | \
    sed -n '1h;1!H;${g;s/resources.*PeriodSeconds: 10//;p;}' | sed '$ d' \
    > deploy/charts/kvdi/templates/deployments.yaml

cat << EOM >> deploy/charts/kvdi/templates/deployments.yaml
          env:
            - name: OPERATOR_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: OPERATOR_NAME
              value: "kvdi"
          {{- with .Values.manager.resources }}
          resources:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- with .Values.manager.securityContext }}
          securityContext:
            {{- toYaml . | nindent 12 }}
          {{- end }}
      {{- with .Values.manager.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
    {{- with .Values.manager.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
    {{- end }}
    {{- with .Values.manager.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
    {{- end }}
      securityContext:
        runAsUser: 65532
      terminationGracePeriodSeconds: 10
      serviceAccountName: {{ include "kvdi.serviceAccountName" . }}
EOM