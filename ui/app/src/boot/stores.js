import Vue from 'vue'

import UserStore from '../store/user.js'
import DesktopSessions from '../store/desktop.js'

UserStore.dispatch('initStore')

Vue.prototype.$userStore = UserStore
Vue.prototype.$desktopSessions = DesktopSessions
