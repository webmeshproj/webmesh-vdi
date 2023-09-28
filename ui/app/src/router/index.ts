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


import { createRouter, createWebHashHistory} from 'vue-router'
import {inject} from 'vue';
import MainLayout from '../layouts/MainLayout.vue'

import Login from '../pages/Login.vue'
import DesktopTemplates from '../pages/DesktopTemplates.vue'
import VNCViewer from '../pages/VNCViewer.vue'
import Settings from '../pages/Settings.vue'
import Profile from '../pages/Profile.vue'
import APIExplorer from '../pages/APIExplorer.vue'
import Metrics from '../pages/Metrics.vue'
import Error404 from '../pages/Error404.vue'
import  {useUserStore} from '../stores/user'
import { useConfigStore } from '../stores/config';

const routes = [
  {
    path: '/',
    redirect: '/templates',
    component: MainLayout,
    children: [
      {
        path: 'login',
        name: 'login',
        component: Login
      },
      {
        path: 'templates',
        name: 'templates',
        component: DesktopTemplates,
        meta: { requiresAuth: true }
      },
      {
        path: 'control',
        name: 'control',
        component: VNCViewer,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings',
        name: 'settings',
        component: Settings,
        meta: { requiresAuth: true }
      },
      {
        path: 'profile',
        name: 'profile',
        component: Profile,
        meta: { requiresAuth: true }
      },
      {
        path: 'swagger',
        name: 'swagger',
        component: APIExplorer
      },
      {
        path: 'metrics',
        name: 'metrics',
        component: Metrics
      },

      {  path: '/:catchAll(.*)*', name: 'bad-not-found', component: Error404 },
    ]
  }
]

/*
 * If not building with SSR mode, you can
 * directly export the Router instantiation;
 *
 * The function below can be async too; either use
 * async/await or return a Promise which resolves
 * with the Router instance.
 */


const router = createRouter({
     scrollBehavior: () => ({ left: 0, top: 0 }),
     routes,
     history: createWebHashHistory(),
 
     // Leave these as they are and change in quasar.conf.js instead!
     // quasar.conf.js -> build -> vueRouterMode
     // quasar.conf.js -> build -> publicPath
     /* mode: import.meta.env.VUE_ROUTER_MODE,
     base: import.meta.env.VUE_ROUTER_BASE */
   })
 
   router.beforeEach((to: any, from: any, next: any) => {


  const store = useUserStore()
     if (to.matched.some((record: any) => record.meta.requiresAuth)) {
       if (store.isLoggedIn) {
         next()
         return
       }
       next('/login')
     } else {
       next()
     }
   })
export default router