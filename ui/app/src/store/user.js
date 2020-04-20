import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

export const UserStore = new Vuex.Store({

  state: {
    status: '',
    token: localStorage.getItem('token') || '',
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

    auth_error (state) {
      state.status = 'error'
    },

    logout (state) {
      state.status = ''
      state.user = ''
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
          commit('auth_got_user', res.data.user)
          commit('auth_success', res.data.token)
          console.log(`Resuming session as ${res.data.user.name}`)
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
        localStorage.setItem('token', token)
        Vue.prototype.$axios.defaults.headers.common['X-Session-Token'] = token
        commit('auth_got_user', user)
        commit('auth_success', token)
      } catch (err) {
        commit('auth_error')
        localStorage.removeItem('token')
        throw err
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
    authStatus: state => state.status,
    username: state => state.user.name,
    token: state => state.token
  }

})

export default UserStore
