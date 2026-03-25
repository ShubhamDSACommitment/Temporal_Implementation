import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import DesignerView from './views/DesignerView.vue'
import WorkflowListView from './views/WorkflowListView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'designer', component: DesignerView },
    { path: '/workflows', name: 'workflows', component: WorkflowListView },
  ],
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
