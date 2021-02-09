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
import axios from 'axios'

function uuidv4 () {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    var r = Math.random() * 16 | 0, v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

function getMsUntilExpire (expiresAt) {
  const now = Math.round((new Date()).getTime() / 1000)
  return (expiresAt - now) * 1000
}

export const UserStore = new Vuex.Store({

  state: {
    status: '',
    token: localStorage.getItem('token') || '',
    renewable: localStorage.getItem('renewable') === 'true' || false,
    requiresMFA: false,
    user: {},
    stateToken: '',
    timeout: null
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

    auth_success (state, { token, renewable }) {
      state.status = 'success'
      state.token = token
      state.stateToken = ''
      state.requiresMFA = false
      state.renewable = renewable
      localStorage.setItem('token', token)
      localStorage.setItem('renewable', String(renewable))
      localStorage.removeItem('state')
    },

    auth_need_mfa (state) {
      state.requiresMFA = true
    },

    auth_error (state) {
      state.status = 'error'
      state.user = {}
      state.token = ''
      state.stateToken = ''
      state.renewable = false
      localStorage.removeItem('token')
      localStorage.removeItem('state')
      localStorage.removeItem('renewable')
    },

    logout (state) {
      state.status = ''
      state.user = {}
      state.token = ''
      state.stateToken = ''
      state.renewable = false
      if (state.timeout !== null) {
        clearTimeout(state.timeout)
        state.timeout = null
      }
      localStorage.removeItem('token')
      localStorage.removeItem('state')
      localStorage.removeItem('renewable')
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
          if (this.getters.renewable) {
            await this.dispatch('refreshToken')
          }
          const res = await Vue.prototype.$axios.get('/api/whoami')
          commit('auth_got_user', res.data)
          if (res.data.sessions) {
            res.data.sessions.forEach((item) => {
              console.log(`Adding existing session ${item.namespace}/${item.name}`)
              Vue.prototype.$desktopSessions.dispatch('addExistingSession', item)
            })
          }
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

        const resState = res.data.state
        if (state.stateToken !== resState) {
          console.log('State token was malformed during request flow!')
          commit('auth_error')
          throw new Error('State token was malformed during request flow!')
        }

        if (res.headers['x-redirect']) {
          window.location = res.headers['x-redirect']
          return
        }

        const token = res.data.token
        const user = res.data.user
        const authorized = res.data.authorized
        const expiresAt = res.data.expiresAt
        const renewable = res.data.renewable

        Vue.prototype.$axios.defaults.headers.common['X-Session-Token'] = token
        commit('auth_got_user', user)
        if (authorized) {
          commit('auth_success', { token, renewable })
          if (renewable) {
            state.timeout = setTimeout(() => { this.dispatch('refreshToken') }, getMsUntilExpire(expiresAt))
          }
          return
        }
        commit('auth_need_mfa')
      } catch (err) {
        commit('auth_error')
        throw err
      }
    },

    async refreshToken ({ commit, state }) {
      console.log('Refreshing access token')
      try {
        const res = await axios({ url: '/api/refresh_token', method: 'GET' })

        const token = res.data.token
        const expiresAt = res.data.expiresAt
        const renewable = res.data.renewable

        Vue.prototype.$axios.defaults.headers.common['X-Session-Token'] = token
        if (renewable) {
          state.timeout = setTimeout(() => { this.dispatch('refreshToken') }, getMsUntilExpire(expiresAt))
        }
        commit('auth_success', { token, renewable })
      } catch (err) {
        commit('auth_error')
        let error
        if (err.response !== undefined && err.response.data !== undefined) {
          error = err.response.data.error
        } else {
          error = err.message
        }
        Vue.prototype.$q.notify({
          color: 'red-4',
          textColor: 'black',
          icon: 'error',
          message: error
        })
      }
    },

    async authorize ({ commit, state }, otp) {
      const res = await axios({ url: '/api/authorize', data: { otp: otp, state: state.stateToken }, method: 'POST' })
      const resState = res.data.state
      if (state.stateToken !== resState) {
        console.log('State token was malformed during request flow!')
        commit('auth_error')
        throw new Error('State token was malformed during request flow!')
      }
      const token = res.data.token
      const authorized = res.data.authorized
      const expiresAt = res.data.expiresAt
      const renewable = res.data.renewable
      Vue.prototype.$axios.defaults.headers.common['X-Session-Token'] = token
      if (authorized) {
        commit('auth_success', { token, renewable })
        if (renewable) {
          state.timeout = setTimeout(() => { this.dispatch('refreshToken') }, getMsUntilExpire(expiresAt))
        }
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
    token: state => state.token,
    stateToken: state => state.stateToken,
    renewable: state => state.renewable
  }

})

export default UserStore
