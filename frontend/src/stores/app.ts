import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  LoadConfig, SaveConfig, GetSchema, GetConditionTypes,
  GetGlobalConfigPath, GetRecentRepos, AddRecentRepo, SelectProjectFolder,
} from '../../wailsjs/go/main/App'
import { type Block, type SectionSchemaDTO, dtoToBlock, blockToDTO } from '../types'

export const useAppStore = defineStore('app', () => {
  // ── Schema (loaded once) ──────────────────────────────────────────────────
  const schema = ref<SectionSchemaDTO[]>([])
  const conditionTypes = ref<string[]>([])

  // ── Global config ─────────────────────────────────────────────────────────
  const globalPath = ref('')
  const globalBlocks = ref<Block[]>([])
  const globalUnsaved = ref(false)

  // ── Project config ────────────────────────────────────────────────────────
  const projectPath = ref('')
  const projectBlocks = ref<Block[]>([])
  const projectUnsaved = ref(false)
  const recentRepos = ref<string[]>([])

  // ── UI state ──────────────────────────────────────────────────────────────
  const status = ref('Ready.')
  const theme = ref<'light' | 'dark'>('light')

  // ── Computed ──────────────────────────────────────────────────────────────

  /** Set of "sectionName\0subsection\0key" strings that the project config overrides */
  const supersededKeys = computed(() => {
    const keys = new Set<string>()
    for (const block of projectBlocks.value) {
      if (block.type === 'section' || block.type === 'custom') {
        for (const entry of block.entries) {
          if (entry.key) {
            keys.add(`${block.schemaName}\0${block.subsection}\0${entry.key}`)
          }
        }
      }
    }
    return keys
  })

  const schemaByName = computed(() => {
    const map: Record<string, SectionSchemaDTO> = {}
    for (const s of schema.value) map[s.name] = s
    return map
  })

  // ── Actions ───────────────────────────────────────────────────────────────

  async function init() {
    const [s, ct, path, recent] = await Promise.all([
      GetSchema(),
      GetConditionTypes(),
      GetGlobalConfigPath(),
      GetRecentRepos(),
    ])
    schema.value = s
    conditionTypes.value = ct
    recentRepos.value = recent
    globalPath.value = path
    await loadGlobal(path)
  }

  async function loadGlobal(path: string) {
    try {
      const dtos = await LoadConfig(path)
      globalBlocks.value = dtos.map(dtoToBlock)
      globalPath.value = path
      globalUnsaved.value = false
      setStatus('Loaded ' + path)
    } catch (e: any) {
      setStatus('Error loading global config: ' + e)
    }
  }

  async function saveGlobal() {
    if (!globalPath.value) return
    try {
      await SaveConfig(globalPath.value, globalBlocks.value.map(blockToDTO))
      globalUnsaved.value = false
      setStatus('Saved ' + globalPath.value)
    } catch (e: any) {
      setStatus('Save failed: ' + e)
    }
  }

  async function openProject() {
    const folder = await SelectProjectFolder()
    if (!folder) return
    const gitConfigPath = folder + '/.git/config'
    await loadProject(gitConfigPath, folder)
  }

  async function loadProject(configPath: string, repoPath?: string) {
    try {
      const dtos = await LoadConfig(configPath)
      projectBlocks.value = dtos.map(dtoToBlock)
      projectPath.value = configPath
      projectUnsaved.value = false
      const repo = repoPath ?? configPath.replace('/.git/config', '')
      recentRepos.value = await (async () => {
        await AddRecentRepo(repo)
        return GetRecentRepos()
      })()
      setStatus('Loaded project: ' + repo)
    } catch (e: any) {
      setStatus('Error loading project config: ' + e)
    }
  }

  async function saveProject() {
    if (!projectPath.value) return
    try {
      await SaveConfig(projectPath.value, projectBlocks.value.map(blockToDTO))
      projectUnsaved.value = false
      setStatus('Saved ' + projectPath.value)
    } catch (e: any) {
      setStatus('Save failed: ' + e)
    }
  }

  async function resetGlobal() {
    if (!globalPath.value) return
    await loadGlobal(globalPath.value)
  }

  async function resetProject() {
    if (!projectPath.value) return
    await loadProject(projectPath.value)
  }

  function markGlobalUnsaved() { globalUnsaved.value = true }
  function markProjectUnsaved() { projectUnsaved.value = true }

  function toggleTheme() {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
    document.documentElement.classList.toggle('dark', theme.value === 'dark')
  }

  function setStatus(msg: string) { status.value = msg }

  return {
    schema, conditionTypes, schemaByName,
    globalPath, globalBlocks, globalUnsaved,
    projectPath, projectBlocks, projectUnsaved,
    recentRepos, supersededKeys,
    status, theme,
    init, loadGlobal, saveGlobal, resetGlobal,
    openProject, loadProject, saveProject, resetProject,
    markGlobalUnsaved, markProjectUnsaved,
    toggleTheme, setStatus,
  }
})
