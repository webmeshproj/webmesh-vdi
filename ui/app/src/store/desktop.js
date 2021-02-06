import Vue from 'vue'
import Vuex from 'vuex'

const equal = function (o1, o2) {
  return o1.name === o2.name && o1.namespace === o2.namespace
}

export const DesktopSessions = new Vuex.Store({

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

    async newSession ({ commit }, { template, namespace }) {
      try {
        const data = { template: template.metadata.name, namespace: namespace }
        const session = await Vue.prototype.$axios.post('/api/sessions', data)
        // add the socket type from the template config so we know how to connect
        // to the display
        if (typeof template.spec.config.socketType === 'string') {
          session.data.socketType = template.spec.config.socketType
        } else {
          session.data.socketType = 'xvnc'
        }
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
