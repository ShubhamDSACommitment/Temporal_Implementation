<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useWorkflowStore, type ActivityNodeData } from '../../stores/workflow'
import { useHistory } from '../../composables/useHistory'
import { useActivityRegistry } from '../../composables/useActivityRegistry'
import type { DesignerNode } from '../../stores/workflow'

const props = defineProps<{ node: DesignerNode }>()

const store = useWorkflowStore()
const { pushHistory } = useHistory()
const { categories: registryCategories } = useActivityRegistry()

const currentInputSchema = computed(() => {
  const name = activityName.value
  if (!name) return []
  for (const cat of registryCategories.value) {
    for (const item of cat.items) {
      if (item.activityName === name && item.inputSchema) {
        return item.inputSchema
      }
    }
  }
  return []
})

const availableTaskQueues = computed(() => {
  const queues = new Set<string>()
  for (const cat of registryCategories.value) {
    for (const item of cat.items) {
      if (item.defaultTaskQueue) queues.add(item.defaultTaskQueue)
    }
  }
  if (taskQueue.value) queues.add(taskQueue.value)
  return Array.from(queues).sort()
})

const activeTab = ref<'general' | 'config' | 'io'>('general')

const label = ref('')
const activityName = ref('')
const taskQueue = ref('')
const timeoutSeconds = ref(30)
const maxAttempts = ref(3)
const initialIntervalSec = ref(1)
const backoffCoefficient = ref(2.0)
const inputRows = ref<{ key: string; value: string }[]>([])

watch(() => props.node, (n) => {
  if (!n) return
  label.value = n.data.label
  activeTab.value = 'general'

  const d = n.data as ActivityNodeData
  activityName.value = d.activityName
  taskQueue.value = d.taskQueue
  timeoutSeconds.value = d.timeoutSeconds
  maxAttempts.value = d.retryPolicy.maxAttempts
  initialIntervalSec.value = d.retryPolicy.initialIntervalSec
  backoffCoefficient.value = d.retryPolicy.backoffCoefficient
  inputRows.value = Object.entries(d.inputMapping || {}).map(([key, value]) => ({ key, value }))
  if (inputRows.value.length === 0) {
    inputRows.value = [{ key: '', value: '' }]
  }
}, { immediate: true })

function applyChanges() {
  const mapping: Record<string, string> = {}
  for (const row of inputRows.value) {
    const k = row.key.trim()
    const v = row.value.trim()
    if (k && v) mapping[k] = v
  }

  store.updateNodeData(props.node.id, {
    label: label.value,
    activityName: activityName.value,
    taskQueue: taskQueue.value,
    timeoutSeconds: timeoutSeconds.value,
    retryPolicy: {
      maxAttempts: maxAttempts.value,
      initialIntervalSec: initialIntervalSec.value,
      backoffCoefficient: backoffCoefficient.value,
    },
    inputMapping: mapping,
  })
  pushHistory()
}

function addRow() {
  inputRows.value.push({ key: '', value: '' })
}

function removeRow(idx: number) {
  inputRows.value.splice(idx, 1)
  applyChanges()
}
</script>

