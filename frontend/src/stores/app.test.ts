import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAppStore } from './app'
import type { Block } from '../types'

// ── Mock the Wails bridge ─────────────────────────────────────────────────────
// The wailsjs bindings call window['go'] which doesn't exist in jsdom.
// All bridge functions are replaced with controllable vi.fn() stubs.

vi.mock('../../wailsjs/go/main/App', () => ({
  LoadConfig: vi.fn().mockResolvedValue([]),
  SaveConfig: vi.fn().mockResolvedValue(undefined),
  GetSchema: vi.fn().mockResolvedValue([]),
  GetConditionTypes: vi.fn().mockResolvedValue([]),
  GetGlobalConfigPath: vi.fn().mockResolvedValue('/home/user/.gitconfig'),
  GetRecentRepos: vi.fn().mockResolvedValue([]),
  AddRecentRepo: vi.fn().mockResolvedValue(undefined),
  SelectProjectFolder: vi.fn().mockResolvedValue(''),
}))

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeBlock(overrides: Partial<Block> = {}): Block {
  return {
    id: 'b1',
    type: 'section',
    schemaName: 'user',
    subsection: '',
    entries: [],
    path: '',
    children: [],
    conditions: [],
    file: null,
    ...overrides,
  }
}

// ── Test suite ────────────────────────────────────────────────────────────────

