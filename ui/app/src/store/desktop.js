import Vuex from 'vuex'

export const DesktopStore = new Vuex.Store({
  state: {},
  mutations: {

    SETBOOTED (state, template) {
      if (state[template] === undefined) {
        state[template] = { booted: true }
      } else {
        state[template].booted = true
      }
    },

    SETSHUTDOWN (state, template) {
      if (state[template] === undefined) {
        state[template] = { booted: false }
      } else {
        state[template].booted = false
      }
    }
  }
})

export const setTemplateBooted = function (template) {
  console.log(`Setting ${template} as booted`)
  DesktopStore.commit('SETBOOTED', template)
}

export const templateIsBooted = function (template) {
  if (DesktopStore.state[this.metadata.name] === undefined) {
    return false
  }
  return DesktopStore.state[this.metadata.name].booted === true
}

export default DesktopStore
