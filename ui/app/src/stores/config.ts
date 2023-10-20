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

  state: (): {_serverConfig: any, axios: axios.Axios,emitter: ReturnType<typeof mitt>} => ({
    _serverConfig: {},
    axios: axios.default,
    emitter: emitter
  }),

  actions: {
    set_server_config (data: any) {
      this._serverConfig = data
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
    serverConfig: (state) => {return state._serverConfig},
    grafanaEnabled: state => {
      console.log(state)
      if (state._serverConfig?.metrics && state._serverConfig.metrics.grafana) {
        return state._serverConfig.metrics.grafana.enabled || false
      }
      return false
    },
    authMethod: state => {
      if (state._serverConfig.auth !== undefined) {
        if (state._serverConfig.auth.ldapAuth !== undefined && state._serverConfig.auth.ldapAuth.url) {
          return 'ldap'
        }
        if (state._serverConfig.auth.oidcAuth !== undefined && state._serverConfig.auth.oidcAuth.issuerURL) {
          return 'oidc'
        }
        if (state._serverConfig.auth.webmeshAuth !== undefined && state._serverConfig.auth.webmeshAuth.metadataURL) {
          return 'webmesh'
        }
      }
      return 'local'
    }
  }
});

