<script setup lang="ts">
import { computed, ref } from 'vue'
import { GripVertical, X, Plus, ChevronsUpDown } from 'lucide-vue-next'
import { useAppStore } from '@/stores/app'
import type { Block, Entry } from '@/types'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import {
  Select, SelectContent, SelectGroup, SelectItem,
  SelectTrigger, SelectValue,
} from '@/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import {
  Command, CommandEmpty, CommandGroup, CommandInput,
  CommandItem, CommandList,
} from '@/components/ui/command'

const props = defineProps<{
  block: Block
  supersededKeys: Set<string>
  onRemove: () => void
  onChange: () => void
}>()

const store = useAppStore()
const addFieldOpen = ref(false)

const sectionSchema = computed(() =>
  store.schemaByName[props.block.schemaName] ?? null
)

const availableFields = computed(() => {
  if (!sectionSchema.value || sectionSchema.value.freeForm) return []
  const present = new Set(props.block.entries.map(e => e.key))
  return sectionSchema.value.fields.filter(f => !present.has(f.key))
})

const isSuperseded = (entry: Entry) => {
  const key = `${props.block.schemaName}\0${props.block.subsection}\0${entry.key}`
  return props.supersededKeys.has(key)
}

function fieldFor(key: string) {
  return sectionSchema.value?.fields.find(f => f.key === key)
}

function updateEntry(idx: number, value: string) {
  props.block.entries[idx].value = value
  props.onChange()
}

function removeEntry(idx: number) {
  props.block.entries.splice(idx, 1)
  props.onChange()
}

function addField(key: string) {
  props.block.entries.push({ key, value: '' })
  props.onChange()
  addFieldOpen.value = false
}

function addFreeEntry() {
  props.block.entries.push({ key: '', value: '' })
  props.onChange()
}

function updateEntryKey(idx: number, key: string) {
  props.block.entries[idx].key = key
  props.onChange()
}
</script>

