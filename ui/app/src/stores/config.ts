/*
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
*/

import { defineStore } from 'pinia'
import * as axios from 'axios'
import mitt from 'mitt';
const emitter = mitt()
export const  useConfigStore =  defineStore('configStore', {

  state: (): {serverConfig: any, axios: axios.Axios,emitter: ReturnType<typeof mitt>} => ({
    serverConfig: {},
    axios: axios.default,
    emitter: emitter
  }),

  actions: {
    set_server_config (data: any) {
      this.serverConfig = data
    },
    async getServerConfig () {
      try {
        const res = await this.axios.get('/api/config')
        this.set_server_config(res.data)
      } catch (err) {
        console.log('Failed to retrieve server config')
        console.error(err)
        throw err
      }
    }
  },

  getters: {
    serverConfig: (state) => {return state.serverConfig},
    grafanaEnabled: state => {
      console.log(state)
      if (state.serverConfig?.metrics && state.serverConfig.metrics.grafana) {
        return state.serverConfig.metrics.grafana.enabled || false
      }
      return false
    },
    authMethod: state => {
      if (state.serverConfig.auth !== undefined) {
        if (state.serverConfig.auth.ldapAuth !== undefined && state.serverConfig.auth.ldapAuth.URL) {
          return 'ldap'
        }
        if (state.serverConfig.auth.oidcAuth !== undefined && state.serverConfig.auth.oidcAuth.IssuerURL) {
          return 'oidc'
        }
      }
      return 'local'
    }
  }
});

