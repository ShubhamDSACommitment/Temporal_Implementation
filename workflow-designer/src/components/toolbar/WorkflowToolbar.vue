<script setup lang="ts">
import { ref } from 'vue'
import { useVueFlow } from '@vue-flow/core'
import { useWorkflowStore } from '../../stores/workflow'
import { useWorkflowExport } from '../../composables/useWorkflowExport'
import { useHistory } from '../../composables/useHistory'
import { useAutoLayout } from '../../composables/useAutoLayout'
import { saveDefinition, updateDefinition, startWorkflow, getWorkflowStatus } from '../../api/client'

const store = useWorkflowStore()
const { exportDefinition } = useWorkflowExport()
const { canUndo, canRedo, undo, redo } = useHistory()
const { applyLayout } = useAutoLayout()
const { viewport } = useVueFlow()

const showRunDialog = ref(false)
const runInput = ref('{\n  "order_id": "ORD-001",\n  "customer_id": "CUST-001",\n  "item": "Widget",\n  "quantity": 2,\n  "price": 49.99\n}')
const statusMessage = ref('')
const statusType = ref<'info' | 'success' | 'error'>('info')

function showStatus(msg: string, type: 'info' | 'success' | 'error' = 'info') {
  statusMessage.value = msg

  statusType.value = type
  setTimeout(() => { statusMessage.value = '' }, 5000)
}

async function handleSave() {
  const def = exportDefinition()
  if (def.steps.length === 0) {
    showStatus('Add at least one activity to the canvas', 'error')
    return
  }
  try {
    if (store.workflowId) {
      def.id = store.workflowId
      def.version = store.workflowVersion
      try {
        await updateDefinition(def)
        showStatus('Workflow updated!', 'success')
      } catch (putErr: any) {
        const msg: string = putErr.response?.data?.error || ''
        if (msg.includes('not found')) {
          // Record no longer exists in DB — create a new one
          def.id = ''
          store.workflowId = ''
          const saved = await saveDefinition(def)
          store.workflowId = saved.id
          showStatus('Workflow saved (re-created)!', 'success')
        } else {
          throw putErr
        }
      }
    } else {
      const saved = await saveDefinition(def)
      store.workflowId = saved.id
      showStatus('Workflow saved!', 'success')
    }
  } catch (e: any) {
    showStatus('Save failed: ' + (e.response?.data?.error || e.message), 'error')
  }
}

