import { createRouter, createWebHistory } from 'vue-router'
import MonitorView from '../views/MonitorView.vue'
import ControlView from '../views/ControlView.vue'
import SettingsView from '../views/SettingsView.vue'

const routes = [
  { path: '/', redirect: '/monitor' },
  { path: '/monitor', name: 'Monitor', component: MonitorView },
  { path: '/control', name: 'Control', component: ControlView },
  { path: '/settings', name: 'Settings', component: SettingsView },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
