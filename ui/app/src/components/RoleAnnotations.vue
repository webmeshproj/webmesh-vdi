<!--
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
-->

<template>
<div>
  <div v-if="isUsingLDAP">
    <q-select
      label="LDAP Groups"
      v-model="ldapGroupSelection"
      use-input
      use-chips
      bottom-slots
      multiple
      :clearable="editable"
      dense
      hide-dropdown-icon
      input-debounce="0"
      new-value-mode="add-unique"
      :disabled="!editable"
    />
  </div>
  <div v-if="isUsingOIDC">
    <q-select
      label="OpenID Groups"
      v-model="oidcGroupSelection"
      use-input
      use-chips
      bottom-slots
      multiple
      :clearable="editable"
      dense
      hide-dropdown-icon
      input-debounce="0"
      new-value-mode="add-unique"
      :disabled="!editable"
    />
  </div>
  <div v-if="isUsingLocalAuth" class="text-caption">
    Annotations are not used for local authentication.
  </div>
</div>
</template>

<script lang="ts">
const LDAPGroupAnnotation = 'kvdi.io/ldap-groups'
const OIDCGroupAnnotation = 'kvdi.io/oidc-groups'

export default {
  name: 'RoleAnnotations',
  props: {
    roleIdx: { type: Number },
    roleName: { type: String },
    annotations: { type: Object },
    editable: {
      type: Boolean,
      required: false,
      default: false
    }
  },
  data () {
    return {
      ldapGroupSelection: [],
      oidcGroupSelection: []
    }
  },
  computed: {
    isUsingOIDC () {
      return this.$configStore.getters.authMethod === 'oidc'
    },
    isUsingLDAP () {
      return this.$configStore.getters.authMethod === 'ldap'
    },
    isUsingLocalAuth () {
      return this.$configStore.getters.authMethod === 'local'
    },
    configuredLdapGroups () {
      const ldapGroups: any[] = []
      if (this.annotations !== undefined) {
        if (this.annotations[LDAPGroupAnnotation] !== undefined) {
          const val = this.annotations[LDAPGroupAnnotation]
          val.split(';').forEach((group) => {
            ldapGroups.push(group)
          })
        }
      }
      return ldapGroups
    },
    configuredOidcGroups () {
      const oidcGroups: any[] = []
      if (this.annotations !== undefined) {
        if (this.annotations[OIDCGroupAnnotation] !== undefined) {
          const val = this.annotations[OIDCGroupAnnotation]
          val.split(';').forEach((group) => {
            oidcGroups.push(group)
          })
        }
      }
      return oidcGroups
    }
  },
  methods: {
    reset () {
      if (this.isUsingLDAP) {
        this.ldapGroupSelection = this.configuredLdapGroups
      }
      if (this.isUsingOIDC) {
        this.oidcGroupSelection = this.configuredOidcGroups
      }
    },
    currentAnnotations () {
      if (this.isUsingLDAP) {
        if (this.ldapGroupSelection.length > 0) {
          return {
            'kvdi.io/ldap-groups': this.ldapGroupSelection.join(';')
          }
        }
      }
      if (this.isUsingOIDC) {
        if (this.oidcGroupSelection.length > 0) {
          return {
            'kvdi.io/oidc-groups': this.oidcGroupSelection.join(';')
          }
        }
      }
      return {}
    }
  },
  mounted () {
    this.$nextTick().then(() => {
      this.reset()
    })
  }
}
</script>

<style scoped>

</style>
