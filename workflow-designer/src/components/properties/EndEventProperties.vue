<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWorkflowStore, type EventNodeData } from '../../stores/workflow'
import { useHistory } from '../../composables/useHistory'
import type { DesignerNode } from '../../stores/workflow'
import type { EndEventConfig, EndEventType, OutputVariable } from '../../types/workflow'

const props = defineProps<{ node: DesignerNode }>()

const store = useWorkflowStore()
const { pushHistory } = useHistory()

const label = ref('')
const endType = ref<EndEventType>('none')
const errorCode = ref('')
const errorMessage = ref('')
const outputVariables = ref<OutputVariable[]>([])

watch(() => props.node, (n) => {
  if (!n) return
  const data = n.data as EventNodeData
  label.value = data.label
  const cfg = data.endConfig ?? { endType: 'none' as EndEventType, outputVariables: [] }
  endType.value = cfg.endType
  errorCode.value = cfg.error?.errorCode ?? ''
  errorMessage.value = cfg.error?.errorMessage ?? ''
  outputVariables.value = cfg.outputVariables ? [...cfg.outputVariables] : []
}, { immediate: true })

function applyChanges() {
  const config: EndEventConfig = {
    endType: endType.value,
    outputVariables: outputVariables.value,
  }

  if (endType.value === 'error') {
    config.error = {
      errorCode: errorCode.value,
      errorMessage: errorMessage.value,
    }
  }

  store.updateNodeData(props.node.id, { label: label.value, endConfig: config })
  pushHistory()
}

function addOutputVariable() {
  outputVariables.value.push({ name: '', expression: '' })
}

function removeOutputVariable(idx: number) {
  outputVariables.value.splice(idx, 1)
  applyChanges()
}
</script>

<template>
  <div class="tab-content">
    <h4 class="section-title">General</h4>
    <div class="field">
      <label>Display Name</label>
      <input v-model="label" @change="applyChanges" />
    </div>

    <h4 class="section-title">End Type</h4>
    <div class="field">
      <label>Type</label>
      <select v-model="endType" @change="applyChanges">
        <option value="none">None</option>
        <option value="error">Error</option>
        <option value="terminate">Terminate</option>
      </select>
    </div>

    <!-- Error config -->
    <template v-if="endType === 'error'">
      <div class="field">
        <label>Error Code</label>
        <input v-model="errorCode" placeholder="ERR_001" @change="applyChanges" />
      </div>
      <div class="field">
        <label>Error Message</label>
        <textarea v-model="errorMessage" rows="3" placeholder="Describe the error..." @change="applyChanges" />
      </div>
    </template>

    <h4 class="section-title">Output Variables</h4>
    <div class="io-table">
      <div class="io-header">
        <span>Name</span>
        <span>Expression</span>
        <span></span>
      </div>
      <div v-for="(ov, idx) in outputVariables" :key="idx" class="io-row">
        <input v-model="ov.name" placeholder="variableName" @change="applyChanges" />
        <input v-model="ov.expression" placeholder="$steps.stepId.field" @change="applyChanges" />
        <button class="io-delete" @click="removeOutputVariable(idx)" title="Remove variable">
          <svg width="12" height="12" viewBox="0 0 12 12">
            <line x1="2" y1="2" x2="10" y2="10" stroke="currentColor" stroke-width="1.5" />
            <line x1="10" y1="2" x2="2" y2="10" stroke="currentColor" stroke-width="1.5" />
          </svg>
        </button>
      </div>
    </div>
    <button class="btn-add-row" @click="addOutputVariable">+ Add Variable</button>
  </div>
</template>

<style scoped>
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

.field textarea {
  resize: vertical;
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
</style>