describe('useAppStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    // Reset the dark-mode class between tests
    document.documentElement.classList.remove('dark')
  })

  // ── supersededKeys ──────────────────────────────────────────────────────────

  describe('supersededKeys', () => {
    it('is empty when there are no project blocks', () => {
      const store = useAppStore()
      expect(store.supersededKeys.size).toBe(0)
    })

    it('includes keys from section blocks in the project config', () => {
      const store = useAppStore()
      store.projectBlocks = [
        makeBlock({
          schemaName: 'user',
          subsection: '',
          entries: [{ key: 'email', value: 'work@example.com', superseded: false }],
        }),
      ]
      expect(store.supersededKeys.has('user\0\0email')).toBe(true)
    })

    it('encodes schemaName, subsection, and key separated by null bytes', () => {
      const store = useAppStore()
      store.projectBlocks = [
        makeBlock({
          schemaName: 'branch',
          subsection: 'main',
          entries: [{ key: 'rebase', value: 'true', superseded: false }],
        }),
      ]
      expect(store.supersededKeys.has('branch\0main\0rebase')).toBe(true)
    })

    it('includes keys from custom blocks', () => {
      const store = useAppStore()
      store.projectBlocks = [
        makeBlock({
          type: 'custom',
          schemaName: 'myplugin',
          subsection: '',
          entries: [{ key: 'opt', value: 'on', superseded: false }],
        }),
      ]
      expect(store.supersededKeys.has('myplugin\0\0opt')).toBe(true)
    })

    it('ignores file blocks', () => {
      const store = useAppStore()
      store.projectBlocks = [
        makeBlock({
          type: 'file',
          path: '~/.gitconfig_extra',
          entries: [{ key: 'path', value: '~/.gitconfig_extra', superseded: false }],
        }),
      ]
      expect(store.supersededKeys.size).toBe(0)
    })

    it('ignores conditional blocks', () => {
      const store = useAppStore()
      store.projectBlocks = [
        makeBlock({
          type: 'conditional',
          path: '~/.gitconfig_work',
          conditions: [{ kind: 'gitdir', value: '~/work/' }],
        }),
      ]
      expect(store.supersededKeys.size).toBe(0)
    })

    it('ignores entries with empty keys (comment lines)', () => {
      const store = useAppStore()
      store.projectBlocks = [
        makeBlock({
          entries: [{ key: '', value: '# a comment', superseded: false }],
        }),
      ]
      expect(store.supersededKeys.size).toBe(0)
    })

    it('accumulates keys from multiple project blocks', () => {
      const store = useAppStore()
      store.projectBlocks = [
        makeBlock({
          schemaName: 'user',
          entries: [{ key: 'name', value: 'Alice', superseded: false }],
        }),
        makeBlock({
          id: 'b2',
          schemaName: 'core',
          entries: [{ key: 'autocrlf', value: 'true', superseded: false }],
        }),
      ]
      expect(store.supersededKeys.has('user\0\0name')).toBe(true)
      expect(store.supersededKeys.has('core\0\0autocrlf')).toBe(true)
      expect(store.supersededKeys.size).toBe(2)
    })
  })

  // ── schemaByName ────────────────────────────────────────────────────────────

  describe('schemaByName', () => {
    it('returns an empty object when schema is not loaded', () => {
      const store = useAppStore()
      expect(store.schemaByName).toEqual({})
    })

    it('builds a lookup map keyed by schema name', () => {
      const store = useAppStore()
      store.schema = [
        { name: 'user', label: 'User', icon: '👤', hasSubsection: false, fields: [], freeForm: false },
        { name: 'core', label: 'Core', icon: '⚙️', hasSubsection: false, fields: [], freeForm: false },
      ] as any
      expect(store.schemaByName['user'].label).toBe('User')
      expect(store.schemaByName['core'].label).toBe('Core')
      expect(store.schemaByName['remote']).toBeUndefined()
    })
  })

  // ── toggleTheme ──────────────────────────────────────────────────────────────

  describe('toggleTheme', () => {
    it('starts in light mode', () => {
      const store = useAppStore()
      expect(store.theme).toBe('light')
    })

    it('toggles from light to dark', () => {
      const store = useAppStore()
      store.toggleTheme()
      expect(store.theme).toBe('dark')
      expect(document.documentElement.classList.contains('dark')).toBe(true)
    })

    it('toggles from dark back to light', () => {
      const store = useAppStore()
      store.toggleTheme() // → dark
      store.toggleTheme() // → light
      expect(store.theme).toBe('light')
      expect(document.documentElement.classList.contains('dark')).toBe(false)
    })
  })

  // ── setStatus ────────────────────────────────────────────────────────────────

  describe('setStatus', () => {
    it('starts at "Ready."', () => {
      const store = useAppStore()
      expect(store.status).toBe('Ready.')
    })

    it('updates the status message', () => {
      const store = useAppStore()
      store.setStatus('Saved successfully')
      expect(store.status).toBe('Saved successfully')
    })

    it('can be called repeatedly', () => {
      const store = useAppStore()
      store.setStatus('First')
      store.setStatus('Second')
      expect(store.status).toBe('Second')
    })
  })

  // ── markGlobalUnsaved / markProjectUnsaved ───────────────────────────────────

  describe('unsaved flags', () => {
    it('globalUnsaved starts false', () => {
      const store = useAppStore()
      expect(store.globalUnsaved).toBe(false)
    })

    it('markGlobalUnsaved sets globalUnsaved to true', () => {
      const store = useAppStore()
      store.markGlobalUnsaved()
      expect(store.globalUnsaved).toBe(true)
    })

    it('projectUnsaved starts false', () => {
      const store = useAppStore()
      expect(store.projectUnsaved).toBe(false)
    })

    it('markProjectUnsaved sets projectUnsaved to true', () => {
      const store = useAppStore()
      store.markProjectUnsaved()
      expect(store.projectUnsaved).toBe(true)
    })
  })

  // ── loadGlobal ──────────────────────────────────────────────────────────────

  describe('loadGlobal', () => {
    it('populates globalBlocks and clears unsaved flag', async () => {
      const mocks = await import('../../wailsjs/go/main/App')
      vi.mocked(mocks.LoadConfig).mockResolvedValueOnce([
        {
          id: '1', type: 'section', schemaName: 'user', subsection: '',
          entries: [{ key: 'name', value: 'Alice', rawLine: '' }],
          path: '', children: [], conditions: [], file: null,
        },
      ] as any)

      const store = useAppStore()
      store.markGlobalUnsaved()
      await store.loadGlobal('/home/user/.gitconfig')

      expect(store.globalBlocks).toHaveLength(1)
      expect(store.globalBlocks[0].schemaName).toBe('user')
      expect(store.globalUnsaved).toBe(false)
    })

    it('updates globalPath', async () => {
      const store = useAppStore()
      await store.loadGlobal('/custom/.gitconfig')
      expect(store.globalPath).toBe('/custom/.gitconfig')
    })

    it('sets an error status when LoadConfig rejects', async () => {
      const mocks = await import('../../wailsjs/go/main/App')
      vi.mocked(mocks.LoadConfig).mockRejectedValueOnce(new Error('file not found'))

      const store = useAppStore()
      await store.loadGlobal('/bad/path')

      expect(store.status).toMatch(/Error loading global config/)
    })
  })

  // ── resetGlobal ──────────────────────────────────────────────────────────────

  describe('resetGlobal', () => {
    it('does nothing when globalPath is empty', async () => {
      const store = useAppStore()
      await store.resetGlobal() // should not throw
    })

    it('does not wipe existing blocks when backend returns empty', async () => {
      // This guards against a transient Wails bridge restart returning nothing.
      const mocks = await import('../../wailsjs/go/main/App')
      vi.mocked(mocks.LoadConfig).mockResolvedValueOnce([])

      const store = useAppStore()
      store.globalPath = '/home/user/.gitconfig'
      store.globalBlocks = [makeBlock()]

      await store.resetGlobal()

      expect(store.globalBlocks).toHaveLength(1) // not wiped
      expect(store.status).toMatch(/skipped/)
    })

    it('replaces blocks when backend returns content', async () => {
      const mocks = await import('../../wailsjs/go/main/App')
      vi.mocked(mocks.LoadConfig).mockResolvedValueOnce([
        {
          id: '99', type: 'section', schemaName: 'core', subsection: '',
          entries: [], path: '', children: [], conditions: [], file: null,
        },
      ] as any)

      const store = useAppStore()
      store.globalPath = '/home/user/.gitconfig'
      store.globalBlocks = [makeBlock()]

      await store.resetGlobal()

      expect(store.globalBlocks).toHaveLength(1)
      expect(store.globalBlocks[0].schemaName).toBe('core')
    })
  })
})
