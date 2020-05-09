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

<script>
import jsyaml from 'js-yaml'

const boilerplate = `apiVersion: kvdi.io/v1alpha1
kind: DesktopTemplate
metadata:
  name: my-new-template
spec:
  image: myrepo/my-image:latest
  config: {}
  tags: {}
`

export default {
  name: 'TemplateEditor',
  components: { editor: require('vue2-ace-editor') },
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
        await this.$axios.post('/api/templates', payload)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Created new DesktopTemplate '${payload.metadata.name}'`
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async doUpdate () {
      const payload = this.getPayload()
      try {
        await this.$axios.put(`/api/templates/${payload.metadata.name}`, payload)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Successfully updated DesktopTemplate '${payload.metadata.name}'`
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
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
