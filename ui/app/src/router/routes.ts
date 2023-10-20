import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/templates',
    component: () => import('layouts/MainLayout.vue'),
    children: [
        { 
            path: 'login',
            name: 'login',
            component: () => import('pages/Login.vue')
        },
        {
            path: 'templates',
            name: 'templates',
            component: () => import('pages/DesktopTemplates.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: 'control',
            name: 'control',
            component: () => import('pages/VNCViewer.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: 'settings',
            name: 'settings',
            component: () => import('pages/Settings.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: 'profile',
            name: 'profile',
            component: () => import('pages/Profile.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: 'swagger',
            name: 'swagger',
            component: () => import('pages/APIExplorer.vue'),
        },
        {
            path: 'metrics',
            name: 'metrics',
            component: () => import('pages/Metrics.vue'),
        },
    ],
  },

  // Always leave this as last one,
  // but you can also remove it
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/Error404.vue'),
  },
];

export default routes;
