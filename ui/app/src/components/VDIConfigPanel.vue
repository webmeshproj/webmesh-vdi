<template>
  <div class="q-pa-md" stretch>
    <q-toggle
      label="Allow multiple sessions when using persistence"
      v-model="readWriteMany"
    />
  </div>
</template>

<script>
export default {
  name: 'VDIConfigPanel',
  computed: {
    readWriteMany: {
      get () {
        return this.$configStore.getters.localConfig.readWriteMany
      },
      set (val) {
        this.$configStore.dispatch('setReadWriteMany', val)
      }
    }
  }
}
</script>

<style>

</style>
