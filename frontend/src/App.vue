<script setup lang="ts">
import { onMounted } from 'vue'
import { Moon, Sun } from 'lucide-vue-next'
import { useAppStore } from '@/stores/app'
import ConfigCanvas from '@/components/layout/ConfigCanvas.vue'
import LibraryPanel from '@/components/layout/LibraryPanel.vue'
import type { Block } from '@/types'

const store = useAppStore()

onMounted(() => store.init())

function addToGlobal(block: Block) {
  store.globalBlocks.push(block)
  store.markGlobalUnsaved()
}

function addToProject(block: Block) {
  if (!store.projectPath) return
  store.projectBlocks.push(block)
  store.markProjectUnsaved()
}

async function loadRecentProject(repoPath: string) {
  await store.loadProject(repoPath + '/.git/config', repoPath)
}
</script>

<template>
  <div class="flex flex-col h-screen overflow-hidden" :class="store.theme === 'dark' ? 'dark' : ''">
    <!-- Top bar -->
    <header class="flex items-center gap-3 px-4 h-11 border-b bg-background shrink-0 drag-region">
      <!-- macOS traffic light space -->
      <div class="w-16 shrink-0 no-drag" />

      <h1 class="font-semibold text-sm">Git Config Manager</h1>

      <div class="flex-1" />

      <button
        class="h-7 w-7 rounded flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent transition-colors no-drag"
        :title="store.theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
        @click="store.toggleTheme"
      >
        <Moon v-if="store.theme === 'light'" class="h-4 w-4" />
        <Sun v-else class="h-4 w-4" />
      </button>
    </header>

    <!-- Main 3-column layout -->
    <div class="flex flex-1 min-h-0 overflow-hidden">
      <!-- Library (fixed width) -->
      <aside class="w-48 shrink-0">
        <LibraryPanel
          :on-add-to-global="addToGlobal"
          :on-add-to-project="addToProject"
        />
      </aside>

      <!-- Global config canvas -->
      <main class="flex-1 min-w-0 border-r">
        <ConfigCanvas
          :blocks="store.globalBlocks"
          :superseded-keys="store.supersededKeys"
          title="Global Config"
          :config-path="store.globalPath"
          :unsaved="store.globalUnsaved"
          :on-save="store.saveGlobal"
          :on-reset="store.resetGlobal"
          :on-change="store.markGlobalUnsaved"
        />
      </main>

      <!-- Project config canvas -->
      <aside class="flex-1 min-w-0">
        <ConfigCanvas
          :blocks="store.projectBlocks"
          :superseded-keys="new Set()"
          title="Project Config"
          :config-path="store.projectPath"
          :unsaved="store.projectUnsaved"
          :on-save="store.saveProject"
          :on-reset="store.resetProject"
          :on-open-project="store.openProject"
          :recent-repos="store.recentRepos"
          :on-load-recent="loadRecentProject"
          :on-change="store.markProjectUnsaved"
        />
      </aside>
    </div>

    <!-- Status bar -->
    <footer class="h-6 px-4 flex items-center border-t bg-muted/40 shrink-0">
      <span class="text-xs text-muted-foreground">{{ store.status }}</span>
    </footer>
  </div>
</template>
