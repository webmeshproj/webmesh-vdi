import Vue from 'vue'
import Vuex from 'vuex'

const equal = function (o1, o2) {
  return o1.name === o2.name && o1.namespace === o2.namespace
}

export const DesktopSessions = new Vuex.Store({

  state: {
    sessions: [],
    audioEnabled: false
  },

  mutations: {

    toggle_audio (state, data) {
      state.audioEnabled = data
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

    async newSession ({ commit }, { template, namespace }) {
      if (!Vue.prototype.$configStore.getters.localConfig.readWriteMany) {
        if (this.getters.sessions.length > 0) {
          throw Error('You cannot run two sessions while using persistence.\n\nTo override this behavior, go to Settings > Configuration -> Allow multiple sessions')
        }
      }
      try {
        const data = { template: template, namespace: namespace }
        const session = await Vue.prototype.$axios.post('/api/sessions', data)
        commit('new_session', session.data)
        commit('set_active_session', session.data)
      } catch (err) {
        console.log(`Failed to launch new session from ${template}`)
        console.error(err)
        throw err
      }
    },
    setActiveSession ({ commit }, data) {
      commit('set_active_session', data)
    },
    async deleteSession ({ commit }, data) {
      await Vue.prototype.$axios.delete(`/api/sessions/${data.namespace}/${data.name}`)
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
