<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listDefinitions, deleteDefinition, getDefinition } from '../api/client'
import { useWorkflowStore } from '../stores/workflow'
import type { WorkflowDefinition } from '../types/workflow'

const router = useRouter()
const store = useWorkflowStore()
const definitions = ref<WorkflowDefinition[]>([])
const loading = ref(true)
const error = ref('')

async function loadDefinitions() {
  loading.value = true
  error.value = ''
  try {
    definitions.value = await listDefinitions()
  } catch (e: any) {
    error.value = e.response?.data?.error || e.message
  } finally {
    loading.value = false
  }
}

async function openDefinition(id: string) {
  try {
    const def = await getDefinition(id)
    store.clearCanvas()
    store.workflowId = def.id
    store.workflowName = def.name
    store.workflowDescription = def.description
    store.workflowVersion = def.version
    store.loadFromSteps(def.steps, def.events, def.gateways, def.edges)
    router.push('/')
  } catch (e: any) {
    error.value = 'Failed to load: ' + (e.response?.data?.error || e.message)
  }
}

async function handleDelete(id: string) {
  if (!confirm('Delete this workflow definition?')) return
  try {
    await deleteDefinition(id)
    await loadDefinitions()
  } catch (e: any) {
    error.value = 'Delete failed: ' + (e.response?.data?.error || e.message)
  }
}

onMounted(loadDefinitions)
</script>

<template>
  <div class="list-view">
    <div class="list-header">
      <h2>Saved Workflows</h2>
      <button class="btn" @click="loadDefinitions">Refresh</button>
    </div>

    <div v-if="error" class="error-msg">{{ error }}</div>

    <div v-if="loading" class="loading">Loading...</div>

    <div v-else-if="definitions.length === 0" class="empty">
      No saved workflows yet. Go to the Designer to create one.
    </div>

    <div v-else class="list-grid">
      <div v-for="def in definitions" :key="def.id" class="list-card">
        <div class="card-header">
          <h3>{{ def.name }}</h3>
          <span class="card-version">v{{ def.version }}</span>
        </div>
        <p class="card-desc">{{ def.description || 'No description' }}</p>
        <p class="card-steps">{{ def.steps?.length || 0 }} steps</p>
        <div class="card-actions">
          <button class="btn btn-sm" @click="openDefinition(def.id)">Open</button>
          <button class="btn btn-sm btn-danger" @click="handleDelete(def.id)">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.list-view {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.list-header h2 {
  color: #e0e0e0;
  font-size: 20px;
}

.btn {
  padding: 6px 14px;
  background: #1e2a3a;
  border: 1px solid #2a4060;
  color: #c0c0d0;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.btn:hover {
  background: #2a3a4a;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 12px;
}

.btn-danger {
  border-color: #c0392b;
  color: #e74c3c;
}

.btn-danger:hover {
  background: rgba(231, 76, 60, 0.15);
}

.error-msg {
  background: rgba(231, 76, 60, 0.15);
  color: #e74c3c;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 16px;
}

.loading, .empty {
  color: #607080;
  text-align: center;
  padding: 40px;
  font-size: 14px;
}

.list-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.list-card {
  background: #1e2a3a;
  border: 1px solid #2a4060;
  border-radius: 8px;
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.card-header h3 {
  color: #ffffff;
  font-size: 15px;
}

.card-version {
  font-size: 11px;
  color: #607080;
  background: #141825;
  padding: 2px 6px;
  border-radius: 3px;
}

.card-desc {
  color: #8090a0;
  font-size: 13px;
  margin-bottom: 4px;
}

.card-steps {
  color: #607080;
  font-size: 12px;
  margin-bottom: 12px;
}

.card-actions {
  display: flex;
  gap: 8px;
}
</style>
