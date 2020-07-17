<template>
  <div>
    <p class="text-h4">Server settings</p>
    <div class="q-px-xl q-mx-xl">
      <q-markdown no-line-numbers :src="serverConfig" />
    </div>
    <p class="text-h4 q-pt-md">Local settings</p>
    <div class="q-px-xl" stretch>
      <q-toggle
        label="Allow multiple sessions when using persistence"
        v-model="readWriteMany"
      />
    </div>
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
    },
    serverConfig () {
      const cfg = this.$configStore.getters.serverConfig
      return `
\`\`\`js
${JSON.stringify(cfg, undefined, 4)}
\`\`\`
`
    }
  }
}
</script>
