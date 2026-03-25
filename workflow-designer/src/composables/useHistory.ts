import { ref, computed } from 'vue'
import { useWorkflowStore } from '../stores/workflow'

interface Snapshot {
  nodes: any[]
  edges: any[]
}

const history = ref<Snapshot[]>([])
const historyIndex = ref(-1)
let debounceTimer: ReturnType<typeof setTimeout> | null = null
let isRestoring = false

export function useHistory() {
  const store = useWorkflowStore()

  const canUndo = computed(() => historyIndex.value > 0)
  const canRedo = computed(() => historyIndex.value < history.value.length - 1)

  function takeSnapshot(): Snapshot {
    return {
      nodes: JSON.parse(JSON.stringify(store.nodes)),
      edges: JSON.parse(JSON.stringify(store.edges)),
    }
  }

  function pushHistory() {
    if (isRestoring) return
    if (debounceTimer) clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      // Discard any future states after current index
      history.value = history.value.slice(0, historyIndex.value + 1)
      const snap = takeSnapshot()
      history.value.push(snap)
      // Keep max 50
      if (history.value.length > 50) {
        history.value.shift()
      } else {
        historyIndex.value = history.value.length - 1
      }
    }, 300)
  }

  function restoreSnapshot(snap: Snapshot) {
    isRestoring = true
    store.nodes.splice(0, store.nodes.length, ...JSON.parse(JSON.stringify(snap.nodes)))
    store.edges.splice(0, store.edges.length, ...JSON.parse(JSON.stringify(snap.edges)))
    isRestoring = false
  }

  function undo() {
    if (!canUndo.value) return
    historyIndex.value--
    restoreSnapshot(history.value[historyIndex.value])
  }

  function redo() {
    if (!canRedo.value) return
    historyIndex.value++
    restoreSnapshot(history.value[historyIndex.value])
  }

  function initHistory() {
    history.value = [takeSnapshot()]
    historyIndex.value = 0
  }

  return { pushHistory, undo, redo, canUndo, canRedo, initHistory }
}
