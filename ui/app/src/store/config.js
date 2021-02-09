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

import Vue from 'vue'
import Vuex from 'vuex'

export const ConfigStore = new Vuex.Store({

  state: {
    serverConfig: {}
  },

  mutations: {

    set_server_config (state, data) {
      state.serverConfig = data
    }

  },

  actions: {
    async getServerConfig ({ commit }) {
      try {
        const res = await Vue.prototype.$axios.get('/api/config')
        commit('set_server_config', res.data)
      } catch (err) {
        console.log('Failed to retrieve server config')
        console.error(err)
        throw err
      }
    }
  },

  getters: {
    serverConfig: state => state.serverConfig,
    grafanaEnabled: state => {
      if (state.serverConfig.metrics && state.serverConfig.metrics.grafana) {
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

})

export default ConfigStore
