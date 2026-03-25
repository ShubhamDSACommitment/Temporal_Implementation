<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWorkflowStore, type EventNodeData } from '../../stores/workflow'
import { useHistory } from '../../composables/useHistory'
import type { DesignerNode } from '../../stores/workflow'
import type { StartEventConfig, StartTriggerType, FormField } from '../../types/workflow'

const props = defineProps<{ node: DesignerNode }>()

const store = useWorkflowStore()
const { pushHistory } = useHistory()

const label = ref('')
const triggerType = ref<StartTriggerType>('manual')
const timerType = ref<'cron' | 'delay'>('cron')
const cronExpression = ref('')
const delayDuration = ref('')
const endpointPath = ref('')
const eventName = ref('')
const formFields = ref<FormField[]>([])

watch(() => props.node, (n) => {
  if (!n) return
  const data = n.data as EventNodeData
  label.value = data.label
  const cfg = data.startConfig ?? { triggerType: 'manual' as StartTriggerType, formFields: [] }
  triggerType.value = cfg.triggerType
  timerType.value = cfg.timer?.type ?? 'cron'
  cronExpression.value = cfg.timer?.cron ?? ''
  delayDuration.value = cfg.timer?.delay ?? ''
  endpointPath.value = cfg.webhook?.endpointPath ?? ''
  eventName.value = cfg.webhook?.eventName ?? ''
  formFields.value = cfg.formFields ? [...cfg.formFields] : []
}, { immediate: true })

function applyChanges() {
  const config: StartEventConfig = {
    triggerType: triggerType.value,
    formFields: formFields.value,
  }

  if (triggerType.value === 'timer') {
    config.timer = {
      type: timerType.value,
      cron: timerType.value === 'cron' ? cronExpression.value : undefined,
      delay: timerType.value === 'delay' ? delayDuration.value : undefined,
    }
  }

  if (triggerType.value === 'webhook') {
    config.webhook = {
      endpointPath: endpointPath.value,
      eventName: eventName.value,
    }
  }

  store.updateNodeData(props.node.id, { label: label.value, startConfig: config })
  pushHistory()
}

function addFormField() {
  formFields.value.push({ name: '', type: 'string', label: '', required: false, defaultValue: '' })
}

function removeFormField(idx: number) {
  formFields.value.splice(idx, 1)
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

    <h4 class="section-title">Trigger</h4>
    <div class="field">
      <label>Type</label>
      <select v-model="triggerType" @change="applyChanges">
        <option value="manual">Manual</option>
        <option value="timer">Timer</option>
        <option value="webhook">Webhook</option>
      </select>
    </div>

    <!-- Timer config -->
    <template v-if="triggerType === 'timer'">
      <div class="field">
        <label>Timer Type</label>
        <select v-model="timerType" @change="applyChanges">
          <option value="cron">Cron</option>
          <option value="delay">Delay</option>
        </select>
      </div>
      <div v-if="timerType === 'cron'" class="field">
        <label>Cron Expression</label>
        <input v-model="cronExpression" placeholder="0 */5 * * *" @change="applyChanges" />
      </div>
      <div v-if="timerType === 'delay'" class="field">
        <label>Delay Duration</label>
        <input v-model="delayDuration" placeholder="30s, 5m, 1h" @change="applyChanges" />
      </div>
    </template>

    <!-- Webhook config -->
    <template v-if="triggerType === 'webhook'">
      <div class="field">
        <label>Endpoint Path</label>
        <input v-model="endpointPath" placeholder="/hooks/my-event" @change="applyChanges" />
      </div>
      <div class="field">
        <label>Event Name</label>
        <input v-model="eventName" placeholder="order.created" @change="applyChanges" />
      </div>
    </template>

    <h4 class="section-title">Form Fields</h4>
    <div class="io-table">
      <div class="io-header">
        <span>Name</span>
        <span>Type</span>
        <span>Label</span>
        <span>Req</span>
        <span>Default</span>
        <span></span>
      </div>
      <div v-for="(ff, idx) in formFields" :key="idx" class="io-row form-row">
        <input v-model="ff.name" placeholder="name" @change="applyChanges" />
        <select v-model="ff.type" @change="applyChanges">
          <option value="string">string</option>
          <option value="number">number</option>
          <option value="boolean">boolean</option>
          <option value="json">json</option>
        </select>
        <input v-model="ff.label" placeholder="label" @change="applyChanges" />
        <input type="checkbox" v-model="ff.required" @change="applyChanges" />
        <input v-model="ff.defaultValue" placeholder="default" @change="applyChanges" />
        <button class="io-delete" @click="removeFormField(idx)" title="Remove field">
          <svg width="12" height="12" viewBox="0 0 12 12">
            <line x1="2" y1="2" x2="10" y2="10" stroke="currentColor" stroke-width="1.5" />
            <line x1="10" y1="2" x2="2" y2="10" stroke="currentColor" stroke-width="1.5" />
          </svg>
        </button>
      </div>
    </div>
    <button class="btn-add-row" @click="addFormField">+ Add Field</button>
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

.io-table {
  margin-bottom: 8px;
}

.io-header {
  display: grid;
  grid-template-columns: 1fr 70px 1fr 28px 1fr 24px;
  gap: 4px;
  margin-bottom: 4px;
  font-size: 10px;
  color: #607080;
  text-transform: uppercase;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 70px 1fr 28px 1fr 24px;
  gap: 4px;
  margin-bottom: 4px;
}

.form-row input,
.form-row select {
  padding: 4px 6px;
  background: #1a1a2e;
  border: 1px solid #2a4060;
  border-radius: 3px;
  color: #e0e0e0;
  font-size: 11px;
  font-family: monospace;
}

.form-row input:focus,
.form-row select:focus {
  outline: none;
  border-color: #5dade2;
}

.form-row input[type="checkbox"] {
  width: auto;
  margin: auto;
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
