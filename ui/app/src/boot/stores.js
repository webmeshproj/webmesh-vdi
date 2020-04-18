import Vue from 'vue'

import UserStore from '../store/user.js'
import DesktopStore, { setTemplateBooted, templateIsBooted } from '../store/desktop.js'

UserStore.dispatch('initStore')

Vue.prototype.$userStore = UserStore
Vue.prototype.$desktopStore = DesktopStore
Vue.prototype.$setTemplateBooted = setTemplateBooted
Vue.prototype.$templateIsBooted = templateIsBooted
