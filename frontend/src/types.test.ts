import { describe, it, expect } from 'vitest'
import { dtoToBlock, blockToDTO, type Block } from './types'

// ── Helpers ───────────────────────────────────────────────────────────────────

/** Minimal valid BlockDTO shape — override individual fields per test. */
function makeDTO(overrides: Record<string, unknown> = {}) {
  return {
    id: 'test-id',
    type: 'section',
    schemaName: 'user',
    subsection: '',
    entries: [],
    path: '',
    children: [],
    conditions: [],
    file: null,
    ...overrides,
  } as any
}

/** Minimal valid Block — override individual fields per test. */
function makeBlock(overrides: Partial<Block> = {}): Block {
  return {
    id: 'test-id',
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

// ── dtoToBlock ────────────────────────────────────────────────────────────────

describe('dtoToBlock', () => {
  it('copies id and type', () => {
    const block = dtoToBlock(makeDTO({ id: 'abc-123', type: 'custom' }))
    expect(block.id).toBe('abc-123')
    expect(block.type).toBe('custom')
  })

  it('defaults missing optional fields', () => {
    const block = dtoToBlock({ id: 'x', type: 'section' } as any)
    expect(block.schemaName).toBe('')
    expect(block.subsection).toBe('')
    expect(block.entries).toEqual([])
    expect(block.path).toBe('')
    expect(block.children).toEqual([])
    expect(block.conditions).toEqual([])
    expect(block.file).toBeNull()
  })

  it('sets superseded: false on every entry', () => {
    const block = dtoToBlock(makeDTO({
      entries: [
        { key: 'name', value: 'Alice', rawLine: '' },
        { key: 'email', value: 'alice@example.com', rawLine: '' },
      ],
    }))
    expect(block.entries).toHaveLength(2)
    for (const entry of block.entries) {
      expect(entry.superseded).toBe(false)
    }
  })

  it('preserves entry key, value, and rawLine', () => {
    const block = dtoToBlock(makeDTO({
      entries: [{ key: 'autocrlf', value: 'input', rawLine: '\tautocrlf = input' }],
    }))
    expect(block.entries[0].key).toBe('autocrlf')
    expect(block.entries[0].value).toBe('input')
    expect(block.entries[0].rawLine).toBe('\tautocrlf = input')
  })

  it('recursively converts children', () => {
    const block = dtoToBlock(makeDTO({
      children: [
        makeDTO({ id: 'child-1', type: 'custom', schemaName: 'myplugin' }),
      ],
    }))
    expect(block.children).toHaveLength(1)
    expect(block.children[0].id).toBe('child-1')
    expect(block.children[0].type).toBe('custom')
    expect(block.children[0].schemaName).toBe('myplugin')
    // Children of children also get superseded:false on entries
  })

  it('converts a nested file block', () => {
    const block = dtoToBlock(makeDTO({
      type: 'conditional',
      file: makeDTO({ id: 'file-1', type: 'file', path: '~/.gitconfig_work' }),
    }))
    expect(block.file).not.toBeNull()
    expect(block.file!.id).toBe('file-1')
    expect(block.file!.type).toBe('file')
    expect(block.file!.path).toBe('~/.gitconfig_work')
  })

  it('sets file to null when DTO has no file', () => {
    const block = dtoToBlock(makeDTO({ file: null }))
    expect(block.file).toBeNull()
  })

  it('converts conditions', () => {
    const block = dtoToBlock(makeDTO({
      type: 'conditional',
      conditions: [
        { kind: 'gitdir', value: '~/work/' },
        { kind: 'onbranch', value: 'main' },
      ],
    }))
    expect(block.conditions).toHaveLength(2)
    expect(block.conditions[0]).toEqual({ kind: 'gitdir', value: '~/work/' })
    expect(block.conditions[1]).toEqual({ kind: 'onbranch', value: 'main' })
  })
})

// ── blockToDTO ────────────────────────────────────────────────────────────────

describe('blockToDTO', () => {
  it('copies id and type', () => {
    const dto = blockToDTO(makeBlock({ id: 'xyz', type: 'file' }))
    expect(dto.id).toBe('xyz')
    expect(dto.type).toBe('file')
  })

  it('converts empty arrays to undefined', () => {
    const dto = blockToDTO(makeBlock())
    expect(dto.entries).toBeUndefined()
    expect(dto.children).toBeUndefined()
    expect(dto.conditions).toBeUndefined()
  })

  it('converts empty strings to undefined', () => {
    const dto = blockToDTO(makeBlock({ schemaName: '', subsection: '', path: '' }))
    expect(dto.schemaName).toBeUndefined()
    expect(dto.subsection).toBeUndefined()
    expect(dto.path).toBeUndefined()
  })

  it('preserves non-empty schemaName and subsection', () => {
    const dto = blockToDTO(makeBlock({ schemaName: 'remote', subsection: 'origin' }))
    expect(dto.schemaName).toBe('remote')
    expect(dto.subsection).toBe('origin')
  })

  it('strips superseded from entries', () => {
    const dto = blockToDTO(makeBlock({
      entries: [{ key: 'name', value: 'Alice', superseded: true }],
    }))
    const entries = dto.entries as any[]
    expect(entries).toHaveLength(1)
    expect('superseded' in entries[0]).toBe(false)
  })

  it('preserves entry key, value, and rawLine in output', () => {
    const dto = blockToDTO(makeBlock({
      entries: [{ key: 'bare', value: 'false', rawLine: '\tbare = false' }],
    }))
    const entries = dto.entries as any[]
    expect(entries[0].key).toBe('bare')
    expect(entries[0].value).toBe('false')
    expect(entries[0].rawLine).toBe('\tbare = false')
  })

  it('recursively converts children', () => {
    const dto = blockToDTO(makeBlock({
      children: [makeBlock({ id: 'child-1', schemaName: 'core' })],
    }))
    const children = dto.children as any[]
    expect(children).toHaveLength(1)
    expect(children[0].id).toBe('child-1')
  })

  it('converts file to undefined when null', () => {
    const dto = blockToDTO(makeBlock({ file: null }))
    expect(dto.file).toBeUndefined()
  })

  it('recursively converts nested file block', () => {
    const dto = blockToDTO(makeBlock({
      type: 'conditional',
      file: makeBlock({ id: 'file-block', type: 'file', path: '~/.work' }),
    }))
    const file = dto.file as any
    expect(file).not.toBeUndefined()
    expect(file.id).toBe('file-block')
  })
})

// ── Round-trip ────────────────────────────────────────────────────────────────

describe('dtoToBlock / blockToDTO round-trip', () => {
  it('preserves core fields for a section block', () => {
    const original = makeDTO({
      id: 'rt-1',
      type: 'section',
      schemaName: 'core',
      subsection: '',
      entries: [{ key: 'autocrlf', value: 'input', rawLine: '' }],
    })
    const dto = blockToDTO(dtoToBlock(original))
    expect(dto.id).toBe('rt-1')
    expect(dto.type).toBe('section')
    expect(dto.schemaName).toBe('core')
    const entries = dto.entries as any[]
    expect(entries[0].key).toBe('autocrlf')
    expect(entries[0].value).toBe('input')
  })

  it('preserves conditions on a conditional block', () => {
    const original = makeDTO({
      id: 'rt-2',
      type: 'conditional',
      path: '~/.gitconfig_work',
      conditions: [{ kind: 'gitdir', value: '~/work/' }],
    })
    const dto = blockToDTO(dtoToBlock(original))
    expect(dto.type).toBe('conditional')
    const conditions = dto.conditions as any[]
    expect(conditions[0]).toEqual({ kind: 'gitdir', value: '~/work/' })
  })

  it('preserves subsection on a branch block', () => {
    const original = makeDTO({
      id: 'rt-3',
      type: 'section',
      schemaName: 'branch',
      subsection: 'main',
      entries: [{ key: 'rebase', value: 'true', rawLine: '' }],
    })
    const dto = blockToDTO(dtoToBlock(original))
    expect(dto.subsection).toBe('main')
  })
})
