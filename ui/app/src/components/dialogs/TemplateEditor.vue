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
  <q-dialog ref="dialog" @hide="onDialogHide" persistent transition-show="scale" transition-hide="scale">
     <q-card style="width: 700px; max-width: 80vw;">

       <q-card-section>
         <div v-if="isExisting" class="text-h6">{{ existing.metadata.name }}</div>
         <div v-if="!isExisting" class="text-h5">New Template</div>
         <q-separator />
       </q-card-section>

        <q-card-section class="q-pt-none">
          <editor v-model="contents" @init="editorInit" lang="yaml" theme="chrome" height="30vh"/>
        </q-card-section>

        <q-card-actions align="right" class="bg-white text-teal">
          <q-btn flat label="Exit" v-close-popup @click="onCancelClick" />
          <q-btn flat label="Save" v-close-popup @click="onOKClick" />
        </q-card-actions>
      </q-card>
  </q-dialog>
</template>

<script lang="ts">
import jsyaml from 'js-yaml'

const boilerplate = `apiVersion: desktops.kvdi.io/v1
kind: Template
metadata:
  name: my-new-template
spec:
  desktop:
    image: myrepo/my-image:latest
  tags: {}
`

export default {
  name: 'TemplateEditor',
  components: { editor: await import('vue3-ace-editor') },
  props: {
    existing: {
      type: Object
    }
  },
  data () {
    return {
      contents: ''
    }
  },
  computed: {
    isExisting () { return this.existing !== undefined }
  },
  methods: {

    editorInit () {
      require('brace/ext/language_tools')
      require('brace/mode/html')
      require('brace/mode/yaml')
      require('brace/mode/less')
      require('brace/theme/chrome')
      require('brace/snippets/yaml')
    },

    getPayload () {
      return jsyaml.safeLoad(this.contents)
    },

    async doCreate () {
      const payload = this.getPayload()
      try {
        await this.configStore.axios.post('/api/templates', payload)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Created new DesktopTemplate '${payload.metadata.name}'`
        })
      } catch (err) {
        this.configStore.emitter.emit('notify-error', err)
      }
    },

    async doUpdate () {
      const payload = this.getPayload()
      try {
        await this.configStore.axios.put(`/api/templates/${payload.metadata.name}`, payload)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Successfully updated DesktopTemplate '${payload.metadata.name}'`
        })
      } catch (err) {
        this.configStore.emitter.emit('notify-error', err)
      }
    },

    show () {
      this.$refs.dialog.show()
    },

    hide () {
      this.$refs.dialog.hide()
    },

    onDialogHide () {
      this.$emit('hide')
    },

    async onOKClick () {
      if (this.isExisting) {
        await this.doUpdate()
      } else {
        await this.doCreate()
      }
      this.$emit('ok')
      this.hide()
    },

    onCancelClick () {
      this.hide()
    }

  },

  async mounted () {
    await this.$nextTick()
    if (!this.isExisting) {
      this.contents = boilerplate
      return
    }
    this.contents = jsyaml.safeDump(this.existing, { sortKeys: true, noArrayIndent: true })
  }
}
</script>

<style scoped>

</style>
