import MainLayout from 'layouts/MainLayout.vue'
import DesktopTemplates from 'pages/DesktopTemplates.vue'
// import VNCIframe from 'pages/VNCIframe.vue'
import VNCViewer from 'pages/VNCViewer.vue'

import Error404 from 'pages/Error404.vue'

const routes = [
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', component: DesktopTemplates },
      { path: 'templates', component: DesktopTemplates },
      { path: 'vnc', component: VNCViewer },
      { path: '*', component: Error404 }
    ]
  }
]

// Always leave this as last one
if (process.env.MODE !== 'ssr') {
  routes.push({
    path: '*',
    component: () => import('pages/Error404.vue')
  })
}

export default routes
