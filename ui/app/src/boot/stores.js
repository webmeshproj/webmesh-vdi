import Vue from 'vue'

import DesktopStore, { setTemplateBooted, templateIsBooted } from '../store/desktop.js'

Vue.prototype.$desktopStore = DesktopStore
Vue.prototype.$setTemplateBooted = setTemplateBooted
Vue.prototype.$templateIsBooted = templateIsBooted
