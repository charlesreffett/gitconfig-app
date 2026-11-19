<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import type { Block, ConditionDTO } from '@/types'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

const props = defineProps<{
  onAddToGlobal: (block: Block) => void
  onAddToProject: (block: Block) => void
}>()

const store = useAppStore()

function makeId() {
  return Math.random().toString(36).slice(2, 10)
}

function makeSection(schemaName: string): Block {
  return {
    id: makeId(),
    type: 'section',
    schemaName,
    subsection: '',
    entries: [],
    path: '',
    children: [],
    conditions: [],
    file: null,
  }
}

function makeFileBlock(): Block {
  return {
    id: makeId(),
    type: 'file',
    schemaName: '',
    subsection: '',
    entries: [],
    path: '',
    children: [],
    conditions: [],
    file: null,
  }
}

function makeConditionalBlock(): Block {
  const cond: ConditionDTO = { kind: store.conditionTypes[0] ?? 'gitdir', value: '' }
  return {
    id: makeId(),
    type: 'conditional',
    schemaName: '',
    subsection: '',
    entries: [],
    path: '',       // path to the included file lives directly on the block
    children: [],   // parsed contents of the included file (from backend; edits saved back to that file)
    conditions: [cond],
    file: null,
  }
}

function makeCustomBlock(): Block {
  return {
    id: makeId(),
    type: 'custom',
    schemaName: '',
    subsection: '',
    entries: [],
    path: '',
    children: [],
    conditions: [],
    file: null,
  }
}

function addTo(target: 'global' | 'project', block: Block) {
  if (target === 'global') props.onAddToGlobal(block)
  else props.onAddToProject(block)
}

const hasProject = () => store.projectPath !== ''
</script>

<template>
  <div class="flex flex-col h-full border-r bg-background overflow-y-auto">
    <div class="px-3 py-3 border-b shrink-0">
      <h2 class="font-semibold text-sm text-foreground">Block Library</h2>
      <p class="text-xs text-muted-foreground mt-0.5">Click or right-click to choose target</p>
    </div>

    <div class="px-2 py-2 space-y-4 flex-1">
      <!-- Boundary blocks -->
      <div>
        <p class="text-xs font-medium text-muted-foreground uppercase tracking-wide px-1 mb-1.5">Boundaries</p>
        <div class="space-y-0.5">
          <!-- Conditional -->
          <DropdownMenu>
            <div class="flex items-stretch rounded-md hover:bg-accent group">
              <Button
                variant="ghost"
                class="flex-1 justify-start gap-2 px-2 h-8 text-sm font-normal rounded-r-none hover:bg-transparent"
                @click="addTo('global', makeConditionalBlock())"
              >
                <span>⚡</span>
                <span>Conditional</span>
              </Button>
              <DropdownMenuTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-8 w-6 rounded-l-none border-l border-transparent group-hover:border-border hover:bg-transparent"
                  title="Choose target config"
                >
                  <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 16 16"><path d="M8 11L3 6h10l-5 5z"/></svg>
                </Button>
              </DropdownMenuTrigger>
            </div>
            <DropdownMenuContent align="start" class="w-44">
              <DropdownMenuLabel class="text-xs">Add to…</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem class="text-xs" @click="addTo('global', makeConditionalBlock())">
                🌍 Global config
              </DropdownMenuItem>
              <DropdownMenuItem class="text-xs" :disabled="!hasProject()" @click="addTo('project', makeConditionalBlock())">
                📁 Project config
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          <!-- File -->
          <DropdownMenu>
            <div class="flex items-stretch rounded-md hover:bg-accent group">
              <Button
                variant="ghost"
                class="flex-1 justify-start gap-2 px-2 h-8 text-sm font-normal rounded-r-none hover:bg-transparent"
                @click="addTo('global', makeFileBlock())"
              >
                <span>📄</span>
                <span>File</span>
              </Button>
              <DropdownMenuTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-8 w-6 rounded-l-none border-l border-transparent group-hover:border-border hover:bg-transparent"
                  title="Choose target config"
                >
                  <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 16 16"><path d="M8 11L3 6h10l-5 5z"/></svg>
                </Button>
              </DropdownMenuTrigger>
            </div>
            <DropdownMenuContent align="start" class="w-44">
              <DropdownMenuLabel class="text-xs">Add to…</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem class="text-xs" @click="addTo('global', makeFileBlock())">
                🌍 Global config
              </DropdownMenuItem>
              <DropdownMenuItem class="text-xs" :disabled="!hasProject()" @click="addTo('project', makeFileBlock())">
                📁 Project config
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <!-- Known sections -->
      <div>
        <p class="text-xs font-medium text-muted-foreground uppercase tracking-wide px-1 mb-1.5">Sections</p>
        <div class="space-y-0.5">
          <DropdownMenu v-for="s in store.schema" :key="s.name">
            <div class="flex items-stretch rounded-md hover:bg-accent group">
              <Button
                variant="ghost"
                class="flex-1 justify-start gap-2 px-2 h-8 text-sm font-normal rounded-r-none hover:bg-transparent"
                @click="addTo('global', makeSection(s.name))"
              >
                <span>{{ s.icon }}</span>
                <span>{{ s.label }}</span>
              </Button>
              <DropdownMenuTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-8 w-6 rounded-l-none border-l border-transparent group-hover:border-border hover:bg-transparent"
                  title="Choose target config"
                >
                  <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 16 16"><path d="M8 11L3 6h10l-5 5z"/></svg>
                </Button>
              </DropdownMenuTrigger>
            </div>
            <DropdownMenuContent align="start" class="w-44">
              <DropdownMenuLabel class="text-xs">Add to…</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem class="text-xs" @click="addTo('global', makeSection(s.name))">
                🌍 Global config
              </DropdownMenuItem>
              <DropdownMenuItem class="text-xs" :disabled="!hasProject()" @click="addTo('project', makeSection(s.name))">
                📁 Project config
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          <!-- Custom (always last) -->
          <DropdownMenu>
            <div class="flex items-stretch rounded-md hover:bg-accent group">
              <Button
                variant="ghost"
                class="flex-1 justify-start gap-2 px-2 h-8 text-sm font-normal rounded-r-none hover:bg-transparent text-muted-foreground"
                @click="addTo('global', makeCustomBlock())"
              >
                <span>⚙</span>
                <span>Custom</span>
              </Button>
              <DropdownMenuTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-8 w-6 rounded-l-none border-l border-transparent group-hover:border-border hover:bg-transparent"
                  title="Choose target config"
                >
                  <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 16 16"><path d="M8 11L3 6h10l-5 5z"/></svg>
                </Button>
              </DropdownMenuTrigger>
            </div>
            <DropdownMenuContent align="start" class="w-44">
              <DropdownMenuLabel class="text-xs">Add to…</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem class="text-xs" @click="addTo('global', makeCustomBlock())">
                🌍 Global config
              </DropdownMenuItem>
              <DropdownMenuItem class="text-xs" :disabled="!hasProject()" @click="addTo('project', makeCustomBlock())">
                📁 Project config
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>
  </div>
</template>
