<template>
  <q-dialog ref="dialog" @hide="onDialogHide">
    <q-card>
      <q-card-section>
        <div class="text-h6 q-mb-md">{{name}}</div>
      </q-card-section>
      <q-card-section>
        <pre>{{src}}</pre>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script>
export default {
  name: 'FilePreviewDialog',
  props: {
    src: {
      type: String,
      required: true
    },
    name: {
      type: String,
      required: true
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
