<script setup lang="ts">
import { computed } from 'vue'
import type { Edge } from '@vue-flow/core'
import { useWorkflowStore } from '../../stores/workflow'
import { useHistory } from '../../composables/useHistory'

const props = defineProps<{ edge: Edge }>()

const store = useWorkflowStore()
const { pushHistory } = useHistory()

const sourceNode = computed(() =>
  store.nodes.find((n) => n.id === props.edge.source)
)

const targetNode = computed(() =>
  store.nodes.find((n) => n.id === props.edge.target)
)

const sourceLabel = computed(() =>
  sourceNode.value ? (sourceNode.value.data as any).label || sourceNode.value.id : props.edge.source
)

const targetLabel = computed(() =>
  targetNode.value ? (targetNode.value.data as any).label || targetNode.value.id : props.edge.target
)

const isGatewayEdge = computed(() =>
  sourceNode.value?.type === 'exclusiveGateway'
)

const condition = computed(() =>
  (props.edge as any).data?.condition || { expression: '', isDefault: false }
)

function onExpressionChange(expression: string) {
  store.updateEdgeData(props.edge.id, {
    condition: { ...condition.value, expression },
    label: expression || (condition.value.isDefault ? '(default)' : ''),
  })
  pushHistory()
}

function onDefaultChange(checked: boolean) {
  if (checked) {
    store.setDefaultEdge(props.edge.source, props.edge.id)
  } else {
    store.updateEdgeData(props.edge.id, {
      condition: { ...condition.value, isDefault: false },
      label: condition.value.expression || '',
    })
  }
  pushHistory()
}
</script>

<template>
  <div class="edge-properties">
    <div class="section">
      <div class="section-title">Sequence Flow</div>
      <div class="flow-row">
        <span class="flow-label">Source:</span>
        <span class="badge">{{ sourceLabel }}</span>
        <span class="flow-arrow">&rarr;</span>
        <span class="flow-label">Target:</span>
        <span class="badge">{{ targetLabel }}</span>
      </div>
    </div>

    <div class="section" v-if="isGatewayEdge">
      <div class="section-title">Condition</div>
      <label class="field-label">Expression</label>
      <textarea
        class="field-textarea"
        :value="condition.expression"
        :disabled="condition.isDefault"
        placeholder='e.g. $steps.checkOrder.amount > 100'
        @input="onExpressionChange(($event.target as HTMLTextAreaElement).value)"
      />
      <label class="checkbox-label">
        <input
          type="checkbox"
          :checked="condition.isDefault"
          @change="onDefaultChange(($event.target as HTMLInputElement).checked)"
        />
        Default path
      </label>
    </div>

    <div class="section" v-else>
      <div class="info-note">
        Conditions are only available on gateway outgoing edges.
      </div>
    </div>
  </div>
</template>

<style scoped>
.edge-properties {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.section-title {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: #8090a0;
  letter-spacing: 0.5px;
  margin-bottom: 4px;
}

.flow-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.flow-label {
  font-size: 12px;
  color: #8090a0;
}

.flow-arrow {
  color: #607080;
  font-size: 14px;
}

.badge {
  display: inline-block;
  background: #2a4060;
  color: #c0c0d0;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
}

.field-label {
  font-size: 12px;
  color: #8090a0;
}

.field-textarea {
  padding: 6px 8px;
  background: #1e2a3a;
  border: 1px solid #2a4060;
  border-radius: 4px;
  color: #e0e0e0;
  font-size: 12px;
  font-family: monospace;
  resize: vertical;
  min-height: 40px;
}

.field-textarea:focus {
  outline: none;
  border-color: #5dade2;
}

.field-textarea:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #c0c0d0;
  cursor: pointer;
}

.info-note {
  font-size: 12px;
  color: #607080;
  font-style: italic;
}
</style>
