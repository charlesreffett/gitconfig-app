<script setup lang="ts">
import { GripVertical, X, Plus } from 'lucide-vue-next'
import type { Block } from '@/types'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

const props = defineProps<{
  block: Block
  onRemove: () => void
  onChange: () => void
}>()

function updateKey(idx: number, key: string) {
  props.block.entries[idx].key = key
  props.onChange()
}

function updateValue(idx: number, value: string) {
  props.block.entries[idx].value = value
  props.onChange()
}

function removeEntry(idx: number) {
  props.block.entries.splice(idx, 1)
  props.onChange()
}

function addEntry() {
  props.block.entries.push({ key: '', value: '' })
  props.onChange()
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm overflow-hidden">
    <!-- Header -->
    <div class="flex items-center gap-2 px-3 py-2 border-b bg-muted/40">
      <GripVertical class="h-4 w-4 text-muted-foreground cursor-grab shrink-0 drag-handle" />
      <span class="text-base leading-none">⚙</span>
      <Input
        v-model="block.schemaName"
        placeholder="section name"
        class="flex-1 h-auto border-none bg-transparent shadow-none text-sm font-semibold focus-visible:ring-0 focus-visible:ring-offset-0 px-0 placeholder:text-muted-foreground"
        @update:model-value="onChange"
      />
      <template v-if="block.subsection !== undefined">
        <Input
          v-model="block.subsection"
          placeholder="subsection"
          class="w-32 h-6 text-xs px-2"
          @update:model-value="onChange"
        />
      </template>
      <Button
        variant="ghost"
        size="icon"
        class="ml-1 h-6 w-6 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
        @click="onRemove"
      >
        <X class="h-3.5 w-3.5" />
      </Button>
    </div>

    <!-- Body -->
    <div class="px-3 py-2 space-y-1.5">
      <div
        v-for="(entry, idx) in block.entries"
        :key="idx"
        class="flex items-center gap-2"
      >
        <Input
          :model-value="entry.key"
          placeholder="key"
          class="h-7 w-32 text-xs font-mono px-2"
          @update:model-value="updateKey(idx, String($event))"
        />
        <span class="text-muted-foreground text-xs shrink-0">=</span>
        <Input
          :model-value="entry.value"
          placeholder="value"
          class="h-7 flex-1 text-xs font-mono px-2"
          @update:model-value="updateValue(idx, String($event))"
        />
        <Button
          variant="ghost"
          size="icon"
          class="h-6 w-6 shrink-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
          @click="removeEntry(idx)"
        >
          <X class="h-3 w-3" />
        </Button>
      </div>

      <p v-if="block.entries.length === 0" class="text-xs text-muted-foreground italic py-1">
        No entries yet.
      </p>

      <Separator class="my-1" />

      <Button
        variant="ghost"
        class="h-auto py-0.5 px-1 text-xs text-muted-foreground hover:text-foreground gap-1"
        @click="addEntry"
      >
        <Plus class="h-3 w-3" />
        Add entry
      </Button>
    </div>
  </div>
</template>
