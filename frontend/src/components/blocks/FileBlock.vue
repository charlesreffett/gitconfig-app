<script setup lang="ts">
import { GripVertical, X } from 'lucide-vue-next'
import type { Block } from '@/types'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import SectionBlock from './SectionBlock.vue'
import CustomBlock from './CustomBlock.vue'

const props = defineProps<{
  block: Block
  supersededKeys: Set<string>
  onRemove: () => void
  onChange: () => void
  /** When true, shows drag handle (root-level file blocks can be reordered) */
  draggable?: boolean
}>()
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm overflow-hidden border-l-4 border-l-purple-400">
    <!-- Header -->
    <div class="flex items-center gap-2 px-3 py-2 border-b bg-muted/40">
      <GripVertical
        v-if="draggable"
        class="h-4 w-4 text-muted-foreground cursor-grab shrink-0 drag-handle"
      />
      <span class="text-base leading-none">📄</span>
      <span class="font-semibold text-sm text-foreground flex-1">File</span>
      <Button
        variant="ghost"
        size="icon"
        class="h-6 w-6 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
        @click="onRemove"
      >
        <X class="h-3.5 w-3.5" />
      </Button>
    </div>

    <!-- Path input -->
    <div class="px-3 py-2 border-b flex items-center gap-2">
      <span class="text-xs text-muted-foreground shrink-0">Path:</span>
      <Input
        :model-value="block.path"
        placeholder="~/path/to/config"
        class="h-7 flex-1 text-xs font-mono px-2"
        @update:model-value="(v) => { block.path = String(v); onChange() }"
      />
    </div>

    <!-- Child blocks (edits saved back to the included file) -->
    <div v-if="block.children.length > 0" class="px-3 py-2 space-y-2">
      <p class="text-xs text-muted-foreground uppercase tracking-wide font-medium pb-1">Included file contents</p>
      <template v-for="child in block.children" :key="child.id">
        <SectionBlock
          v-if="child.type === 'section'"
          :block="child"
          :superseded-keys="supersededKeys"
          :on-remove="() => {}"
          :on-change="onChange"
        />
        <CustomBlock
          v-else-if="child.type === 'custom'"
          :block="child"
          :on-remove="() => {}"
          :on-change="onChange"
        />
      </template>
    </div>

    <div v-else class="px-3 py-2">
      <p class="text-xs text-muted-foreground italic">
        File not found or empty — contents will appear here once the path is valid.
      </p>
    </div>
  </div>
</template>
