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
  <div class="fixed-center text-center">
    <div class="text-h6">Please login to use<br/>Webmesh Desktop</div>
    <br />
    <q-form
      @submit="onSubmit"
      @reset="onReset"
    >
      <q-input
        :loading="loading"
        input-style="width: 300px;"
        rounded standout
        v-model="username"
        label="Username"
        lazy-rules
        :rules="[ val => val && val.length > 0 || 'Username cannot be blank']"
      />
      <q-input
        rounded standout
        type="password"
        v-model="password"
        label="Password"
      />
      <br />
      <q-btn label="Login" type="submit" color="primary"/>
      <q-btn label="Reset" type="reset" color="primary" flat class="q-ml-sm" />
    </q-form>
  </div>
</template>

<script lang="ts">
import MFADialog from '../components/dialogs/MFADialog.vue'

import { defineComponent } from 'vue'
import { useConfigStore } from '../stores/config'
import { useUserStore } from 'src/stores/user'

export default defineComponent({
  name: 'Login',

  data () {
    return {
      username: null,
      password: null,
      loading: false,
      userStore: useUserStore(),
      configStore: useConfigStore()
    }
  },

  methods: {
    async initAuthFlow () {
      try {
        await this.userStore.initStore()
        const requiresMFA = this.userStore.requiresMFA
        if (requiresMFA) {
          // MFA Required
          await this.$q.dialog({
            component: MFADialog,
            componentProps: {
              parent: this
            }
          }).onOk(() => {
            this.notifyLoggedIn()
          }).onCancel(() => {
          }).onDismiss(() => {
          })
          return
        }
        await this.notifyLoggedIn()
      } catch (err) {
        this.configStore.emitter.emit('notify-error', err)
      }
    },

    async onSubmit () {
      try {
        await this.userStore.login({ username: this.username, password: this.password })
        const requiresMFA = this.userStore.requiresMFA
        if (requiresMFA) {
          // MFA Required
          await this.$q.dialog({
            component: MFADialog,
           componentProps: {
            parent: this
           }
          }).onOk(() => {
            this.notifyLoggedIn()
          }).onCancel(() => {
          }).onDismiss(() => {
          })
          return
        }
        await this.notifyLoggedIn()
      } catch (err) {
        console.error(err)
        this.configStore.emitter.emit('notify-error', err)
      }
    },

    onReset () {
      this.username = null
      this.password = null
    },

    async getAuthMethod(): Promise<string> {
      const res = await this.configStore.axios.get('/api/auth_methods')
      const data = res.data
      if (data.local) {
        return 'local'
      } else if (data.ldap) {
        return 'ldap'
      } else if (data.oidc) {
        return 'oidc'
      } else if (data.webmesh) {
        return 'webmesh'
      } else {
        return 'local'
      }
    },

    async notifyLoggedIn () {
      await this.configStore.getServerConfig()
      this.configStore.emitter.emit('set-logged-in', this.username)
      this.configStore.emitter.emit('set-active-title', 'Desktop Templates')
      this.$router.push('templates')
      this.$q.notify({
        color: 'green-4',
        textColor: 'white',
        icon: 'cloud_done',
        message: `Logged in as ${this.username}`
      })
    }
  },

  mounted () {
    this.$nextTick().then(() => {
      this.configStore.emitter.emit('set-active-title', 'Login')
      this.getAuthMethod().then((authMethod) => {
        if (authMethod === 'webmesh') {
          this.configStore.axios.get('http://169.254.169.254/id-tokens/issue').then((res) => {
              this.username = res.data.id
              this.password = res.data.token
              this.onSubmit()
          }).catch((err: Error) => {
            console.error(err)
          })
        }
      })
    })
  }
})
</script>