function handleExportJSON() {
  const def = exportDefinition()
  const blob = new Blob([JSON.stringify(def, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${def.name.replace(/\s+/g, '_')}.json`
  a.click()
  URL.revokeObjectURL(url)
}

function handleNew() {
  store.clearCanvas()
  showStatus('Canvas cleared', 'info')
}

function openRunDialog() {
  if (store.nodes.length === 0) {
    showStatus('Add activities before running', 'error')
    return
  }
  showRunDialog.value = true
}

async function handleRun() {
  showRunDialog.value = false
  let input: Record<string, any>
  try {
    input = JSON.parse(runInput.value)
  } catch {
    showStatus('Invalid JSON input', 'error')
    return
  }

  const def = exportDefinition()
  if (def.steps.length === 0) return
  try {
    // Always save latest canvas state before running
    let defId = store.workflowId
    if (defId) {
      def.id = defId
      def.version = store.workflowVersion
      await updateDefinition(def)
    } else {
      const saved = await saveDefinition(def)
      store.workflowId = saved.id
      defId = saved.id
    }

    showStatus('Starting workflow...', 'info')
    const result = await startWorkflow({ definition_id: store.workflowId, input })
    showStatus(`Workflow started: ${result.workflow_id}`, 'success')

    pollStatus(result.workflow_id)
  } catch (e: any) {
    showStatus('Run failed: ' + (e.response?.data?.error || e.message), 'error')
  }
}

async function pollStatus(workflowId: string) {
  for (let i = 0; i < 30; i++) {
    await new Promise((r) => setTimeout(r, 2000))
    try {
      const status = await getWorkflowStatus(workflowId)
      if (status.status === 'Completed') {
        showStatus(`Workflow completed! Result: ${JSON.stringify(status.result).substring(0, 100)}...`, 'success')
        return
      }
      if (status.status === 'Failed' || status.status === 'Terminated') {
        showStatus(`Workflow ${status.status}: ${status.error || 'unknown error'}`, 'error')
        return
      }
    } catch {
      // ignore polling errors
    }
  }
  showStatus('Workflow still running. Check Temporal UI for status.', 'info')
}

defineExpose({ handleSave })
</script>

<template>
  <div class="toolbar">
    <div class="toolbar-left">
      <input
        v-model="store.workflowName"
        class="workflow-name-input"
        placeholder="Workflow Name"
      />
    </div>

    <div class="toolbar-group">
      <button class="btn btn-icon" :disabled="!canUndo" title="Undo (Ctrl+Z)" @click="undo">
        <svg width="16" height="16" viewBox="0 0 16 16">
          <path d="M5 8L2 5L5 2" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
          <path d="M2 5H10C12.2 5 14 6.8 14 9C14 11.2 12.2 13 10 13H7" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
        </svg>
      </button>
      <button class="btn btn-icon" :disabled="!canRedo" title="Redo (Ctrl+Y)" @click="redo">
        <svg width="16" height="16" viewBox="0 0 16 16">
          <path d="M11 8L14 5L11 2" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
          <path d="M14 5H6C3.8 5 2 6.8 2 9C2 11.2 3.8 13 6 13H9" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
        </svg>
      </button>
      <div class="separator"></div>
      <button class="btn btn-icon" title="Auto Layout" @click="applyLayout">
        <svg width="16" height="16" viewBox="0 0 16 16">
          <rect x="1" y="6" width="4" height="4" rx="1" fill="none" stroke="currentColor" stroke-width="1.2" />
          <rect x="11" y="2" width="4" height="4" rx="1" fill="none" stroke="currentColor" stroke-width="1.2" />
          <rect x="11" y="10" width="4" height="4" rx="1" fill="none" stroke="currentColor" stroke-width="1.2" />
          <path d="M5 8H8V4H11M8 4V12H11" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" />
        </svg>
      </button>
      <div class="separator"></div>
      <span class="zoom-indicator">{{ Math.round(viewport.zoom * 100) }}%</span>
    </div>

    <div class="toolbar-actions">
      <button class="btn" @click="handleNew">New</button>
      <button class="btn btn-primary" @click="handleSave">Save</button>
      <button class="btn" @click="handleExportJSON">Export JSON</button>
      <button class="btn btn-run" @click="openRunDialog">Run</button>
    </div>
    <div v-if="statusMessage" class="status-bar" :class="statusType">
      {{ statusMessage }}
    </div>
  </div>

  <!-- Run Dialog -->
  <Teleport to="body">
    <div v-if="showRunDialog" class="dialog-overlay" @click.self="showRunDialog = false">
      <div class="dialog">
        <h3>Run Workflow</h3>
        <p class="dialog-hint">Enter workflow input as JSON:</p>
        <textarea v-model="runInput" rows="10" class="dialog-input"></textarea>
        <div class="dialog-actions">
          <button class="btn" @click="showRunDialog = false">Cancel</button>
          <button class="btn btn-run" @click="handleRun">Execute</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  background: #16213e;
  border-bottom: 1px solid #0f3460;
  gap: 16px;
  flex-wrap: wrap;
}

.toolbar-left {
  flex: 1;
  min-width: 150px;
}

.workflow-name-input {
  background: transparent;
  border: 1px solid transparent;
  color: #e0e0e0;
  font-size: 15px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 4px;
  width: 250px;
}

.workflow-name-input:focus {
  outline: none;
  border-color: #2a4060;
  background: #1a1a2e;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.separator {
  width: 1px;
  height: 20px;
  background: #2a4060;
  margin: 0 4px;
}

.zoom-indicator {
  font-size: 11px;
  color: #607080;
  min-width: 36px;
  text-align: center;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 6px 14px;
  background: #1e2a3a;
  border: 1px solid #2a4060;
  color: #c0c0d0;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.btn:hover:not(:disabled) {
  background: #2a3a4a;
  border-color: #3a5070;
}

.btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.btn-icon {
  padding: 4px 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-primary {
  background: #0f3460;
  border-color: #1a4a80;
  color: #ffffff;
}

.btn-primary:hover {
  background: #1a4a80;
}

.btn-run {
  background: #1a6030;
  border-color: #2a8040;
  color: #ffffff;
}

.btn-run:hover {
  background: #2a8040;
}

.status-bar {
  width: 100%;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
}

.status-bar.info {
  background: rgba(52, 152, 219, 0.15);
  color: #5dade2;
}

.status-bar.success {
  background: rgba(39, 174, 96, 0.15);
  color: #2ecc71;
}

.status-bar.error {
  background: rgba(231, 76, 60, 0.15);
  color: #e74c3c;
}

.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: #1e2a3a;
  border: 1px solid #2a4060;
  border-radius: 8px;
  padding: 24px;
  width: 500px;
  max-width: 90vw;
}

.dialog h3 {
  color: #ffffff;
  margin-bottom: 8px;
}

.dialog-hint {
  color: #8090a0;
  font-size: 13px;
  margin-bottom: 12px;
}

.dialog-input {
  width: 100%;
  padding: 8px;
  background: #141825;
  border: 1px solid #2a4060;
  border-radius: 4px;
  color: #e0e0e0;
  font-family: monospace;
  font-size: 13px;
  resize: vertical;
}

.dialog-input:focus {
  outline: none;
  border-color: #5dade2;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}
</style>
