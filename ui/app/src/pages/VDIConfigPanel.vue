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
  <div>
    <p class="text-h4">Server settings</p>
    <div class="q-px-xl q-mx-xl">
      <q-markdown no-line-numbers :src="serverConfig" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { useConfigStore } from '../stores/config'
export default defineComponent({
  name: 'VDIConfigPanel',

  setup (){

    return { configStore: useConfigStore()}
  },
  computed: {
    serverConfig () {
      const cfg = this.configStore.serverConfig
      return `
\`\`\`js
${JSON.stringify(cfg, undefined, 4)}
\`\`\`
`
    }
  }
})
</script>
../stores/config