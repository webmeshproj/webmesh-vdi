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
  <q-page padding>
    <div>
      <q-space />
    </div>
    <div style="float: right">
      <q-btn flat color="primary" :loading="refreshLoading" @click="refreshData" label="Refresh" />
      <q-btn flat color="primary" label="New Template" @click="onNewTemplate" />
    </div>

    <div style="clear: right">
      <SkeletonTable v-if="loading"/>
      <q-table
        class="templates-table"
        title="Desktop Templates"
        :rows="data"
        :columns="columns"
        row-key="name"
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
            <q-td key="useTemplate" :props="props">
              <q-btn round dense flat icon="cast"  size="md" color="blue" @click="onLaunchTemplate(props.row)">
                <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Launch Template</q-tooltip>
              </q-btn>
              <q-btn round dense flat icon="create"  size="md" color="orange" @click="onEditTemplate(props.row)">
                <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Edit Template</q-tooltip>
              </q-btn>
              <q-btn round dense flat icon="remove_circle"  size="md" color="red" @click="onConfirmDeleteTemplate(props.row.metadata.name)">
                <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Delete Template</q-tooltip>
              </q-btn>
            </q-td>

            <q-td key="image" :props="props">
              <strong>{{ imageName(props.row.spec) }}</strong>
            </q-td>

            <q-td key="root" :props="props">
              <q-avatar v-if="rootEnabled(props.row.spec)" size="27px" font-size="20px" color="green" text-color="white" icon="done" />
            </q-td>

            <q-td key="fileXfer" :props="props">
              <q-avatar v-if="xferEnabled(props.row.spec)" size="27px" font-size="20px" color="green" text-color="white" icon="done" />
            </q-td>

            <q-td key="tags" :props="props">
              <div class="tags-wrapper">
                <li class="inline-tags" v-for="tag in tagsToArray(props.row.spec.tags)" :key="tag" dense>
                    <q-chip dense icon="bookmark">{{ tag }}</q-chip>
                </li>
              </div>
            </q-td>

            <q-td key="serviceaccount" :props="props">
              <ServiceAccountSelector :ref="`sa-${props.row.metadata.name}`" :parentRefs="$refs" :tmplName="props.row.metadata.name" :label="`Service Account (default)`" />
            </q-td>

            <q-td key="namespace" :props="props">
              <NamespaceSelector :ref="`ns-${props.row.metadata.name}`" :multiSelect="false" :showAllOption="false" :label="`Launch Namespace (${defaultNamespace})`" />
            </q-td>

 
          </q-tr>
        </template>
      </q-table>
    </div>

  </q-page>
</template>

<script lang="ts">
import SkeletonTable from '../components/SkeletonTable.vue'
import NamespaceSelector from '../components/inputs/NamespaceSelector.vue'
import ServiceAccountSelector from '../components/inputs/ServiceAccountSelector.vue'
import TemplateEditor from '../components/dialogs/TemplateEditor.vue'
import ConfirmDelete from '../components/dialogs/ConfirmDelete.vue'

