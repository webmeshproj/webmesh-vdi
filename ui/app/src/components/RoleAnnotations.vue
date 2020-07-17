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
  <div v-if="isUsingLocalAuth" class="text-caption">
    Annotations are not used for local authentication.
  </div>
</div>
</template>

<script>
const LDAPGroupAnnotation = 'kvdi.io/ldap-groups'

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
      ldapGroupSelection: []
    }
  },
  computed: {
    isUsingLDAP () {
      return this.$configStore.getters.authMethod === 'ldap'
    },
    isUsingLocalAuth () {
      return this.$configStore.getters.authMethod === 'local'
    },
    configuredLdapGroups () {
      const ldapGroups = []
      if (this.annotations !== undefined) {
        if (this.annotations[LDAPGroupAnnotation] !== undefined) {
          const val = this.annotations[LDAPGroupAnnotation]
          val.split(';').forEach((group) => {
            ldapGroups.push(group)
          })
        }
      }
      return ldapGroups
    }
  },
  methods: {
    reset () {
      this.ldapGroupSelection = this.configuredLdapGroups
    },
    currentAnnotations () {
      if (this.isUsingLDAP) {
        if (this.ldapGroupSelection.length > 0) {
          return {
            'kvdi.io/ldap-groups': this.ldapGroupSelection.join(';')
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
