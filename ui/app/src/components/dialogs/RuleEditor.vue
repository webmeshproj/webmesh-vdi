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
  <q-dialog ref="dialog" @hide="onDialogHide">
    <q-card>
      <!-- Verbs -->
      <q-card-section>
        <div class="text-h6">Actions</div>
        <q-checkbox v-for="opt in verbOptions" v-model="verbSelections[opt.name]" :key="opt.name" :label="opt.display" :color="opt.color" />
      </q-card-section>
      <!-- Resources -->
      <q-card-section>
        <div class="text-h6">Resources</div>
        <q-checkbox v-for="opt in resourceOptions" v-model="resourceSelections[opt.name]" :key="opt.name" :label="opt.display" :color="opt.color" />
      </q-card-section>
      <!-- Resource Patterns -->
      <q-card-section>
        <div class="text-h6">Patterns</div>
        <PatternSelector ref="patterns" />
      </q-card-section>
      <!-- Namespaces -->
      <q-card-section>
        <div class="text-h6">Namespaces</div>
        <NamespaceSelector ref="namespaces" :showAllOption="true" :multiSelect="true" />
      </q-card-section>
      <!-- Actions -->
      <q-card-section>
        <q-btn flat label="Cancel" v-close-popup @click="onCancelClick" />
        <q-btn flat label="Save Rule" v-close-popup @click="onOKClick" />
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script lang="ts">
import PatternSelector from '../inputs/PatternSelector.vue'
import NamespaceSelector from '../inputs/NamespaceSelector.vue'

export default {
  name: 'RuleEditor',

  components: { PatternSelector, NamespaceSelector },

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
        { name: 'create', color: 'green', display: 'Create' },
        { name: 'read', color: 'blue', display: 'Read' },
        { name: 'update', color: 'orange', display: 'Update' },
        { name: 'delete', color: 'red', display: 'Delete' },
        { name: 'use', color: 'teal', display: 'Use' },
        { name: 'launch', color: 'purple', display: 'Launch' }
      ],
      resourceOptions: [
        { name: 'users', color: 'green', display: 'Users' },
        { name: 'roles', color: 'blue', display: 'Roles' },
        { name: 'templates', color: 'teal', display: 'Templates' },
        { name: 'serviceaccounts', color: 'purple', display: 'ServiceAccounts' }
      ],
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
        templates: false,
        serviceaccounts: false
      },
      resourcePatternSelections: []
    }
  },

  methods: {

    buildPayload () {
      return {
        verbs: this.getVerbs(),
        resources: this.getResources(),
        resourcePatterns: this.$refs.patterns.selection,
        namespaces: this.$refs.namespaces.selection
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

  async mounted () {
    await this.$nextTick()
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
            templates: true,
            serviceaccounts: true
          }
          return
        }
        this.resourceSelections[resource] = true
      })
    }
    if (this.resourcePatterns !== undefined) {
      this.$refs.patterns.selection = this.resourcePatterns
    } else {
      this.$refs.patterns.selection = []
    }
    if (this.namespaces !== undefined) {
      this.$refs.namespaces.selection = this.namespaces
    } else {
      this.$refs.namespaces.selection = []
    }
  }
}
</script>
