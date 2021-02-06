<template>
  <div>
    <p class="text-h4">Server settings</p>
    <div class="q-px-xl q-mx-xl">
      <q-markdown no-line-numbers :src="serverConfig" />
    </div>
  </div>
</template>

<script>
export default {
  name: 'VDIConfigPanel',
  computed: {
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
