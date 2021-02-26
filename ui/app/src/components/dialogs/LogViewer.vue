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
    <q-card style="width: 1400px; max-width: 80vw;">

      <!-- Header -->
      <q-card-section>
        <div class="text-h6 q-mb-md">kvdi-proxy logs</div>
      </q-card-section>

      <!-- Log content -->
      <q-card-section style="max-height: 30vh" class="scroll" id="logs">
        <pre>{{logData}}</pre>
      </q-card-section>

      <q-separator />

      <!-- Close dialog  -->
      <q-card-actions align="right">
        <q-btn flat label="Download" color="primary" @click="onDownload" />
        <q-btn flat :label="toggleLabel" color="primary" @click="onToggle" />
        <q-btn flat label="Close" color="primary" v-close-popup />
      </q-card-actions>

    </q-card>
  </q-dialog>
</template>

<script>
import DesktopAddressGetter from 'src/lib/addresses.js'

export default {
  name: 'LogViewerDialog',

  props: {
    namespace: {
      type: String,
      required: true
    },
    name: {
      type: String,
      required: true
    }
  },

  data () {
    return {
      follow: false,
      socket: null,
      urls: null,
      paused: false,
      logData: '',
      buffer: []
    }
  },

  beforeDestroy () {
    if (this.socket) {
      this.socket.close()
    }
  },

  async mounted () {
    this.urls = new DesktopAddressGetter(this.$userStore, this.namespace, this.name)
    this.streamLogData()
  },

  computed: {
    toggleLabel () {
      if (this.paused) {
        return 'Resume'
      }
      return 'Pause'
    }
  },

  methods: {

    onToggle () {
      this.paused = !this.paused
      if (!this.paused) {
        if (this.buffer.length > 0) {
          this.buffer.forEach((msg) => {
            this.logData = this.logData + msg
          })
          this.buffer = []
        }
      }
    },

    onDownload () {
      const localname = `${this.name}_logs.txt`
      const fileURL = window.URL.createObjectURL(new Blob([this.logData]))
      const fileLink = document.createElement('a')
      fileLink.href = fileURL
      fileLink.setAttribute('download', localname)
      document.body.appendChild(fileLink)
      fileLink.click()
    },

    streamLogData (retry) {
      if (this.socket) {
        return
      }
      this.socket = new WebSocket(this.urls.logsFollowURL('kvdi-proxy'))
      this.socket.addEventListener('message', (ev) => {
        if (ev.data.replace(/\s/g, '') === '') {
          return
        }
        if (this.paused) {
          this.buffer.push(ev.data)
          return
        }
        this.logData = this.logData + ev.data
        const logDiv = document.getElementById('logs')
        if (logDiv) {
          logDiv.scrollTop = logDiv.scrollHeight
        }
      })
      this.socket.addEventListener('close', (ev) => {
        if (!ev.wasClean && ev.code === 1006 && !retry) {
          this.$userStore.dispatch('refreshToken')
            .then(() => {
              this.socket = null
              this.streamLogData(true)
            })
            .catch((err) => {
              throw err
            })
          return
        }
        this.socket = null
      })
    },

    show () {
      this.$refs.dialog.show()
    },

    hide () {
      if (this.socket) {
        this.socket.close()
        this.socket = null
      }
      this.$refs.dialog.hide()
    },

    onDialogHide () {
      this.$emit('hide')
    },

    onOKClick () {
      this.$emit('ok')
      this.hide()
    },

    onCancelClick () {
      this.hide()
    }
  }
}
</script>
