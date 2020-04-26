<template>
  <q-card style="min-width: 350px">

    <q-card-section v-if="editorFunction == 'create'">
      <q-input dense debounce="500" label="Name" v-model="name" :rules="[validateRole]"/>
    </q-card-section>

    <q-card-section class="q-pt-none">

      <q-select
        v-model="grantSelection"
        use-input
        use-chips
        multiple
        clearable
        dense
        :loading="loading"
        transition-show="scale"
        transition-hide="scale"
        :virtual-scroll-slice-size="5"
        popup-content-class="grant-selection"
        @filter="grantsFilterFn"
        label="Grants"
        :options="grants"
        :rules="[val => val.length > 0 || 'You must select at least one grant']"
      >
        <template v-slot:no-option>
          <q-item>
            <q-item-section class="text-grey">
              No results
            </q-item-section>
          </q-item>
        </template>
      </q-select>

      <q-select
        v-model="namespaceSelection"
        use-input
        use-chips
        multiple
        clearable
        dense
        :loading="loading"
        transition-show="scale"
        transition-hide="scale"
        :virtual-scroll-slice-size="5"
        popup-content-class="grant-selection"
        @filter="namespaceFilterFn"
        label="Namespaces (default: All)"
        :options="namespaces"
      >
        <template v-slot:no-option>
          <q-item>
            <q-item-section class="text-grey">
              No results
            </q-item-section>
          </q-item>
        </template>
      </q-select>

      <q-select
        label="Template patterns (default: All)"
        v-model="templatePatterns"
        use-input
        use-chips
        bottom-slots
        multiple
        dense
        :loading="loading"
        hide-dropdown-icon
        input-debounce="0"
        new-value-mode="add"
        hint="Press Enter to add patterns"
      />

    </q-card-section>

    <q-card-actions align="right" class="text-primary">
      <q-btn flat label="Cancel" v-close-popup />
      <q-btn flat :label="submitLabel" v-close-popup @click="submitFunc" />
    </q-card-actions>

  </q-card>

</template>

<script>

export default {
  name: 'RoleEditor',
  props: {
    editorFunction: {
      type: String,
      required: false,
      default: 'create'
    },
    roleToEdit: {
      type: String,
      required: false,
      default: ''
    }
  },
  data () {
    return {
      name: '',
      loading: true,
      grantMap: {},
      grantSelection: [],
      grants: [],
      namespaceSelection: [],
      namespaces: [],
      templatePatterns: []
    }
  },
  computed: {
    submitLabel () {
      if (this.editorFunction === 'create') {
        return 'Create Role'
      }
      return 'Update Role'
    },
    submitFunc () {
      if (this.editorFunction === 'create') {
        return this.addRole
      }
      return this.updateRole
    }
  },
  methods: {

    async validateRole (val) {
      if (!val) {
        return 'Name is required'
      }
      try {
        await this.$axios.get(`/api/roles/${val}`)
        return 'Role already exists'
      } catch (err) {}
    },

    async addRole () {
      let grantValue = 0
      this.grantSelection.forEach((val) => {
        grantValue = grantValue | this.grantMap[val]
      })
      const payload = {
        name: this.name,
        grants: grantValue,
        namespaces: this.namespaceSelection,
        templatePatterns: this.templatePatterns
      }
      console.log(payload)
      try {
        await this.$axios.post('/api/roles', payload)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `New role '${this.name}' created`
        })
        this.$root.$emit('reload-roles')
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async updateRole () {
      let grantValue = 0
      this.grantSelection.forEach((val) => {
        grantValue = grantValue | this.grantMap[val]
      })
      const payload = {
        grants: grantValue,
        namespaces: this.namespaceSelection,
        templatePatterns: this.templatePatterns
      }
      console.log(payload)
      try {
        await this.$axios.put(`/api/roles/${this.roleToEdit}`, payload)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Role '${this.name}' updated succesfully`
        })
        this.$root.$emit('reload-roles')
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async grantsFilterFn (val, update) {
      try {
        const res = await this.$axios.get('/api/grants')
        const grants = []
        Object.keys(res.data).forEach((key) => {
          grants.push(key)
        })
        if (val === '') {
          update(() => {
            this.grants = grants
          })
        }
        update(() => {
          const needle = val.toLowerCase()
          this.grants = grants.filter(v => v.toLowerCase().indexOf(needle) > -1)
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async namespaceFilterFn (val, update) {
      try {
        const res = await this.$axios.get('/api/namespaces')
        if (res.data.length === 1) {
          this.namespaceSelection = res.data[0]
        }
        if (val === '') {
          update(() => {
            this.namespaces = res.data
          })
        }
        update(() => {
          const needle = val.toLowerCase()
          this.namespaces = res.data.filter(v => v.toLowerCase().indexOf(needle) > -1)
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    }

  },

  mounted () {
    this.$nextTick().then(() => {
      this.$axios.get('/api/grants')
        .then((res) => {
          Object.keys(res.data).forEach((key) => {
            this.grantMap[key] = res.data[key]
          })
        })
        .catch((err) => {
          this.$root.$emit('notify-error', err)
        })
      if (this.roleToEdit !== '') {
        this.$axios.get(`/api/roles/${this.roleToEdit}`)
          .then((res) => {
            this.grantSelection = res.data.grants
            this.namespaceSelection = res.data.namespaces
            this.templatePatterns = res.data.templatePatterns
            this.loading = false
          })
          .catch((err) => {
            this.$root.$emit('notify-error', err)
          })
      } else {
        this.loading = false
      }
    })
  }
}
</script>

<style>
.grant-selection {
  max-height: 250px;
  font-size: 12px
}
</style>
