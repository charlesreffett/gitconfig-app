<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import SectionBlock from '@/components/blocks/SectionBlock.vue'
import FileBlock from '@/components/blocks/FileBlock.vue'
import ConditionalBlock from '@/components/blocks/ConditionalBlock.vue'
import CustomBlock from '@/components/blocks/CustomBlock.vue'
import { Button } from '@/components/ui/button'
import type { Block } from '@/types'

const props = defineProps<{
  blocks: Block[]
  supersededKeys: Set<string>
  title: string
  configPath: string
  unsaved: boolean
  onSave: () => void
  onReset: () => void
  onOpenProject?: () => void
  recentRepos?: string[]
  onLoadRecent?: (repoPath: string) => void
  onChange: () => void
}>()


/** Remove a block by stable ID (not positional index, which can be stale). */
function removeBlock(id: string) {
  const idx = props.blocks.findIndex(b => b.id === id)
  if (idx !== -1) {
    props.blocks.splice(idx, 1)
    props.onChange()
  }
}

/**
 * VueDraggable emits update:modelValue after a drag reorder.
 * We cannot reassign props.blocks (Vue 3 prop is readonly), so we
 * splice-in-place to keep the same Pinia array reference while
 * reflecting the new sort order.
 */
function onDragUpdate(newList: Block[]) {
  props.blocks.splice(0, props.blocks.length, ...newList)
  props.onChange()
}
</script>

<template>
  <div class="flex flex-col h-full min-h-0">
    <!-- Panel header -->
    <div class="flex items-center gap-2 px-4 py-2.5 border-b shrink-0 bg-background">
      <div class="flex-1 min-w-0">
        <h2 class="font-semibold text-sm truncate">{{ title }}</h2>
        <p class="text-xs text-muted-foreground truncate">{{ configPath || 'No file loaded' }}</p>
      </div>

      <Button
        v-if="onOpenProject"
        variant="outline"
        class="h-7 px-3 text-xs"
        @click="onOpenProject"
      >
        Open Project…
      </Button>

      <Button
        variant="outline"
        class="h-7 px-3 text-xs"
        :disabled="!unsaved"
        @click="onReset"
      >
        Reset
      </Button>

      <Button
        class="h-7 px-3 text-xs font-medium"
        :variant="unsaved ? 'default' : 'secondary'"
        :disabled="!unsaved"
        @click="onSave"
      >
        {{ unsaved ? 'Save*' : 'Saved' }}
      </Button>
    </div>

    <!-- Recent repos (project panel only, when no project loaded) -->
    <div
      v-if="onOpenProject && !configPath && recentRepos && recentRepos.length > 0"
      class="px-4 py-3 border-b shrink-0"
    >
      <p class="text-xs text-muted-foreground mb-2 font-medium uppercase tracking-wide">Recent</p>
      <ul class="space-y-1">
        <li
          v-for="repo in recentRepos"
          :key="repo"
          class="text-xs text-muted-foreground hover:text-foreground cursor-pointer truncate hover:underline"
          @click="onLoadRecent && onLoadRecent(repo)"
        >
          {{ repo }}
        </li>
      </ul>
    </div>

    <!-- Block list -->
    <div class="flex-1 overflow-y-auto min-h-0 px-4 py-3">
      <VueDraggable
        :model-value="props.blocks"
        handle=".drag-handle"
        :animation="150"
        ghost-class="opacity-40"
        class="space-y-3"
        @update:model-value="onDragUpdate"
      >
        <template v-for="block in props.blocks" :key="block.id">
          <SectionBlock
            v-if="block.type === 'section'"
            :block="block"
            :superseded-keys="supersededKeys"
            :on-remove="() => removeBlock(block.id)"
            :on-change="onChange"
          />
          <CustomBlock
            v-else-if="block.type === 'custom'"
            :block="block"
            :on-remove="() => removeBlock(block.id)"
            :on-change="onChange"
          />
          <FileBlock
            v-else-if="block.type === 'file'"
            :block="block"
            :superseded-keys="supersededKeys"
            :on-remove="() => removeBlock(block.id)"
            :on-change="onChange"
            :draggable="true"
          />
          <ConditionalBlock
            v-else-if="block.type === 'conditional'"
            :block="block"
            :superseded-keys="supersededKeys"
            :on-remove="() => removeBlock(block.id)"
            :on-change="onChange"
          />
        </template>
      </VueDraggable>

      <div
        v-if="props.blocks.length === 0"
        class="flex flex-col items-center justify-center py-16 text-center"
      >
        <p class="text-sm text-muted-foreground">No blocks yet.</p>
        <p class="text-xs text-muted-foreground mt-1">Click a block in the library to add it here.</p>
      </div>
    </div>
  </div>
</template>
