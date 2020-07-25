<template>
  <q-page padding>
    <div>
      <q-space />
    </div>
    <div style="float: right">
      <q-btn flat color="primary" icon-right="add" label="New Template" @click="onNewTemplate" />
    </div>

    <div style="clear: right">
      <SkeletonTable v-if="loading"/>

      <q-table
        class="templates-table"
        title="Desktop Templates"
        :data="data"
        :columns="columns"
        row-key="idx"
        v-if="!loading"
        ref="table"
      >

        <!-- No results -->
        <template v-slot:no-data>
          <div class="full-width row flex-left text-secondary q-gutter-md">
            <q-icon size="2em" name="sentiment_dissatisfied" />
            <span style="display:inline-block; margin-top: 19px">
              No DesktopTemplates found
            </span>
            <q-btn flat size="sm" :loading="refreshLoading" color="secondary" @click="refreshData" label="Refresh" />
          </div>
        </template>

        <template v-slot:body="props">
          <q-tr :props="props">

            <q-td key="name" :props="props">
              <strong>{{ props.row.metadata.name }}</strong>
            </q-td>

            <q-td key="image" :props="props">
              <strong>{{ props.row.spec.image }}</strong>
            </q-td>

            <q-td key="sound" :props="props">
              <q-avatar v-if="props.row.spec.config.enableSound" size="27px" font-size="20px" color="green" text-color="white" icon="done" />
            </q-td>

            <q-td key="root" :props="props">
              <q-avatar v-if="props.row.spec.config.allowRoot" size="27px" font-size="20px" color="green" text-color="white" icon="done" />
            </q-td>

            <q-td key="tags" :props="props">
              <div class="tags-wrapper">
                <li class="inline-tags" v-for="tag in tagsToArray(props.row.spec.tags)" :key="tag" dense>
                    <q-chip dense icon="bookmark">{{ tag }}</q-chip>
                </li>
              </div>
            </q-td>

            <q-td key="namespace" :props="props">
              <NamespaceSelector :ref="`ns-${props.row.idx}`" :multiSelect="false" :showAllOption="false" label="Launch Namespace (default)" />
            </q-td>

            <q-td key="useTemplate" :props="props">
              <q-btn round dense flat icon="cast"  size="md" color="blue" @click="onLaunchTemplate(props.row.idx, props.row.metadata.name)">
                <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Launch Template</q-tooltip>
              </q-btn>
              <q-btn round dense flat icon="create"  size="md" color="orange" @click="onEditTemplate(props.row)">
                <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Edit Template</q-tooltip>
              </q-btn>
              <q-btn round dense flat icon="remove_circle"  size="md" color="red" @click="onConfirmDeleteTemplate(props.row.metadata.name)">
                <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Delete Template</q-tooltip>
              </q-btn>
            </q-td>

          </q-tr>
        </template>
      </q-table>
    </div>

  </q-page>
</template>

<script>
import SkeletonTable from 'components/SkeletonTable.vue'
import NamespaceSelector from 'components/inputs/NamespaceSelector.vue'
import TemplateEditor from 'components/dialogs/TemplateEditor.vue'
import ConfirmDelete from 'components/dialogs/ConfirmDelete.vue'

const templateColums = [
  {
    name: 'name',
    required: true,
    label: 'Template',
    align: 'left',
    field: row => row.name,
    format: val => `${val}`,
    sortable: true,
    classes: 'bg-grey-2 ellipsis',
    headerClasses: 'bg-primary text-white'
  },
  {
    name: 'image',
    align: 'left',
    label: 'Image'
  },
  {
    name: 'sound',
    align: 'center',
    label: 'Sound'
  },
  {
    name: 'root',
    align: 'center',
    label: 'Root'
  },
  {
    name: 'tags',
    align: 'center',
    label: 'Tags'
  },
  {
    name: 'namespace',
    align: 'center',
    label: 'Namespace'
  },
  {
    name: 'useTemplate',
    align: 'center'
  }
]