<template>
  <div class="tabs">
    <button
      class="tab"
      :class="{ active: activeTab === 'general' }"
      @click="activeTab = 'general'"
    >General</button>
    <button
      class="tab"
      :class="{ active: activeTab === 'config' }"
      @click="activeTab = 'config'"
    >Configuration</button>
    <button
      class="tab"
      :class="{ active: activeTab === 'io' }"
      @click="activeTab = 'io'"
    >Input/Output</button>
  </div>

  <!-- General tab -->
  <div v-show="activeTab === 'general'" class="tab-content">
    <div class="field">
      <label>Display Name</label>
      <input v-model="label" @change="applyChanges" />
    </div>
    <div class="field">
      <label>Activity Name</label>
      <input v-model="activityName" @change="applyChanges" />
    </div>
    <div class="field">
      <label>Task Queue</label>
      <select v-model="taskQueue" @change="applyChanges">
        <option v-for="q in availableTaskQueues" :key="q" :value="q">{{ q }}</option>
      </select>
    </div>
  </div>

  <!-- Configuration tab -->
  <div v-show="activeTab === 'config'" class="tab-content">
    <div class="field">
      <label>Timeout (seconds)</label>
      <input type="number" v-model.number="timeoutSeconds" @change="applyChanges" />
    </div>

    <h4 class="section-title">Retry Policy</h4>

    <div class="field">
      <label>Max Attempts</label>
      <input type="number" v-model.number="maxAttempts" @change="applyChanges" />
    </div>
    <div class="field">
      <label>Initial Interval (sec)</label>
      <input type="number" v-model.number="initialIntervalSec" @change="applyChanges" />
    </div>
    <div class="field">
      <label>Backoff Coefficient</label>
      <input type="number" step="0.1" v-model.number="backoffCoefficient" @change="applyChanges" />
    </div>
  </div>

  <!-- Input/Output tab -->
  <div v-show="activeTab === 'io'" class="tab-content">
    <div v-if="currentInputSchema.length" class="schema-hint">
      <h4 class="section-title">Expected Fields</h4>
      <div v-for="field in currentInputSchema" :key="field.name" class="schema-field">
        <span class="schema-name">{{ field.name }}</span>
        <span class="schema-type">{{ field.type }}</span>
        <span v-if="field.required" class="schema-required">required</span>
      </div>
    </div>
    <h4 class="section-title">Input Mapping</h4>
    <div class="io-table">
      <div class="io-header">
        <span>Key</span>
        <span>Value</span>
        <span></span>
      </div>
      <div v-for="(row, idx) in inputRows" :key="idx" class="io-row">
        <input v-model="row.key" placeholder="key" @change="applyChanges" />
        <input v-model="row.value" placeholder="$input.field" @change="applyChanges" />
        <button class="io-delete" @click="removeRow(idx)" title="Remove row">
          <svg width="12" height="12" viewBox="0 0 12 12">
            <line x1="2" y1="2" x2="10" y2="10" stroke="currentColor" stroke-width="1.5" />
            <line x1="10" y1="2" x2="2" y2="10" stroke="currentColor" stroke-width="1.5" />
          </svg>
        </button>
      </div>
    </div>
    <button class="btn-add-row" @click="addRow">+ Add Row</button>
  </div>
</template>

<style scoped>
.tabs {
  display: flex;
  gap: 0;
  margin-bottom: 16px;
  border-bottom: 1px solid #2a4060;
}

.tab {
  flex: 1;
  padding: 6px 4px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: #607080;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.tab:hover {
  color: #c0c0d0;
}

.tab.active {
  color: #5dade2;
  border-bottom-color: #5dade2;
}

.tab-content {
  margin-bottom: 8px;
}

.section-title {
  font-size: 12px;
  font-weight: 600;
  color: #8090a0;
  margin-top: 16px;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.field {
  margin-bottom: 10px;
}

.field label {
  display: block;
  font-size: 11px;
  color: #8090a0;
  margin-bottom: 4px;
}

.field input,
.field textarea,
.field select {
  width: 100%;
  padding: 6px 8px;
  background: #1a1a2e;
  border: 1px solid #2a4060;
  border-radius: 4px;
  color: #e0e0e0;
  font-size: 13px;
  font-family: inherit;
}

.field input:focus,
.field textarea:focus,
.field select:focus {
  outline: none;
  border-color: #5dade2;
}

.field select {
  cursor: pointer;
}

.field select option {
  background: #1a1a2e;
  color: #e0e0e0;
}

.io-table {
  margin-bottom: 8px;
}

.io-header {
  display: grid;
  grid-template-columns: 1fr 1fr 24px;
  gap: 4px;
  margin-bottom: 4px;
  font-size: 10px;
  color: #607080;
  text-transform: uppercase;
}

.io-row {
  display: grid;
  grid-template-columns: 1fr 1fr 24px;
  gap: 4px;
  margin-bottom: 4px;
}

.io-row input {
  padding: 4px 6px;
  background: #1a1a2e;
  border: 1px solid #2a4060;
  border-radius: 3px;
  color: #e0e0e0;
  font-size: 11px;
  font-family: monospace;
}

.io-row input:focus {
  outline: none;
  border-color: #5dade2;
}

.io-delete {
  width: 24px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: #607080;
  cursor: pointer;
  border-radius: 3px;
}

.io-delete:hover {
  background: rgba(231, 76, 60, 0.15);
  color: #e74c3c;
}

.btn-add-row {
  width: 100%;
  padding: 4px;
  background: transparent;
  border: 1px dashed #2a4060;
  color: #607080;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  transition: all 0.2s;
}

.btn-add-row:hover {
  border-color: #5dade2;
  color: #5dade2;
}

.schema-hint {
  margin-bottom: 12px;
  padding: 8px;
  background: rgba(93, 173, 226, 0.05);
  border: 1px solid #2a4060;
  border-radius: 4px;
}

.schema-field {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 2px 0;
  font-size: 11px;
}

.schema-name {
  color: #c0c0d0;
  font-family: monospace;
}

.schema-type {
  color: #607080;
  font-size: 10px;
}

.schema-required {
  color: #e67e22;
  font-size: 9px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
</style>
