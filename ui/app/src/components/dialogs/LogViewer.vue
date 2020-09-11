<template>
  <q-dialog ref="dialog" @hide="onDialogHide" full-width>
    <q-card>
      <q-card-section>
        <div class="text-h6 q-mb-md">kvdi-proxy logs</div>
      </q-card-section>
      <q-card-section>
        <pre>{{logData}}</pre>
      </q-card-section>
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
      logData: ''
    }
  },

  beforeDestroy () {
    if (this.socket) {
      this.socket.close()
    }
  },

  async mounted () {
    this.urls = new DesktopAddressGetter(this.$userStore, this.namespace, this.name)
    try {
      const res = await this.$axios.get(this.urls.logsURL('kvdi-proxy'))
      this.logData = res.data
    } catch (err) {
      this.$root.$emit('notify-error', err)
      this.hide()
    }
  },

  methods: {
    show () {
      this.$refs.dialog.show()
    },

    hide () {
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
