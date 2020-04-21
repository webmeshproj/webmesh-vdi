import Vue from 'vue'

import UserStore from '../store/user.js'
import DesktopSessions from '../store/desktop.js'

Vue.prototype.$userStore = UserStore
Vue.prototype.$desktopSessions = DesktopSessions