export default {
  name: 'DesktopTemplates',
  components: { SkeletonTable, NamespaceSelector },

  data () {
    return {
      loading: false,
      refreshLoading: false,
      columns: templateColums,
      data: []
    }
  },

  methods: {
    async onNewTemplate () {
      this.$q.dialog({
        component: TemplateEditor,
        parent: this
      }).onOk(async () => {
        await new Promise((resolve, reject) => setTimeout(resolve, 300))
        this.fetchData()
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    onLaunchTemplate (templateIdx, templateName) {
      const ns = this.$refs[`ns-${templateIdx}`].selection
      const payload = { template: templateName }
      // When the attribute comes back as an object, it actually means
      // no selection
      if (typeof ns !== 'object') {
        payload.namespace = ns
      }
      this.doLaunchTemplate(payload)
    },

    async onEditTemplate (template) {
      this.$q.dialog({
        component: TemplateEditor,
        parent: this,
        existing: this.pruneTemplateObject(template)
      }).onOk(async () => {
        await new Promise((resolve, reject) => setTimeout(resolve, 300))
        this.fetchData()
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    onConfirmDeleteTemplate (templateName) {
      this.$q.dialog({
        component: ConfirmDelete,
        parent: this,
        resourceName: templateName
      }).onOk(() => {
        this.doDeleteTemplate(templateName)
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    pruneTemplateObject (template) {
      delete template.idx
      delete template.metadata.creationTimestamp
      delete template.metadata.generation
      delete template.metadata.managedFields
      delete template.metadata.resourceVersion
      delete template.metadata.selfLink
      delete template.metadata.uid
      if (template.metadata.annotations !== undefined && template.metadata.annotations !== null) {
        delete template.metadata.annotations['kubectl.kubernetes.io/last-applied-configuration']
      }
      delete template.status
      return template
    },

    tagsToArray (tagsObj) {
      const tags = []
      if (tagsObj === undefined || tagsObj === null) { return tags }
      Object.keys(tagsObj).forEach((key) => {
        tags.push(`${key}: ${tagsObj[key]}`)
      })
      return tags
    },

    async doLaunchTemplate (payload) {
      try {
        await this.$desktopSessions.dispatch('newSession', payload)
        this.$root.$emit('set-control')
        this.$router.push('control')
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async doDeleteTemplate (templateName) {
      try {
        await this.$axios.delete(`/api/templates/${templateName}`)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Deleted DesktopTemplate '${templateName}'`
        })
        this.fetchData()
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async refreshData () {
      this.data = []
      this.refreshLoading = true
      await new Promise((resolve, reject) => setTimeout(resolve, 500))
      this.fetchData()
      this.refreshLoading = false
    },

    async fetchData () {
      try {
        this.data = []
        const res = await this.$axios.get('/api/templates')
        res.data.forEach((tmpl, idx) => {
          this.data.push({
            idx: idx,
            ...tmpl
          })
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    }
  },

  async mounted () {
    await this.$nextTick()
    this.loading = true
    await new Promise((resolve, reject) => setTimeout(resolve, 500))
    await this.fetchData()
    this.loading = false
  }
}
</script>

<style lang="sass" scoped>
.tags-wrapper
  position: relative
  width: 40vh

.inline-tags
  display: inline

.templates-table

  background-color: $grey-3

  /* height or max-height is important */
  max-height: 500px

  // /* specifying max-width so the example can
  //   highlight the sticky column on any browser window */
  // max-width: 600px

  td:first-child
    /* bg color is important for td; just specify one */
    // background-color: #c1f4cd !important

  tr th
    position: sticky
    /* higher than z-index for td below */
    z-index: 2
    /* bg color is important; just specify one */
    background: #fff

  /* this will be the loading indicator */
  thead tr:last-child th
    /* height of all previous header rows */
    top: 48px
    /* highest z-index */
    z-index: 3
  thead tr:first-child th
    top: 0
    z-index: 1
  tr:first-child th:first-child
    /* highest z-index */
    z-index: 3

  td:first-child
    z-index: 1

  td:first-child, th:first-child
    position: sticky
    left: 0
</style>
