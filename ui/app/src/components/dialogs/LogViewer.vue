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
        <q-btn flat :label="toggleLabel" color="primary" @click="onToggle" />
        <q-btn flat label="Close" color="primary" v-close-popup />
      </q-card-actions>

    </q-card>
  </q-dialog>
</template>

<script>
import { DesktopAddressGetter } from '../../lib/displayManager.js'

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

    streamLogData () {
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
        logDiv.scrollTop = logDiv.scrollHeight
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
