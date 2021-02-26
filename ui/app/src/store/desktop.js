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

const equal = function (o1, o2) {
  return o1.name === o2.name && o1.namespace === o2.namespace
}

const DesktopSessions = new Vuex.Store({

  state: {
    sessions: [], // deciding against local storage here, but it is still an option
    audioEnabled: false,
    recordingEnabled: false
  },

  mutations: {

    toggle_audio (state, data) {
      state.audioEnabled = data
    },

    toggle_recording (state, data) {
      state.recordingEnabled = data
    },

    new_session (state, data) {
      data.active = true
      state.sessions.push(data)
    },

    set_active_session (state, data) {
      const newSessions = []
      state.sessions.forEach((val) => {
        if (equal(val, data)) {
          val.active = true
        } else {
          val.active = false
        }
        newSessions.push(val)
      })
      state.sessions = newSessions
    },

    delete_session (state, data) {
      state.sessions = state.sessions.filter((val) => {
        return !equal(val, data)
      })
      if (state.sessions.length !== 0) {
        state.sessions[0].active = true
      }
    }

  },

  actions: {
    toggleAudio ({ commit }, data) {
      commit('toggle_audio', data)
    },

    toggleRecording ({ commit }, data) {
      commit('toggle_recording', data)
    },

    addExistingSession ({ commit }, data) {
      commit('new_session', data)
    },

    async newSession ({ commit }, { template, namespace, serviceAccount }) {
      try {
        const data = { template: template.metadata.name, namespace: namespace }
        if (serviceAccount) {
          data.serviceAccount = serviceAccount
        }
        const session = await Vue.prototype.$axios.post('/api/sessions', data)
        session.data.template = template
        commit('new_session', session.data)
        commit('set_active_session', session.data)
      } catch (err) {
        console.log(`Failed to launch new session from ${template.metadata.name}`)
        console.error(err)
        throw err
      }
    },

    setActiveSession ({ commit }, data) {
      commit('set_active_session', data)
    },

    deleteSessionOffline ({ commit }, data) {
      commit('delete_session', data)
    },

    async deleteSession ({ commit }, data) {
      try {
        await Vue.prototype.$axios.delete(`/api/sessions/${data.namespace}/${data.name}`)
      } catch (err) {
        console.log(`Error sending delete to API: ${err}`)
      }
      commit('delete_session', data)
    },

    async clearSessions ({ commit }) {
      this.getters.sessions.forEach(async (session) => {
        await this.dispatch('deleteSession', session)
      })
    }

  },

  getters: {
    sessions: state => state.sessions,
    activeSession: state => state.sessions.filter(sess => sess.active)[0],
    audioEnabled: state => state.audioEnabled,
    recordingEnabled: state => state.recordingEnabled,
    sessionStatus: (state) => async (data) => {
      try {
        const res = await Vue.prototype.$axios.get(
          `/api/sessions/${data.namespace}/${data.name}`
        )
        return res.data
      } catch (err) {
        console.log(`Failed to fetch session status for ${data.namespace}/${data.name}`)
        console.error(err)
        throw err
      }
    }
  }

})

export default DesktopSessions
