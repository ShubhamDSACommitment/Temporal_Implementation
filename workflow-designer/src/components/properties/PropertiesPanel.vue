<script setup lang="ts">
import { computed } from 'vue'
import { useWorkflowStore } from '../../stores/workflow'
import { useHistory } from '../../composables/useHistory'
import ActivityProperties from './ActivityProperties.vue'
import StartEventProperties from './StartEventProperties.vue'
import EndEventProperties from './EndEventProperties.vue'
import GatewayProperties from './GatewayProperties.vue'
import EdgeProperties from './EdgeProperties.vue'

const store = useWorkflowStore()
const { pushHistory } = useHistory()
const node = computed(() => store.selectedNode)
const edge = computed(() => store.selectedEdge)

function deleteNode() {
  if (node.value) {
    store.removeNode(node.value.id)
    pushHistory()
  }
}
</script>

<template>
  <div class="properties-panel">
    <template v-if="node">
      <h3 class="panel-title">Properties</h3>
      <StartEventProperties v-if="node.type === 'startEvent'" :node="node" />
      <EndEventProperties v-else-if="node.type === 'endEvent'" :node="node" />
      <GatewayProperties v-else-if="node.type === 'exclusiveGateway'" :node="node" />
      <ActivityProperties v-else :node="node" />
      <button class="btn-delete" @click="deleteNode">Delete Node</button>
    </template>
    <template v-else-if="edge">
      <h3 class="panel-title">Properties</h3>
      <EdgeProperties :edge="edge" />
    </template>
    <template v-else>
      <div class="no-selection">
        <p>Select a node or edge to edit its properties</p>
      </div>
    </template>
  </div>
</template>

<style scoped>
.properties-panel {
  width: 280px;
  padding: 16px;
  background: #16213e;
  border-left: 1px solid #0f3460;
  overflow-y: auto;
}

.panel-title {
  font-size: 14px;
  font-weight: 700;
  color: #e0e0e0;
  margin-bottom: 12px;
}

.btn-delete {
  width: 100%;
  margin-top: 20px;
  padding: 8px;
  background: transparent;
  border: 1px solid #c0392b;
  color: #e74c3c;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  transition: background 0.2s;
}

.btn-delete:hover {
  background: rgba(231, 76, 60, 0.15);
}

.no-selection {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #607080;
  font-size: 13px;
  text-align: center;
}
</style>