const templateColums = [
  {
    name: 'name',
    required: true,
    label: 'Template',
    align: 'left' as const,
    field: row => row.name,
    format: val => `${val}`,
    sortable: true,/* 
    classes: 'bg-grey-2 ellipsis',
    headerClasses: 'bg-primary text-white'  */
  },
  {
    name: 'useTemplate',
    align: 'center',
    label: 'Actions'
  } ,
  {
    name: 'image',
    align: 'left',
    label: 'Image'
  },

  {
    name: 'root',
    align: 'center',
    label: 'Root'
  },
  {
    name: 'fileXfer',
    align: 'center',
    label: 'File Transfer'
  },
  {
    name: 'tags',
    align: 'center',
    label: 'Tags'
  },
  {
    name: 'serviceaccount',
    align: 'center',
    label: 'ServiceAccount'
  },
  {
    name: 'namespace',
    align: 'center',
    label: 'Namespace'
  },
  
]
import { defineComponent,reactive,ref } from 'vue'
import { useConfigStore } from 'src/stores/config'
import { useDesktopSessions } from 'src/stores/desktop'
export default defineComponent({
  name: 'DesktopTemplates',
  components: { SkeletonTable, NamespaceSelector, ServiceAccountSelector },

  setup () {
    return {
      configStore: useConfigStore(),
      desktopSessions: useDesktopSessions(),
      loading: ref(false),
      refreshLoading: false,
      columns: templateColums,
      data: ref([] as any[])
    }
  },

  computed: {
    defaultNamespace () { return this.configStore._serverConfig.appNamespace || 'default' }
  },

  methods: {
    rootEnabled (spec) { return spec.desktop && spec.desktop.allowRoot },
    xferEnabled (spec) { return spec.proxy && spec.proxy.allowFileTransfer },

    imageName (spec) {
      if (spec.desktop && spec.desktop.image) { return spec.desktop.image }
      if (spec.qemu && spec.qemu.diskImage) { return spec.qemu.diskImage }
      return ''
    },

    async onNewTemplate () {
      this.$q.dialog({
        component: TemplateEditor,
       componentProps:{
        parent: this
       }
      }).onOk(async () => {
        await new Promise((resolve) => setTimeout(resolve, 300))
        this.refreshData()
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    onLaunchTemplate (template) {
      console.log(`Launching: ${template.metadata.name}`)
      const ns = this.$refs[`ns-${template.metadata.name}`].selection
      const sa = this.$refs[`sa-${template.metadata.name}`].selection
      const payload: any = { template: template }
      // When the attribute comes back as an object, it actually means
      // no selection
      if (typeof ns !== 'object') {
        payload.namespace = ns
      } else {
        // default to the app namespace for now.
        // this is so read-only users select the correct namespace by default.
        payload.namespace = this.configStore._serverConfig.appNamespace
      }
      if (typeof sa !== 'object') {
        payload.serviceAccount = sa
      }
      this.doLaunchTemplate(payload)
    },

    async onEditTemplate (template) {
      this.$q.dialog({
        component: TemplateEditor,
       componentProps: {
        parent: this,
        existing: this.pruneTemplateObject(template)
       }
      }).onOk(async () => {
        await new Promise((resolve) => setTimeout(resolve, 300))
        this.refreshData()
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    onConfirmDeleteTemplate (templateName) {
      this.$q.dialog({
        component: ConfirmDelete,
        componentProps: {
          parent: this,
        resourceName: templateName
        }
      }).onOk(() => {
        this.doDeleteTemplate(templateName)
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    pruneTemplateObject (template) {
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
      const tags: string[] = []
      if (tagsObj === undefined || tagsObj === null) { return tags }
      Object.keys(tagsObj).forEach((key) => {
        tags.push(`${key}: ${tagsObj[key]}`)
      })
      return tags
    },

    async doLaunchTemplate (payload) {
      try {
        await this.desktopSessions.newSession( payload)
        this.configStore.emitter.emit('set-control')
        this.$router.push('control')
      } catch (err) {
        this.configStore.emitter.emit('notify-error', err)
      }
    },

    async doDeleteTemplate (templateName) {
      try {
        await this.configStore.axios.delete(`/api/templates/${templateName}`)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Deleted DesktopTemplate '${templateName}'`
        })
        this.fetchData()
      } catch (err) {
        this.configStore.emitter.emit('notify-error', err)
      }
    },

    async refreshData () {
    //  this.data.splice(0) // clear array
      this.refreshLoading = true
      await new Promise((resolve) => setTimeout(resolve, 500))
      this.fetchData()
      this.refreshLoading = false
    },

    async fetchData () {
      try {
       // this.data.splice(0) // clear array
        const res = await this.configStore.axios.get('/api/templates')
        res.data.forEach((tmpl) => { this.data.push(tmpl) })
        console.log(this.data)
      } catch (err) {
        this.configStore.emitter.emit('notify-error', err)
      }
    }
  },

  async mounted () {
    await this.$nextTick()
    this.loading = true
    await new Promise((resolve) => setTimeout(resolve, 500))
    await this.fetchData()
    this.loading = false
  }
})
</script>


<style lang="scss" scoped>
.tags-wrapper {
  position: relative;
  width: 40vh;
}

.inline-tags
  {display: inline;}

.templates-table {

/*   background-color: $grey-3 */

  /* height or max-height is important */
  max-height: 500px;

  // specifying max-width so the example can  highlight the sticky column on any browser window 
  // max-width: 600px


  tr th {
    position: sticky;
    /* higher than z-index for td below */
    z-index: 2;
    /* bg color is important; just specify one */
    background: #fff;
  }

  /* this will be the loading indicator */
  thead tr:last-child th {

    /* height of all previous header rows */
    top: 48px;
    /* highest z-index */
    z-index: 3;
  }
  thead tr:first-child th {

    top: 0;
    z-index: 1;
  }
  tr:first-child th:first-child
  {
      /* highest z-index */
      z-index: 3;
  }

  td:first-child {
      z-index: 1;
  }

  td:first-child, th:first-child {
    position: sticky;
    left: 0;
  }
}
</style>
 