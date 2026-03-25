import { onMounted, onUnmounted } from 'vue'
import { useHistory } from './useHistory'
import { useWorkflowStore } from '../stores/workflow'

export function useKeyboardShortcuts(onSave: () => void) {
  const { undo, redo } = useHistory()
  const store = useWorkflowStore()

  function handler(e: KeyboardEvent) {
    const mod = e.metaKey || e.ctrlKey

    // Ctrl+Z / Cmd+Z → undo
    if (mod && e.key === 'z' && !e.shiftKey) {
      e.preventDefault()
      undo()
      return
    }

    // Ctrl+Y / Cmd+Shift+Z → redo
    if ((mod && e.key === 'y') || (mod && e.shiftKey && e.key === 'z') || (mod && e.shiftKey && e.key === 'Z')) {
      e.preventDefault()
      redo()
      return
    }

    // Delete / Backspace → delete selected node
    if ((e.key === 'Delete' || e.key === 'Backspace') && !isInputFocused()) {
      e.preventDefault()
      if (store.selectedNodeId) {
        store.removeNode(store.selectedNodeId)
      }
      return
    }

    // Ctrl+S / Cmd+S → save
    if (mod && e.key === 's') {
      e.preventDefault()
      onSave()
      return
    }
  }

  function isInputFocused(): boolean {
    const el = document.activeElement
    if (!el) return false
    const tag = el.tagName.toLowerCase()
    return tag === 'input' || tag === 'textarea' || tag === 'select' || (el as HTMLElement).isContentEditable
  }

  onMounted(() => {
    window.addEventListener('keydown', handler)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handler)
  })
}
