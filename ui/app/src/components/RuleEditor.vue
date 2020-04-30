<template>
  <q-dialog ref="dialog" @hide="onDialogHide">
    <q-card>
      <q-card-section>
        <div class="text-h6">Actions</div>
        <q-checkbox v-for="opt in verbOptions" v-model="verbSelections[opt.name]" :key="opt.name" :label="toTitleCase(opt.name)" :color="opt.color" />
      </q-card-section>
      <q-card-section>
        <div class="text-h6">Resources</div>
        <q-checkbox v-for="opt in resourceOptions" v-model="resourceSelections[opt.name]" :key="opt.name" :label="toTitleCase(opt.name)" :color="opt.color" />
      </q-card-section>
      <q-card-section>
        <div class="text-h6">Patterns</div>
        <q-select
          label="Resource patterns (Use '.*' All)"
          v-model="resourcePatternSelections"
          use-input
          use-chips
          bottom-slots
          multiple
          clearable
          dense
          hide-dropdown-icon
          input-debounce="0"
          new-value-mode="add-unique"
          :v-close-popup="false"
          hint="Press Enter to add patterns"
        />
      </q-card-section>
      <q-card-section>
        <div class="text-h6">Namespaces</div>
        <q-select
          v-model="namespaceSelections"
          use-input
          use-chips
          multiple
          clearable
          dense
          :loading="loading"
          transition-show="scale"
          transition-hide="scale"
          :virtual-scroll-slice-size="5"
          @filter="namespaceFilterFn"
          label="Namespaces"
          :options="namespaceOptions"
        >
          <template v-slot:no-option>
            <q-item>
              <q-item-section class="text-grey">
                No results
              </q-item-section>
            </q-item>
          </template>
        </q-select>
      </q-card-section>
      <q-card-section>
        <q-btn flat label="Cancel" v-close-popup @click="onCancelClick" />
        <q-btn flat label="Save Rule" v-close-popup @click="onOKClick" />
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script>

export default {
  name: 'RuleEditor',

  props: {
    verbs: {
      type: Array
    },
    resources: {
      type: Array
    },
    resourcePatterns: {
      type: Array
    },
    namespaces: {
      type: Array
    }
  },

  data () {
    return {
      loading: false,
      verbOptions: [
        { name: 'create', color: 'green' },
        { name: 'read', color: 'blue' },
        { name: 'update', color: 'orange' },
        { name: 'delete', color: 'red' },
        { name: 'use', color: 'teal' },
        { name: 'launch', color: 'purple' }
      ],
      resourceOptions: [
        { name: 'users', color: 'green' },
        { name: 'roles', color: 'blue' },
        { name: 'templates', color: 'teal' }
      ],
      namespaceOptions: [],
      verbSelections: {
        create: false,
        read: false,
        update: false,
        delete: false,
        use: false,
        launch: false
      },
      resourceSelections: {
        users: false,
        roles: false,
        templates: false
      },
      resourcePatternSelections: [],
      namespaceSelections: []
    }
  },

  methods: {

    async namespaceFilterFn (val, update) {
      this.loading = true
      try {
        const res = await this.$axios.get('/api/namespaces')
        if (res.data.length === 1) {
          this.namespaces = res.data[0]
        }
        if (val === '') {
          update(() => {
            this.namespaceOptions = res.data.unshift('*')
          })
        }
        update(() => {
          const needle = val.toLowerCase()
          this.namespaceOptions = res.data.filter(v => v.toLowerCase().indexOf(needle) > -1)
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
      this.loading = false
    },

    toTitleCase (str) {
      return str.replace(/\w\S*/g, (txt) => {
        return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase()
      })
    },

    buildPayload () {
      return {
        verbs: this.getVerbs(),
        resources: this.getResources(),
        resourcePatterns: this.resourcePatternSelections,
        namespaces: this.namespaceSelections
      }
    },

    getVerbs () {
      let verbs = []
      Object.keys(this.verbSelections).forEach((verb) => {
        if (this.verbSelections[verb]) {
          verbs.push(verb)
        }
      })
      if (verbs.length === Object.keys(this.verbSelections).length) {
        verbs = ['*']
      }
      return verbs
    },

    getResources () {
      let resources = []
      Object.keys(this.resourceSelections).forEach((resource) => {
        if (this.resourceSelections[resource]) {
          resources.push(resource)
        }
      })
      if (resources.length === Object.keys(this.resourceSelections).length) {
        resources = ['*']
      }
      return resources
    },

    // following method is REQUIRED
    // (don't change its name --> "show")
    show () {
      this.$refs.dialog.show()
    },

    // following method is REQUIRED
    // (don't change its name --> "hide")
    hide () {
      this.$refs.dialog.hide()
    },

    onDialogHide () {
      // required to be emitted
      // when QDialog emits "hide" event
      this.$emit('hide')
    },

    onOKClick () {
      // on OK, it is REQUIRED to
      // emit "ok" event (with optional payload)
      // before hiding the QDialog
      this.$emit('ok', this.buildPayload())
      // or with payload: this.$emit('ok', { ... })

      // then hiding dialog
      this.hide()
    },

    onCancelClick () {
      // we just need to hide dialog
      this.hide()
    }
  },

  mounted () {
    if (this.verbs !== undefined) {
      this.verbs.forEach((verb) => {
        if (verb === '*') {
          this.verbSelections = {
            create: true,
            read: true,
            update: true,
            delete: true,
            use: true,
            launch: true
          }
          return
        }
        this.verbSelections[verb] = true
      })
    }
    if (this.resources !== undefined) {
      this.resources.forEach((resource) => {
        if (resource === '*') {
          this.resourceSelections = {
            users: true,
            roles: true,
            templates: true
          }
          return
        }
        this.resourceSelections[resource] = true
      })
    }
    if (this.resourcePatterns !== undefined) {
      this.resourcePatternSelections = this.resourcePatterns
    } else {
      this.resourcePatternSelections = []
    }
    if (this.namespaces !== undefined) {
      this.namespaceSelections = this.namespaces
    } else {
      this.namespaceSelections = []
    }
  }
}
</script>
