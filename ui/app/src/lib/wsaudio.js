import { Howl, Howler } from 'howler'
import Vue from 'vue';
import Vuex from 'vuex';

Vue.use(Vuex)

const HowlerBank = new Vuex.Store({
  state: {
    playlist: [],
    running: false
  },

  mutations: {
    set_running (state, val) {
      state.running = val
    },
    new_howl (state, howl) {
      state.playlist.push(howl)
    }
  },

  actions: {
    addHowl ({ commit }, howl) {
      commit('new_howl', howl)
    },
    startHowls ({ commit, dispatch }) {
      commit('set_running', true)
      dispatch('runHowls')
    },
    stopHowls ({ commit }) {
      commit('set_running', false)
    },
    runHowls ({ state, dispatch, commit }) {
      if (!state.running) { return }
      if (state.playlist.length < 3) {
        setTimeout(() => { dispatch('runHowls') }, 0)
        return
      }
      console.log('playing sound')
      const sound = state.playlist.shift()
      sound.play()
      setTimeout(() => { dispatch('runHowls') }, sound.duration)
    }
  }

})

export default class {

  constructor (config) {
    this.config = config
    this.chunks = []
    this.sounds = []
    this.running = false
    this.locked = false
  }

  start () {
    this.running = true
    this.socket = new WebSocket(this.config.server.url)
    this.socket.onmessage = (message) => {
      if (message.data instanceof Blob) {
        while (this.locked) {}
        this.chunks.push(message.data)
      }
    }
    this.buildChunks()
    HowlerBank.dispatch('startHowls')
    // this.playChunks()
  }

  playChunks () {
    if (!this.running) { return }
    if (this.sounds.length < 3) {
      setTimeout(() => { this.playChunks() }, 0)
      return
    }
    console.log('playing sound')
    const sound = this.sounds.shift()
    sound.play()
    setTimeout(() => { this.playChunks() }, sound.duration)
  }

  buildChunks () {
    if (!this.running) { return }
    if (this.chunks.length < 5) {
      setTimeout(() => { this.buildChunks() }, 0)
      return
    }
    console.log('building chunks')
    this.locked = true
    var sound = new Howl({
      autoplay: false,
      src: [URL.createObjectURL(new Blob(this.chunks))],
      html5: true,
      preload: true,
      format: ['ogg', 'mp3']
    })
    this.chunks = []
    HowlerBank.dispatch('addHowl', sound)
    this.locked = false
    // this.sounds.push(sound)
    this.buildChunks()
  }

  stop () {
    this.socket.close()
    this.running = false
    HowlerBank.dispatch('stopHowls')
  }

}
