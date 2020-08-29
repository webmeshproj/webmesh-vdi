<template>
  <q-dialog ref="dialog" @hide="onDialogHide" transition-show="scale" transition-hide="scale" full-width>
    <q-card>
      <q-card-section>
        <q-splitter v-model="splitterModel">
          <template v-slot:before>
            <div class="q-pa-md">
              <q-tree
                :nodes="nodes"
                node-key="fullPath"
                selected-color="primary"
                :selected.sync="selected"
                @lazy-load="onLazyLoad"
                default-expand-all
              />
            </div>
          </template>

          <template v-slot:after>
            <q-tab-panels
              v-model="selected"
              animated
              transition-prev="jump-up"
              transition-next="jump-up"
            >
              <q-tab-panel v-for="info in nodeInfo" :key="info.fullPath" :name="info.fullPath">
                <div>
                  <div class="text-h6 q-mb-md">{{info.label}}</div>
                  <p><strong>Full Path: </strong>{{info.fullPath}}</p>
                  <p v-if="info.expandable && !info.lazy"><strong>Items: </strong>{{info.children.length}}</p>
                  <p v-if="!info.expandable"><strong>Size: </strong>{{fileSize(info.size)}}</p>
                  <div v-if="!info.expandable">
                    <q-btn flat :loading="previewing" label="Preview" @click="() => { fetchNodePreview(info) }" />
                    <q-btn flat :loading="downloading" label="Download" @click="() => { fetchNode(info) }" />
                  </div>
                </div>
              </q-tab-panel>

            </q-tab-panels>
          </template>
        </q-splitter>
      </q-card-section>

      <q-card-section>
        <q-btn flat label="Close" v-close-popup @click="onCancelClick" />
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script>
import FilePreviewDialog from './FilePreview.vue'

var path = require('path')

export default {
  name: 'FileTransferDialog',

  props: {
    desktopNamespace: { type: String },
    desktopName: { type: String }
  },

  data () {
    return {
      nodes: [],
      nodeInfo: [],
      splitterModel: 50,
      selected: '',
      downloading: false,
      previewing: false
    }
  },

  computed: {
    urlBase () { return `/api/desktops/fs/${this.desktopNamespace}/${this.desktopName}` },
    homeDir () {
      const user = this.$userStore.getters.user
      return `/home/${user.name}`
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
    },

    handleError (err) {
      this.$root.$emit('notify-error', err)
    },

    fileSize (bytes) {
      if (bytes === 0) {
        return '0.00 B'
      }
      const e = Math.floor(Math.log(bytes) / Math.log(1024))
      return (bytes / Math.pow(1024, e)).toFixed(2) + ' ' + ' KMGTP'.charAt(e) + 'B'
    },

    async onLazyLoad ({ node, key, done, fail }) {
      await new Promise((resolve, reject) => setTimeout(resolve, 250))
      const stat = await this.statPath(node.fullPath)
      if (!stat) { fail() }
      const newNode = await this.evaluateDirNode(stat, node.fullPath, false)
      node.lazy = false
      done(newNode.children)
    },

    async evaluateDirNode (stat, fullPath, isHome) {
      let icon
      let label
      if (isHome) {
        icon = 'home'
        label = this.homeDir
      } else {
        icon = 'folder'
        label = stat.name
      }
      const node = {
        fullPath: fullPath,
        label: label,
        icon: icon,
        children: [],
        expandable: true,
        lazy: false
      }
      this.nodeInfo.push(node)

      if (!stat.contents || !stat.contents.length) {
        return node
      }

      stat.contents.forEach((child) => {
        const childNode = {
          fullPath: `${node.fullPath}/${child.name}`,
          label: child.name,
          size: child.size,
          expandable: child.isDirectory
        }
        if (child.isDirectory) {
          childNode.icon = 'folder'
          childNode.lazy = true
        } else {
          childNode.icon = 'insert_drive_file'
        }

        this.nodeInfo.push(childNode)
        node.children.push(childNode)
      })

      return node
    },

    async fetchNodePreview (node) {
      const fpath = node.fullPath.replace(this.homeDir, '.')
      if (!node.size) {
        this.handleError(new Error(`${path.basename(fpath)} is an empty file`))
        return
      } else if (node.size > 1000000) {
        this.handleError(new Error('Preview is not supported for files over 1MB'))
        return
      }
      this.previewing = true
      await new Promise((resolve, reject) => setTimeout(resolve, 250))
      try {
        const res = await this.$axios.get(`${this.urlBase}/get/${fpath}`)
        this.previewing = false
        await this.$q.dialog({
          component: FilePreviewDialog,
          parent: this,
          src: res.data,
          name: path.basename(fpath)
        }).onOk(() => {
        }).onCancel(() => {
        }).onDismiss(() => {
        })
      } catch (err) {
        this.handleError(new Error(`Failed to download ${path.basename(fpath)}`))
      }
    },

    async fetchNode (node) {
      const fpath = node.fullPath.replace(this.homeDir, '.')
      if (!node.size) {
        this.handleError(new Error(`${path.basename(fpath)} is an empty file`))
        return
      }
      this.downloading = true
      try {
        const res = await this.$axios.get(`${this.urlBase}/get/${fpath}`, { responseType: 'blob' })

        const fileURL = window.URL.createObjectURL(new Blob([res.data]))
        const fileLink = document.createElement('a')

        fileLink.href = fileURL
        fileLink.setAttribute('download', path.basename(fpath))
        document.body.appendChild(fileLink)

        fileLink.click()
      } catch (err) {
        this.handleError(new Error(`Failed to download ${path.basename(fpath)}`))
      }
      this.downloading = false
    },

    async statPath (fpath) {
      try {
        fpath = fpath.replace(this.homeDir, '.')
        const res = await this.$axios.get(`${this.urlBase}/stat/${fpath}`)
        return res.data.stat
      } catch (err) {
        this.handleError(err)
        return null
      }
    }

  },

  async mounted () {
    await this.$nextTick()
    const root = await this.statPath('.')
    if (!root) {
      this.hide()
    }
    const rootNode = await this.evaluateDirNode(root, this.homeDir, true)
    this.selected = this.homeDir
    this.nodes = [rootNode]
  }

}
</script>
