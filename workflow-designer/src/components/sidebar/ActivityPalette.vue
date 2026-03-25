<script setup lang="ts">
import { reactive, onMounted, computed } from 'vue'
import { PALETTE_CATEGORIES, type PaletteItem, type PaletteCategory } from '../../types/workflow'
import { useActivityRegistry } from '../../composables/useActivityRegistry'

const { categories: registryCategories, loading, error: registryError, fetchActivities } = useActivityRegistry()

// Static events category
const eventCategories = reactive(PALETTE_CATEGORIES.map((c) => ({ ...c, collapsed: c.collapsed })))

const allCategories = computed<PaletteCategory[]>(() => [
  ...eventCategories,
  ...registryCategories.value,
])

onMounted(() => {
  fetchActivities()
})

function toggleCategory(idx: number) {
  if (idx < eventCategories.length) {
    eventCategories[idx].collapsed = !eventCategories[idx].collapsed
  }
  // Dynamic categories are always open (or could be toggled via a separate reactive)
}

function onDragStart(event: DragEvent, item: PaletteItem) {
  if (!event.dataTransfer) return
  event.dataTransfer.setData('application/node-kind', item.kind)
  event.dataTransfer.setData('application/display-name', item.displayName)
  if (item.activityName) {
    event.dataTransfer.setData('application/activity-name', item.activityName)
  }
  if (item.defaultTaskQueue) {
    event.dataTransfer.setData('application/task-queue', item.defaultTaskQueue)
  }
  if (item.defaultInputMapping) {
    event.dataTransfer.setData('application/input-mapping', JSON.stringify(item.defaultInputMapping))
  }
  event.dataTransfer.effectAllowed = 'move'
}
</script>

<template>
  <div class="palette">
    <h3 class="palette-title">Palette</h3>
    <div class="palette-hint">Drag to canvas</div>

    <div v-for="(cat, idx) in allCategories" :key="cat.name" class="palette-category">
      <div class="category-header" @click="toggleCategory(idx)">
        <svg
          class="chevron"
          :class="{ collapsed: cat.collapsed }"
          width="12" height="12" viewBox="0 0 12 12"
        >
          <path d="M3 4.5L6 7.5L9 4.5" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
        </svg>
        <span>{{ cat.name }}</span>
      </div>

      <div v-show="!cat.collapsed" class="category-items">
        <div
          v-for="item in cat.items"
          :key="item.displayName"
          class="palette-item"
          draggable="true"
          @dragstart="onDragStart($event, item)"
        >
          <div class="palette-item-icon">
            <!-- Start event icon -->
            <svg v-if="item.icon === 'start'" width="18" height="18" viewBox="0 0 18 18">
              <circle cx="9" cy="9" r="8" fill="none" stroke="#27ae60" stroke-width="1.5" />
              <polygon points="7,5 7,13 13,9" fill="#27ae60" />
            </svg>
            <!-- End event icon -->
            <svg v-else-if="item.icon === 'end'" width="18" height="18" viewBox="0 0 18 18">
              <circle cx="9" cy="9" r="8" fill="none" stroke="#e74c3c" stroke-width="2.5" />
              <rect x="6" y="6" width="6" height="6" fill="#e74c3c" />
            </svg>
            <!-- Gateway icon -->
            <svg v-else-if="item.icon === 'gateway'" width="18" height="18" viewBox="0 0 18 18">
              <rect x="9" y="0" width="12.73" height="12.73" rx="1" transform="rotate(45 9 9)" fill="none" stroke="#f39c12" stroke-width="1.5" />
              <line x1="6.5" y1="6.5" x2="11.5" y2="11.5" stroke="#f39c12" stroke-width="1.2" stroke-linecap="round" />
              <line x1="11.5" y1="6.5" x2="6.5" y2="11.5" stroke="#f39c12" stroke-width="1.2" stroke-linecap="round" />
            </svg>
            <!-- Gear icon for activities -->
            <svg v-else width="18" height="18" viewBox="0 0 18 18">
              <circle cx="9" cy="9" r="5" fill="none" stroke="#60a0d0" stroke-width="1.2" />
              <circle cx="9" cy="9" r="2" fill="#60a0d0" />
              <line x1="9" y1="2" x2="9" y2="4" stroke="#60a0d0" stroke-width="1" />
              <line x1="9" y1="14" x2="9" y2="16" stroke="#60a0d0" stroke-width="1" />
              <line x1="2" y1="9" x2="4" y2="9" stroke="#60a0d0" stroke-width="1" />
              <line x1="14" y1="9" x2="16" y2="9" stroke="#60a0d0" stroke-width="1" />
            </svg>
          </div>
          <div class="palette-item-text">
            <div class="palette-item-name">{{ item.displayName }}</div>
            <div v-if="item.defaultTaskQueue" class="palette-item-queue">{{ item.defaultTaskQueue }}</div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="loading" class="palette-loading">
      <div class="spinner"></div>
      <span>Loading activities...</span>
    </div>
    <div v-else-if="registryError" class="palette-error">
      No activities registered
    </div>
  </div>
</template>

<style scoped>
.palette {
  width: 220px;
  padding: 16px;
  background: #16213e;
  border-right: 1px solid #0f3460;
  overflow-y: auto;
}

.palette-title {
  font-size: 14px;
  font-weight: 700;
  color: #e0e0e0;
  margin-bottom: 4px;
}

.palette-hint {
  font-size: 11px;
  color: #607080;
  margin-bottom: 16px;
}

.palette-category {
  margin-bottom: 12px;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  color: #8090a0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 4px 0;
  user-select: none;
}

.category-header:hover {
  color: #c0c0d0;
}

.chevron {
  transition: transform 0.2s;
  color: #607080;
}

.chevron.collapsed {
  transform: rotate(-90deg);
}

.category-items {
  margin-top: 6px;
}

.palette-item {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #1e2a3a;
  border: 1px solid #2a4060;
  border-radius: 6px;
  padding: 8px 10px;
  margin-bottom: 6px;
  cursor: grab;
  transition: border-color 0.2s, transform 0.1s;
}

.palette-item:hover {
  border-color: #5dade2;
  transform: translateY(-1px);
}

.palette-item:active {
  cursor: grabbing;
}

.palette-item-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.palette-item-text {
  flex: 1;
  min-width: 0;
}

.palette-item-name {
  font-size: 12px;
  font-weight: 600;
  color: #ffffff;
}

.palette-item-queue {
  font-size: 10px;
  color: #607080;
  margin-top: 1px;
  font-family: monospace;
}

.palette-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 0;
  color: #607080;
  font-size: 12px;
}

.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid #2a4060;
  border-top-color: #5dade2;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.palette-error {
  padding: 12px 0;
  color: #607080;
  font-size: 12px;
  font-style: italic;
}
</style>
