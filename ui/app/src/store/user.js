import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

export const UserStore = new Vuex.Store({

  state: {
    status: '',
    token: localStorage.getItem('token') || '',
    needMFA: false,
    user: {}
  },

  mutations: {

    auth_request (state) {
      state.status = 'loading'
    },

    auth_got_user (state, user) {
      state.user = user
    },

    auth_success (state, token) {
      state.status = 'success'
      state.token = token
    },

    auth_need_mfa (state) {
      state.needMFA = true
    },

    auth_error (state) {
      state.status = 'error'
    },

    logout (state) {
      state.status = ''
      state.user = {}
      state.token = ''
    }

  },

  actions: {

    async initStore ({ commit }) {
      if (!this.getters.isLoggedIn) {
        console.log('Attempting anonymous login')
        try {
          return await this.dispatch('login', { username: 'anonymous' })
        } catch (err) {
          console.log('Could not authenticate as anonymous')
        }
      } else {
        Vue.prototype.$axios.defaults.headers.common['X-Session-Token'] = this.state.token
        try {
          console.log('Retrieving user information')
          const res = await Vue.prototype.$axios.get('/api/whoami')
          commit('auth_got_user', res.data)
          console.log(`Resuming session as ${res.data.name}`)
        } catch (err) {
          console.log('Could not fetch user information')
          console.log(err)
          commit('auth_error')
          commit('logout')
          localStorage.removeItem('token')
          throw err
        }
      }
    },

    async login ({ commit }, credentials) {
      try {
        commit('auth_request')
        const res = await axios({ url: '/api/login', data: credentials, method: 'POST' })
        const token = res.data.token
        const user = res.data.user
        const authorized = res.data.authorized
        localStorage.setItem('token', token)
        Vue.prototype.$axios.defaults.headers.common['X-Session-Token'] = token
        commit('auth_got_user', user)
        if (authorized) {
          commit('auth_success', token)
          return
        }
      } catch (err) {
        commit('auth_error')
        localStorage.removeItem('token')
        throw err
      }
    },

    async authorize ({ commit }, otp) {
      const res = await axios({ url: '/api/authorize', data: { otp: otp }, method: 'POST' })
      const token = res.data.token
      const authorized = res.data.authorized
      localStorage.setItem('token', token)
      Vue.prototype.$axios.defaults.headers.common['X-Session-Token'] = token
      if (authorized) {
        commit('auth_success', token)
      }
    },

    async logout ({ commit }) {
      commit('logout')
      localStorage.removeItem('token')
      try {
        const res = await Vue.prototype.$axios.post('/api/logout')
        delete Vue.prototype.$axios.defaults.headers.common['X-Session-Token']
        return res
      } catch (err) {
        console.error(err)
        throw err
      }
    }

  },

  getters: {
    isLoggedIn: state => !!state.token,
    requiresMFA: state => state.needMFA,
    authStatus: state => state.status,
    user: state => state.user,
    token: state => state.token
  }

})

export default UserStore
