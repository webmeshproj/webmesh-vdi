<template>
  <q-dialog ref="dialog" seamless position="bottom" @hide="onDialogHide" @mouseleave="onDialogHide">
    <q-card square style="width: 300px">
      <q-card-actions align="left">
         <q-btn flat round icon="fullscreen" color="primary" @click="onFullscreen">
          <q-tooltip anchor="top right" self="top middle">Fullscreen</q-tooltip>
        </q-btn>
        <q-btn v-if="!knownAudioState" flat round icon="volume_off" color="red" @click="doToggleAudio">
          <q-tooltip anchor="top right" self="top middle">Enable audio</q-tooltip>
        </q-btn>
        <q-btn v-if="knownAudioState" flat round icon="volume_up" color="green" @click="doToggleAudio">
          <q-tooltip anchor="top right" self="top middle">Disable audio</q-tooltip>
        </q-btn>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script>
export default {
  name: 'VNCControls',
  props: {
    audioState: {
      type: Boolean
    },
    onToggleAudio: {
      type: Function
    }
  },
  data () {
    return {
      knownAudioState: false
    }
  },
  async mounted () {
    await this.$nextTick()
    this.knownAudioState = this.audioState
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
    },

    onFullscreen () {
      this.$q.fullscreen.request()
    },

    doToggleAudio () {
      this.onToggleAudio()
      this.knownAudioState = !this.knownAudioState
    }
  }
}
</script>
