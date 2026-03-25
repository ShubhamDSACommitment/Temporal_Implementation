import { ref, readonly } from 'vue'
import { listActivities, type ActivityCategory } from '../api/client'
import type { PaletteCategory, PaletteItem } from '../types/workflow'

const categories = ref<PaletteCategory[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
let fetched = false

async function fetchActivities() {
  if (fetched) return
  loading.value = true
  error.value = null
  try {
    const apiCategories: ActivityCategory[] = await listActivities()
    categories.value = apiCategories.map((cat) => ({
      name: cat.name,
      collapsed: false,
      items: cat.activities.map(
        (a): PaletteItem => ({
          kind: 'activity',
          displayName: a.display_name,
          activityName: a.activity_name,
          defaultTaskQueue: a.task_queue,
          defaultInputMapping: a.default_mapping,
          inputSchema: a.input_schema,
          icon: 'gear',
        }),
      ),
    }))
    fetched = true
  } catch (e: any) {
    error.value = e?.message ?? 'Failed to load activities'
  } finally {
    loading.value = false
  }
}

function refresh() {
  fetched = false
  return fetchActivities()
}

export function useActivityRegistry() {
  return {
    categories: readonly(categories),
    loading: readonly(loading),
    error: readonly(error),
    fetchActivities,
    refresh,
  }
}
