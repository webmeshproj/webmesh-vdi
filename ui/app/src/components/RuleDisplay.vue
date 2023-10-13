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
  <!-- Button class when editable causes the div to wiggle -->
  <div :class="buttonClass">
    <!-- Delete rule button -->
    <q-btn v-if="editable" round dense flat icon="remove"  size="sm" color="red" @click="onDeleteRule">
      <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Delete Rule</q-tooltip>
    </q-btn>
    <!-- Edit rule button -->
    <q-btn rounded flat dense @click="onEditRule">
      <!-- Breadcrumbs give a nice effect for displaying rule interactions -->
      <q-breadcrumbs :class="`text-${color}-10`" :active-color="`${color}-10`">
        <template v-slot:separator>
          <q-icon size="1.5em" name="arrow_forward" color="black"/>
        </template>
        <q-breadcrumbs-el to="" :label="`ACTIONS: ${display('verbs')}`" icon="settings_remote" />
        <q-breadcrumbs-el :label="`RESOURCES: ${display('resources')}`" icon="widgets" />
        <q-breadcrumbs-el :label="`MATCHING: ${display('resourcePatterns')}`" icon="label" />
        <q-breadcrumbs-el :label="`IN NAMESPACES: ${display('namespaces')}`" icon="web" />
      </q-breadcrumbs>
      <q-tooltip v-if="editable" anchor="center right" self="center middle">Click to edit this rule</q-tooltip>
    </q-btn>
    <q-space />
  </div>
</template>

<script lang="ts">
import { useConfigStore } from 'src/stores/config'
import RuleEditor from './dialogs/RuleEditor.vue'
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'RuleDisplay',
  props: {
    ruleIdx: { type: Number },
    roleIdx: { type: Number },
    roleName: { type: String },
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
    },
    editable: {
      type: Boolean,
      required: false,
      default: false
    }
  },
  setup () {
    return {
      editorOpen: false,
      configStore: useConfigStore()
    }
  },
  methods: {
    onDeleteRule () {
      if (this.roleName) {
        this.configStore.emitter.emit(this.roleName, {
          roleIdx: this.roleIdx,
          ruleIdx: this.ruleIdx,
          deleteRule: true
        })
      }
    },
    onEditRule () {
      if (!this.editable) { return }
      this.editorOpen = true
      this.$q.dialog({
        component: RuleEditor,
       componentProps: {
        parent: this,
        verbs: this.verbs,
        resources: this.resources,
        resourcePatterns: this.resourcePatterns,
        namespaces: this.namespaces
       }
      }).onOk((payload) => {
        if (this.roleName) {
          this.configStore.emitter.emit(this.roleName, {
            roleIdx: this.roleIdx,
            ruleIdx: this.ruleIdx,
            rulePayload: payload
          })
        }
      }).onCancel(() => {
        console.log('Cancelled rule edit')
      }).onDismiss(() => {
        this.editorOpen = false
      })
    },
    display (item) {
      if (this[item] === undefined || this[item].length === 0) {
        return 'NONE'
      }
      let itemAll = false
      this[item].forEach((x) => {
        if (x === '*' || x === '.*') {
          itemAll = true
        }
      })
      if (itemAll) {
        return 'ANY'
      }
      return this[item].join(',')
    }
  },
  computed: {

    buttonClass () {
      if (this.editable && !this.editorOpen) {
        return 'rule-editable'
      }
      return ''
    },

    color () {
      if (this.editable) {
        return 'blue'
      }
      if (this.fullAccess) {
        return 'green'
      }
      if (this.noAccess) {
        return 'red'
      }
      if (this.limitedAccess) {
        return 'orange'
      }
      return 'cyan'
    },

    fullAccess () {
      return this.display('verbs') === 'ANY' &&
        this.display('resources') === 'ANY' &&
        this.display('resourcePatterns') === 'ANY' &&
        this.display('namespaces') === 'ANY'
    },

    limitedAccess () {
      return this.display('verbs') === 'NONE' ||
        this.display('resources') === 'NONE' ||
        this.display('resourcePatterns') === 'NONE' ||
        this.display('namespaces') === 'NONE'
    },

    noAccess () {
      return this.display('verbs') === 'NONE' &&
        this.display('resources') === 'NONE' &&
        this.display('resourcePatterns') === 'NONE' &&
        this.display('namespaces') === 'NONE'
    }

  }

})
</script>

<style scoped>
.rule-editable {
  animation: shake 0.82s cubic-bezier(.36,.07,.19,.97) both;
  animation-iteration-count: infinite;
  transform: translate3d(0, 0, 0);
  backface-visibility: hidden;
  perspective: 1000px;
}

@keyframes shake {
  10%, 90% {
    transform: translate3d(-1px, 0, 0)
  }

  20%, 80% {
    transform: translate3d(1px, 0, 0)
  }

  30%, 50%, 70% {
    transform: translate3d(-1px, 0, 0)
  }

  40%, 60% {
    transform: translate3d(1px, 0, 0)
  }
}
</style>