<template>
  <div class="rounded-lg border bg-card text-card-foreground shadow-sm overflow-hidden">
    <!-- Header -->
    <div class="flex items-center gap-2 px-3 py-2 border-b bg-muted/40">
      <GripVertical class="h-4 w-4 text-muted-foreground cursor-grab shrink-0 drag-handle" />
      <span class="text-base leading-none">{{ sectionSchema?.icon ?? '⚙' }}</span>
      <span class="font-semibold text-sm flex-1">{{ sectionSchema?.label ?? block.schemaName }}</span>

      <!-- Subsection input (for remote "origin", branch "main", etc.) -->
      <template v-if="sectionSchema?.hasSubsection">
        <span class="text-xs text-muted-foreground">{{ sectionSchema.subsectionLabel }}</span>
        <Input
          :model-value="block.subsection"
          :placeholder="sectionSchema.subsectionPlaceholder"
          class="h-6 w-36 text-xs px-2"
          @update:model-value="(v) => { block.subsection = String(v); onChange() }"
        />
      </template>

      <Button
        variant="ghost"
        size="icon"
        class="ml-auto h-6 w-6 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
        @click="onRemove"
      >
        <X class="h-3.5 w-3.5" />
      </Button>
    </div>

    <!-- Body -->
    <div class="px-3 py-2 space-y-1.5">
      <!-- Field rows -->
      <template v-if="sectionSchema?.freeForm">
        <!-- Free-form rows (alias, etc.) -->
        <div
          v-for="(entry, idx) in block.entries"
          :key="idx"
          class="flex items-center gap-2"
        >
          <Input
            :model-value="entry.key"
            placeholder="key"
            class="h-7 w-32 text-xs font-mono px-2"
            @update:model-value="updateEntryKey(idx, String($event))"
          />
          <span class="text-muted-foreground text-xs shrink-0">=</span>
          <Input
            :model-value="entry.value"
            placeholder="value"
            class="h-7 flex-1 text-xs font-mono px-2"
            @update:model-value="updateEntry(idx, String($event))"
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
      </template>

      <template v-else>
        <!-- Schema-typed rows -->
        <div
          v-for="(entry, idx) in block.entries"
          :key="entry.key + idx"
          class="flex items-center gap-2"
          :class="{ 'opacity-40': isSuperseded(entry) }"
        >
          <Label
            class="w-40 shrink-0 text-xs font-normal text-muted-foreground truncate cursor-default"
            :title="fieldFor(entry.key)?.docs ?? entry.key"
          >
            {{ fieldFor(entry.key)?.label ?? entry.key }}
          </Label>

          <!-- Toggle -->
          <template v-if="fieldFor(entry.key)?.type === 'toggle'">
            <Switch
              :checked="entry.value === 'true'"
              @update:checked="(v: boolean) => updateEntry(idx, v ? 'true' : 'false')"
            />
          </template>

          <!-- Dropdown -->
          <template v-else-if="fieldFor(entry.key)?.type === 'dropdown'">
            <Select
              :model-value="entry.value"
              @update:model-value="(v) => v !== null && updateEntry(idx, String(v))"
            >
              <SelectTrigger class="h-7 flex-1 text-xs px-2">
                <SelectValue :placeholder="fieldFor(entry.key)?.default ?? 'Select…'" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem
                    v-for="opt in fieldFor(entry.key)?.options ?? []"
                    :key="opt"
                    :value="opt"
                    class="text-xs"
                  >
                    {{ opt }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </template>

          <!-- Multiline -->
          <template v-else-if="fieldFor(entry.key)?.type === 'multiline'">
            <textarea
              :value="entry.value"
              rows="2"
              class="flex-1 rounded-md border border-input bg-background px-2 py-1 text-xs font-mono resize-y focus:outline-none focus:ring-2 focus:ring-ring"
              @input="updateEntry(idx, ($event.target as HTMLTextAreaElement).value)"
            />
          </template>

          <!-- Text (default) -->
          <template v-else>
            <Input
              :model-value="entry.value"
              :placeholder="fieldFor(entry.key)?.default ?? ''"
              class="h-7 flex-1 text-xs px-2"
              @update:model-value="updateEntry(idx, String($event))"
            />
          </template>

          <Button
            variant="ghost"
            size="icon"
            class="h-6 w-6 shrink-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
            @click="removeEntry(idx)"
          >
            <X class="h-3 w-3" />
          </Button>
        </div>
      </template>

      <!-- Placeholder when empty -->
      <p
        v-if="block.entries.length === 0"
        class="text-xs text-muted-foreground italic py-1"
      >
        No fields yet — use Add field below.
      </p>

      <!-- Divider -->
      <Separator class="my-1" />

      <!-- Add field — free-form mode -->
      <template v-if="sectionSchema?.freeForm">
        <Button
          variant="ghost"
          class="h-auto py-0.5 px-1 text-xs text-muted-foreground hover:text-foreground gap-1"
          @click="addFreeEntry"
        >
          <Plus class="h-3 w-3" />
          Add entry
        </Button>
      </template>

      <!-- Add field — schema-driven combobox -->
      <template v-else-if="availableFields.length > 0">
        <Popover v-model:open="addFieldOpen">
          <PopoverTrigger as-child>
            <Button
              variant="outline"
              class="h-7 w-full justify-start text-xs font-normal border-dashed border-muted-foreground/40 text-muted-foreground hover:text-foreground hover:border-muted-foreground gap-1.5"
            >
              <Plus class="h-3 w-3 shrink-0" />
              Add field…
              <ChevronsUpDown class="ml-auto h-3 w-3 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent class="w-72 p-0" align="start">
            <Command>
              <CommandInput placeholder="Search fields…" class="h-9 text-xs" />
              <CommandList>
                <CommandEmpty class="py-3 text-xs text-center text-muted-foreground">
                  No matching fields.
                </CommandEmpty>
                <CommandGroup>
                  <CommandItem
                    v-for="field in availableFields"
                    :key="field.key"
                    :value="field.label + ' ' + field.key"
                    class="text-xs cursor-pointer"
                    :title="field.docs"
                    @select="addField(field.key)"
                  >
                    <span>{{ field.label }}</span>
                    <span class="ml-auto font-mono text-muted-foreground">{{ field.key }}</span>
                  </CommandItem>
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
      </template>

      <p
        v-else-if="!sectionSchema?.freeForm && availableFields.length === 0 && block.entries.length > 0"
        class="text-xs text-muted-foreground italic"
      >
        All known fields added.
      </p>
    </div>
  </div>
</template>
