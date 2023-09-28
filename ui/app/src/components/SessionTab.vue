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
  <q-btn-dropdown
    :unelevated="!active"
    :outline="active"
    :flat="!active"
    dense
    auto-close stretch split
    @click="onConnect"
  >
    <template v-slot:label>
      <div>
        <div class="row justify-around items-center no-wrap">
          <q-icon name="laptop" />
        </div>
        <div class="row items-center no-wrap">
          {{ name }}
        </div>
      </div>
    </template>

    <q-list>
      <q-item clickable @click="onLogs">
        <q-item-section>Logs</q-item-section>
      </q-item>
      <q-separator />
      <q-item clickable @click="onDisconnect">
        <q-item-section>Disconnect</q-item-section>
      </q-item>
    </q-list>
  </q-btn-dropdown>
</template>

<script lang="ts">
import { useDesktopSessions } from 'src/stores/desktop'
import LogViewerDialog from '../components/dialogs/LogViewer.vue'
import { useConfigStore } from 'src/stores/config'

export default {
  name: 'SessionTab',

  setup: () => {
     return {configStore: useConfigStore(), desktopSessions: useDesktopSessions()}
  },

  props: {
    name: {
      type: String,
      required: true
    },

    namespace: {
      type: String,
      required: true
    },

    active: {
      type: Boolean,
      required: false,
      default: false
    }
  },

  methods: {
    onConnect () {
      console.log(`Setting active session to ${this.namespace}/${this.name}`)
      this.desktopSessions.setActiveSession(this)
      if (this.$router.currentRoute.value.name !== 'control') {
        this.configStore.emitter.emit('set-control')
        this.$router.push('control')
      }
    },
    onLogs () {
      this.$q.dialog({
        component: LogViewerDialog,
        parent: this,
        name: this.name,
        namespace: this.namespace
      }).onOk(() => {
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },
    onDisconnect () {
      this.desktopSessions.deleteSession(this)
    }
  }
}
</script>
