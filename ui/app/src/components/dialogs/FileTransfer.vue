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
                v-model:selected="selected"
                v-model:expanded="expanded"
                accordion
                @lazy-load="onLazyLoad"
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
                  <div>
                    <q-btn v-if="!info.expandable" flat :loading="previewing" label="Preview" @click="() => { fetchNodePreview(info) }" />
                    <q-btn flat label="Download" @click="() => { fetchNode(info) }" />
                  </div>
                  <div>
                    <q-linear-progress v-if="downloading || downloaded" :indeterminate="downloadIndeterminate" :value="downloadProgress" rounded color="secondary" size="25px">
                      <div class="absolute-full flex flex-center">
                        <q-badge color="white" text-color="secondary" :label="downloadLabel" />
                      </div>
                    </q-linear-progress>
                  </div>
                </div>
              </q-tab-panel>

            </q-tab-panels>
          </template>
        </q-splitter>
      </q-card-section>

      <q-card-section style="display: inline-block; width: 50vw;">
        <q-file dense filled bottom-slots v-model="fileToUpload" label="Upload a file" counter>
          <template v-slot:prepend>
            <q-icon name="cloud_upload" @click.stop />
          </template>
          <template v-slot:append>
            <q-icon name="close" @click.stop="fileToUpload = null" class="cursor-pointer" />
          </template>
          <template v-slot:hint>
            Select a file to upload to {{homeDir}}/Uploads
          </template>
          <template v-slot:after>
            <q-btn round dense flat icon="send" :loading="uploading" @click="onUpload" />
          </template>
        </q-file>
        <q-btn flat label="Close" v-close-popup @click="onCancelClick" />
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script lang="ts">
import FilePreviewDialog from './FilePreview.vue'
import { getErrorMessage } from '../../lib/util.js'
import path from 'path'

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
      expanded: [],
      previewing: false,
      uploading: false,
      fileToUpload: null,
      downloading: false,
      downloaded: false,
      downloadProgress: 0,
      downloadIndeterminate: true
    }
  },

  computed: {
    urlBase () { return `/api/desktops/fs/${this.desktopNamespace}/${this.desktopName}` },
    homeDir () {
      const user = this.userStore.getters.user
      return `/home/${user.name}`
    },
    downloadLabel () {
      if (this.downloadIndeterminate) {
        return 'Preparing download'
      } else if (this.downloadProgress !== 1) {
        return (this.downloadProgress * 100).toFixed(2) + '%'
      }
      return 'Saving to downloads'
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
      this.configStore.emitter.emit('notify-error', err)
    },

    fileSize (bytes) {
      if (bytes === 0) {
        return '0.00 B'
      }
      const e = Math.floor(Math.log(bytes) / Math.log(1024))
      return (bytes / Math.pow(1024, e)).toFixed(2) + ' ' + ' KMGTP'.charAt(e) + 'B'
    },

    async onUpload () {
      if (!this.fileToUpload) { return }
      await new Promise((resolve, reject) => setTimeout(resolve, 250))
      this.uploading = true
      const formData = new FormData()
      formData.append('file', this.fileToUpload)
      try {
        await this.$axios.put(`${this.urlBase}/put`, formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
      } catch (err) {
        this.handleError(err)
      }
      this.$q.notify({
        color: 'green-4',
        textColor: 'white',
        icon: 'cloud_done',
        message: `${this.fileToUpload.name} uploaded to ${this.homeDir}/Uploads/${this.fileToUpload.name}`
      })
      // await this.syncRootNode()
      // if (this.expanded.indexOf(`${this.homeDir}/Uploads`) === -1) {
      //   this.expanded.push(`${this.homeDir}/Uploads`)
      // }
      this.fileToUpload = null
      this.uploading = false
      this.hide()
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
        lazy: false,
        // reset progress bar if we switch tabs
        handler: () => { this.downloaded = false }
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
          expandable: child.isDirectory,
          // reset progress bar if we switch tabs
          handler: () => { this.downloaded = false }
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
        const errMsg = await getErrorMessage(err)
        this.handleError(new Error(`Failed to preview ${path.basename(fpath)}: ${errMsg}`))
      }
    },

    onDownloadProgress (ev) {
      this.downloadIndeterminate = false
      const current = ev.loaded
      let total
      if (ev.lengthComputable) {
        total = ev.total
      } else {
        const contentLength = ev.target.getResponseHeader('x-decompressed-content-length')
        total = parseInt(contentLength, 10)
      }
      this.downloadProgress = (current / total).toFixed(4)
    },

    async fetchNode (node) {
      const fpath = node.fullPath.replace(this.homeDir, '.')
      // Check if its an empty file. The home dir will be considered as 0 also so
      // make sure it isn't a directory.
      if (!node.size && !node.expandable) {
        this.handleError(new Error(`${path.basename(fpath)} is an empty file`))
        return
      }
      this.downloadIndeterminate = true
      this.downloadProgress = 0
      this.downloading = true
      this.downloaded = false
      try {
        // Use a longer timeout to give the server time to archive in case it's a directory
        const res = await this.$axios.get(`${this.urlBase}/get/${fpath}`, {
          timeout: 300 * 1000,
          responseType: 'blob',
          onDownloadProgress: this.onDownloadProgress
        })
        // convert the downloaded body to a link object and download it
        const localname = this.getFilenameForDownload(fpath, res.headers)
        const fileURL = window.URL.createObjectURL(new Blob([res.data]))
        const fileLink = document.createElement('a')
        fileLink.href = fileURL
        fileLink.setAttribute('download', localname)
        document.body.appendChild(fileLink)
        fileLink.click()
      } catch (err) {
        const errMsg = await getErrorMessage(err)
        this.handleError(new Error(`Failed to download ${path.basename(fpath)}: ${errMsg}`))

        this.downloaded = false
        this.downloading = false
        return
      }

      this.downloadProgress = 1
      this.downloading = false
      this.downloaded = true
    },

    getFilenameForDownload (fpath, headers) {
      const disposition = headers['content-disposition']
      const suggestedFname = headers['x-suggested-filename']
      // try the disposition header first
      if (disposition && disposition !== '') {
        const fname = disposition.split('filename=')[1].split(';')[0]
        console.log(`Using ${fname} from disposition header`)
        return fname
      }
      // next try the suggested filename
      if (suggestedFname && suggestedFname !== '') {
        console.log(`Using ${suggestedFname} from suggested filename header`)
        return suggestedFname
      }
      // finally use the name of the path in the url
      const pathname = path.basename(fpath)
      console.log(`Falling back to ${pathname} from URL path`)
      return pathname
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
    },

    async syncRootNode () {
      this.expanded = []
      const root = await this.statPath('.')
      if (!root) {
        this.hide()
      }
      const rootNode = await this.evaluateDirNode(root, this.homeDir, true)
      this.selected = this.homeDir
      this.expanded = [this.homeDir]
      this.nodes = [rootNode]
    }

  },

  mounted () {
    this.$nextTick().then(() => { this.syncRootNode() })
  }

}
</script>
