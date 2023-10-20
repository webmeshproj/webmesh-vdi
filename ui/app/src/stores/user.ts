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
import axios from 'axios'
import { useDesktopSessions } from './desktop'
import { useConfigStore } from './config'
import { useQuasar } from 'quasar'

function uuidv4 () {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = Math.random() * 16 | 0, v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

const broadcastNewToken = new BroadcastChannel('kvdi_new_token')

export const useUserStore = defineStore('userStore',{


  state: (): {_status: string, _token: string, _renewable: boolean, _requiresMFA: boolean, _user:any, _stateToken:string} =>  ({
    _status: '',
    _token: localStorage.getItem('token') || '',
    _renewable: localStorage.getItem('renewable') === 'true' || false,
    _requiresMFA: false,
    _user: {},
    _stateToken: ''
  }),

  actions: {
    async auth_request () {
      this._status = 'loading'
      const stateToken = localStorage.getItem('state')
      if (stateToken) {
        this._stateToken = stateToken
        return
      }
      this._stateToken = uuidv4()
      localStorage.setItem('state', this._stateToken)
    },

    auth_got_user (user: any) {
      this._user = user
    },

    auth_success ( { token, renewable }: any) {
      this._status = 'success'
      this._token = token
      this._renewable = renewable
      localStorage.setItem('token', token)
      localStorage.setItem('renewable', String(renewable))

      this._stateToken = ''
      this._requiresMFA = false
      localStorage.removeItem('state')
    },

    auth_need_mfa () {
      this._requiresMFA = true
    },

    auth_error () {
      this._status = 'error'
      this._user = {}
      this._token = ''
      this._stateToken = ''
      this._renewable = false
      localStorage.removeItem('token')
      localStorage.removeItem('state')
      localStorage.removeItem('renewable')
    },

    logout_mut () {
      this._status = ''
      this._user = {}
      this._token = ''
      this._stateToken = ''
      this._renewable = false
      localStorage.removeItem('token')
      localStorage.removeItem('state')
      localStorage.removeItem('renewable')
    },

    async initStore () {
      /* throw new Error("TODO") */
      useConfigStore().axios.interceptors.response.use(null, (error: any) => {
        if (error.config && error.response) {
          const { config, response: { status } } = error
          const originalRequest = config
          if (status === 401) {
            return this.refreshToken().then((token) => {
              originalRequest.headers['X-Session-Token'] = token
              return useConfigStore().axios.request(originalRequest)
            })
          }
          return Promise.reject(error)
        }
        return Promise.reject(error)
      })
      broadcastNewToken.addEventListener('message', (ev) => {
        console.log('Got new token from other browser session')
        this.auth_success({ token: ev.data.token, renewable: ev.data.renewable })
        useConfigStore().axios.defaults.headers.common['X-Session-Token'] = ev.data.token
      })
      if (!this.isLoggedIn) {
        console.log('Attempting anonymous/state login')
        try {
          return await this.login({ username: 'anonymous' })
        } catch (err) {
          console.log('Could not authenticate as anonymous')
        }
      } else {
        useConfigStore().axios.defaults.headers.common['X-Session-Token'] = this._token
        try {
          console.log('Retrieving user information')
          const res = await useConfigStore().axios.get('/api/whoami')
          this.auth_got_user(res.data)
          if (res.data.sessions) {
            res.data.sessions.forEach(async (item: any) => {
              console.log(`Adding existing session ${item.namespace}/${item.name}`)
              try {
                const templateData = await useConfigStore().axios.get(`/api/templates/${item.template}`)
                item.template = { spec: templateData }
                useDesktopSessions().addExistingSession( item)
              } catch (err) {
                useQuasar().notify({
                  color: 'red-4',
                  textColor: 'black',
                  icon: 'error',
                  message: `Error fetching metadata for ${item.namespace}/${item.name} - you will not be able to reconnect`
                }) 
                console.error(err)
              }
            })
          }
          console.log(`Resuming session as ${res.data.name}`)
        } catch (err) {
          console.log('Could not fetch user information')
          console.log(err)
          this.logout()
        }
      }
    },

    async login(credentials: any) {
      try {
        await this.auth_request()
        credentials.state = this._stateToken
        const res = await axios({ url: '/api/login', data: credentials, method: 'POST' })
        const resState = res.data.state
        if (this._stateToken !== resState) {
          console.log('State token was malformed during request flow!')
          this.auth_error()
          throw new Error('State token was malformed during request flow!')
        }

        if (res.headers['x-redirect']) {
          window.location = res.headers['x-redirect']
          return
        }

        const token = res.data.token
        const user = res.data.user
        const authorized = res.data.authorized
        const renewable = res.data.renewable

        useConfigStore().axios.defaults.headers.common['X-Session-Token'] = token
        this.auth_got_user(user)
        if (authorized) {
          this.auth_success( { token, renewable })
          return
        }
       this.auth_need_mfa()
      } catch (err) {
        console.error(err)
        this.auth_error()
        throw err
      }
    },

    async refreshToken () {
      console.log('Refreshing access token')
      try {
        const res = await axios({ url: '/api/refresh_token', method: 'GET' })

        const token = res.data.token
        const renewable = res.data.renewable

        useConfigStore().axios.defaults.headers.common['X-Session-Token'] = token
        this.auth_success({ token, renewable })
        broadcastNewToken.postMessage({ token, renewable })
        return token
      } catch (err: any) {
        this.auth_error()
        let error
        if (err.response !== undefined && err.response.data !== undefined) {
          error = err.response.data.error
        } else {
          error = err.message
        }
        useQuasar().notify({
          color: 'red-4',
          textColor: 'black',
          icon: 'error',
          message: error
        }) 
        throw err
      }
    },

    async authorize ( otp: any ) {
      const res = await axios({ url: '/api/authorize', data: { otp: otp, state: this._stateToken }, method: 'POST' })
      const resState = res.data.state
      if (this._stateToken !== resState) {
        console.log('State token was malformed during request flow!')
       this.auth_error()
        throw new Error('State token was malformed during request flow!')
      }
      const token = res.data.token
      const authorized = res.data.authorized
      const renewable = res.data.renewable
      useConfigStore().axios.defaults.headers.common['X-Session-Token'] = token
      if (authorized) {
        this.auth_success({ token, renewable })
      }
    },

    async logout () {
      await useDesktopSessions().clearSessions()
      this.logout_mut()
      try {
        await useConfigStore().axios.post('/api/logout')
      } catch (err: any) {
        console.log(err)
        let error
        if (err.response !== undefined && err.response.data !== undefined) {
          error = err.response.data.error
        } else {
          error = err.message
        }
        useQuasar().notify({
          color: 'red-4',
          textColor: 'black',
          icon: 'error',
          message: error
        }) 
      }
      delete useConfigStore().axios.defaults.headers.common['X-Session-Token']
      window.location.href = '/#/login'
    }

  },

  getters: {
    isLoggedIn: (state) => !!state._token,
    requiresMFA: state => state._requiresMFA,
    authStatus: state => state._status,
    user: state => state._user,
    token: state => state._token,
    stateToken: state => state._stateToken,
    renewable: state => state._renewable
  }

})

