<script setup lang="ts">
import { GripVertical, X, Plus } from 'lucide-vue-next'
import { useAppStore } from '@/stores/app'
import type { Block } from '@/types'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import SectionBlock from './SectionBlock.vue'
import CustomBlock from './CustomBlock.vue'

const props = defineProps<{
  block: Block
  supersededKeys: Set<string>
  onRemove: () => void
  onChange: () => void
}>()

const store = useAppStore()

function addCondition() {
  props.block.conditions.push({ kind: store.conditionTypes[0] ?? 'gitdir', value: '' })
  props.onChange()
}

function removeCondition(idx: number) {
  if (props.block.conditions.length <= 1) return
  props.block.conditions.splice(idx, 1)
  props.onChange()
}

function updateConditionKind(idx: number, kind: string) {
  props.block.conditions[idx].kind = kind
  props.onChange()
}

function updateConditionValue(idx: number, value: string) {
  props.block.conditions[idx].value = value
  props.onChange()
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm overflow-hidden border-l-4 border-l-amber-400">
    <!-- Header -->
    <div class="flex items-center gap-2 px-3 py-2 border-b bg-muted/40">
      <GripVertical class="h-4 w-4 text-muted-foreground cursor-grab shrink-0 drag-handle" />
      <span class="text-base leading-none">⚡</span>
      <span class="font-semibold text-sm text-foreground flex-1">Conditional Include</span>
      <Button
        variant="ghost"
        size="icon"
        class="h-6 w-6 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
        @click="onRemove"
      >
        <X class="h-3.5 w-3.5" />
      </Button>
    </div>

    <!-- Condition rows -->
    <div class="px-3 py-2 space-y-1.5">
      <div
        v-for="(cond, idx) in block.conditions"
        :key="idx"
        class="flex items-center gap-2"
      >
        <!-- Native select: avoids async-hydration issues with Reka UI Select -->
        <select
          :value="cond.kind"
          class="h-7 w-32 shrink-0 rounded-md border border-input bg-background px-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring appearance-none cursor-pointer"
          @change="updateConditionKind(idx, ($event.target as HTMLSelectElement).value)"
        >
          <option v-for="ct in store.conditionTypes" :key="ct" :value="ct">{{ ct }}</option>
        </select>

        <Input
          :model-value="cond.value"
          placeholder="e.g. ~/work/**"
          class="h-7 flex-1 text-xs font-mono px-2"
          @update:model-value="updateConditionValue(idx, String($event))"
        />

        <Button
          variant="ghost"
          size="icon"
          class="h-6 w-6 shrink-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
          :class="{ 'opacity-30 pointer-events-none': block.conditions.length <= 1 }"
          :disabled="block.conditions.length <= 1"
          @click="removeCondition(idx)"
        >
          <X class="h-3 w-3" />
        </Button>
      </div>

      <Button
        variant="ghost"
        class="h-auto py-0.5 px-1 text-xs text-muted-foreground hover:text-foreground gap-1"
        @click="addCondition"
      >
        <Plus class="h-3 w-3" />
        Add condition
      </Button>

      <Separator class="my-1.5" />

      <!-- Included file path (inline — no nested FileBlock) -->
      <div class="flex items-center gap-2">
        <span class="text-xs text-muted-foreground shrink-0">Path:</span>
        <Input
          :model-value="block.path"
          placeholder="~/path/to/config"
          class="h-7 flex-1 text-xs font-mono px-2"
          @update:model-value="(v) => { block.path = String(v); onChange() }"
        />
      </div>

      <!-- Included file child blocks (read-only) -->
      <div v-if="block.children.length > 0" class="pt-1 space-y-2">
        <p class="text-xs text-muted-foreground uppercase tracking-wide font-medium">Included file contents</p>
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

      <p v-else-if="!block.path" class="text-xs text-muted-foreground italic">
        Enter a path above to preview the included file's contents.
      </p>
      <p v-else class="text-xs text-muted-foreground italic">
        File not found or empty — contents appear here once the path is valid.
      </p>
    </div>
  </div>
</template>
