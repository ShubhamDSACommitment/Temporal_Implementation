<script setup lang="ts">
import { ref } from 'vue'
import WorkflowToolbar from '../components/toolbar/WorkflowToolbar.vue'
import ActivityPalette from '../components/sidebar/ActivityPalette.vue'
import WorkflowCanvas from '../components/canvas/WorkflowCanvas.vue'
import PropertiesPanel from '../components/properties/PropertiesPanel.vue'
import { useKeyboardShortcuts } from '../composables/useKeyboardShortcuts'
import { useHistory } from '../composables/useHistory'

const toolbarRef = ref<InstanceType<typeof WorkflowToolbar> | null>(null)

const { initHistory } = useHistory()
initHistory()

useKeyboardShortcuts(() => {
  toolbarRef.value?.handleSave()
})
</script>

<template>
  <div class="designer-layout">
    <WorkflowToolbar ref="toolbarRef" />
    <div class="designer-body">
      <ActivityPalette />
      <WorkflowCanvas />
      <PropertiesPanel />
    </div>
  </div>
</template>

<style scoped>
.designer-layout {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.designer-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}
</style>
