<template>
  <q-dialog ref="dialog" @hide="onDialogHide">
    <q-card style="min-width: 350px">
      <q-card-section>
        <div class="text-h6">New User</div>
      </q-card-section>

      <UserEditor editorFunction="create" />

    </q-card>
  </q-dialog>
</template>

<script>
import UserEditor from 'components/UserEditor'

export default {
  name: 'NewUserDialog',
  components: { UserEditor },
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
