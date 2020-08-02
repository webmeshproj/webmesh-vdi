import Vue from 'vue'
import Vuex from 'vuex'

export const ConfigStore = new Vuex.Store({

  state: {
    serverConfig: {},
    localConfig: {
      readWriteMany: false
    }
  },

  mutations: {

    set_server_config (state, data) {
      state.serverConfig = data
    },

    set_read_write_many (state, val) {
      state.localConfig.readWriteMany = val
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
    },
    setReadWriteMany ({ commit }, val) {
      commit('set_read_write_many', val)
    }
  },

  getters: {
    localConfig: state => state.localConfig,
    serverConfig: state => state.serverConfig,
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
