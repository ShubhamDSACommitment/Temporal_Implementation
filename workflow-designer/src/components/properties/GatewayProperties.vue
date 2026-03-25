<script setup lang="ts">
import { computed } from 'vue'
import { useWorkflowStore, type DesignerNode } from '../../stores/workflow'
import { useHistory } from '../../composables/useHistory'

const props = defineProps<{ node: DesignerNode }>()

const store = useWorkflowStore()
const { pushHistory } = useHistory()

const displayName = computed({
  get: () => props.node.data.label,
  set: (val: string) => {
    store.updateNodeData(props.node.id, { label: val })
    pushHistory()
  },
})

const outgoingEdges = computed(() =>
  store.edges.filter((e) => e.source === props.node.id)
)

function getTargetLabel(targetId: string): string {
  const node = store.nodes.find((n) => n.id === targetId)
  return node ? (node.data as any).label || node.id : targetId
}

function getCondition(edge: any) {
  return edge.data?.condition || { expression: '', isDefault: false }
}

function onExpressionChange(edgeId: string, expression: string) {
  const edge = store.edges.find((e) => e.id === edgeId) as any
  const existing = edge?.data?.condition || { expression: '', isDefault: false }
  store.updateEdgeData(edgeId, {
    condition: { ...existing, expression },
    label: expression || (existing.isDefault ? '(default)' : ''),
  })
  pushHistory()
}

function onDefaultChange(edgeId: string, checked: boolean) {
  if (checked) {
    store.setDefaultEdge(props.node.id, edgeId)
  } else {
    const edge = store.edges.find((e) => e.id === edgeId) as any
    const existing = edge?.data?.condition || { expression: '', isDefault: false }
    store.updateEdgeData(edgeId, {
      condition: { ...existing, isDefault: false },
      label: existing.expression || '',
    })
  }
  pushHistory()
}
</script>

<template>
  <div class="gateway-properties">
    <div class="section">
      <div class="section-title">General</div>
      <label class="field-label">Display Name</label>
      <input v-model="displayName" class="field-input" type="text" />
    </div>

    <div class="section">
      <div class="section-title">Outgoing Paths</div>
      <div v-if="outgoingEdges.length === 0" class="empty-hint">
        No outgoing edges. Connect this gateway to other nodes.
      </div>
      <div v-for="edge in outgoingEdges" :key="edge.id" class="edge-item">
        <div class="edge-target">
          <span class="badge">{{ getTargetLabel(edge.target) }}</span>
        </div>
        <label class="field-label">Condition Expression</label>
        <textarea
          class="field-textarea"
          :value="getCondition(edge).expression"
          :disabled="getCondition(edge).isDefault"
          placeholder='e.g. $steps.checkOrder.amount > 100'
          @input="onExpressionChange(edge.id, ($event.target as HTMLTextAreaElement).value)"
        />
        <label class="checkbox-label">
          <input
            type="checkbox"
            :checked="getCondition(edge).isDefault"
            @change="onDefaultChange(edge.id, ($event.target as HTMLInputElement).checked)"
          />
          Default path
        </label>
      </div>
    </div>
  </div>
</template>

<style scoped>
.gateway-properties {
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

.field-label {
  font-size: 12px;
  color: #8090a0;
}

.field-input {
  padding: 6px 8px;
  background: #141825;
  border: 1px solid #2a4060;
  border-radius: 4px;
  color: #e0e0e0;
  font-size: 13px;
}

.field-input:focus {
  outline: none;
  border-color: #5dade2;
}

.edge-item {
  background: #141825;
  border: 1px solid #2a4060;
  border-radius: 6px;
  padding: 10px;
  margin-bottom: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.edge-target {
  margin-bottom: 4px;
}

.badge {
  display: inline-block;
  background: #2a4060;
  color: #c0c0d0;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
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

.empty-hint {
  font-size: 12px;
  color: #607080;
  font-style: italic;
}
</style>
