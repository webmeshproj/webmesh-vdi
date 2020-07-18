import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

function uuidv4 () {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    var r = Math.random() * 16 | 0, v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

export const UserStore = new Vuex.Store({

  state: {
    status: '',
    token: localStorage.getItem('token') || '',
    requiresMFA: false,
    user: {},
    stateToken: ''
  },

  mutations: {

    async auth_request (state) {
      state.status = 'loading'
      const stateToken = localStorage.getItem('state')
      if (stateToken) {
        state.stateToken = stateToken
        return
      }
      state.stateToken = uuidv4()
      localStorage.setItem('state', state.stateToken)
    },

    auth_got_user (state, user) {
      state.user = user
    },

    auth_success (state, token) {
      state.status = 'success'
      state.token = token
      state.stateToken = ''
      state.requiresMFA = false
      localStorage.setItem('token', token)
      localStorage.removeItem('state')
    },

    auth_need_mfa (state) {
      state.requiresMFA = true
    },

    auth_error (state) {
      state.status = 'error'
      localStorage.removeItem('token')
      localStorage.removeItem('state')
    },

    logout (state) {
      state.status = ''
      state.user = {}
      state.token = ''
      state.stateToken = ''
      localStorage.removeItem('token')
      localStorage.removeItem('state')
    }

  },

  actions: {

    async initStore ({ commit }) {
      if (!this.getters.isLoggedIn) {
        console.log('Attempting anonymous/state login')
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
          commit('logout')
          this.dispatch('initStore')
        }
      }
    },

    async login ({ commit, state }, credentials) {
      try {
        await commit('auth_request')
        credentials.state = state.stateToken
        const res = await axios({ url: '/api/login', data: credentials, method: 'POST' })
        if (res.headers['x-redirect']) {
          window.location = res.headers['x-redirect']
          return
        }
        const token = res.data.token
        const user = res.data.user
        const authorized = res.data.authorized
        Vue.prototype.$axios.defaults.headers.common['X-Session-Token'] = token
        commit('auth_got_user', user)
        if (authorized) {
          commit('auth_success', token)
          return
        }
        commit('auth_need_mfa')
      } catch (err) {
        commit('auth_error')
        throw err
      }
    },

    async authorize ({ commit }, otp) {
      const res = await axios({ url: '/api/authorize', data: { otp: otp }, method: 'POST' })
      const token = res.data.token
      const authorized = res.data.authorized
      Vue.prototype.$axios.defaults.headers.common['X-Session-Token'] = token
      if (authorized) {
        commit('auth_success', token)
      }
    },

    async logout ({ commit }) {
      commit('logout')
      try {
        await Vue.prototype.$axios.post('/api/logout')
        delete Vue.prototype.$axios.defaults.headers.common['X-Session-Token']
        this.dispatch('initStore')
      } catch (err) {
        console.error(err)
        throw err
      }
    }

  },

  getters: {
    isLoggedIn: state => !!state.token,
    requiresMFA: state => state.requiresMFA,
    authStatus: state => state.status,
    user: state => state.user,
    token: state => state.token
  }

})

export default UserStore
