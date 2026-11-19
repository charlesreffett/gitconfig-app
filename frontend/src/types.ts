// ── Wire types — owned by the Wails auto-generated models, not duplicated here ─
// The generator produces classes (with convertValues) and uses loose `string`
// for discriminant fields. We re-export them as type aliases so the rest of the
// codebase can import from one place.

import type { main } from '../wailsjs/go/models'

export type BlockDTO = main.BlockDTO
export type EntryDTO = main.EntryDTO
export type ConditionDTO = main.ConditionDTO
export type FieldDTO = main.FieldDTO
export type SectionSchemaDTO = main.SectionSchemaDTO

// ── Richer client-side types ──────────────────────────────────────────────────
// These are plain interfaces (no convertValues) used in the Vue store and
// components. Literal-union `type` fields are safe here because we control
// every place that creates a Block.

export interface Entry {
  key: string
  value: string
  rawLine?: string
  superseded?: boolean
}

export interface Block {
  id: string
  type: 'section' | 'file' | 'conditional' | 'custom'
  // section + custom
  schemaName: string
  subsection: string
  entries: Entry[]
  // file
  path: string
  children: Block[]
  // conditional
  conditions: ConditionDTO[]
  file: Block | null
}

// ── Converters ────────────────────────────────────────────────────────────────

/** Convert a Go DTO (auto-generated class) into the richer client Block. */
export function dtoToBlock(dto: BlockDTO): Block {
  return {
    id: dto.id,
    // The generated class uses `string`; we know it's always one of our four
    // values because Go sends exactly those strings.
    type: dto.type as Block['type'],
    schemaName: dto.schemaName ?? '',
    subsection: dto.subsection ?? '',
    entries: (dto.entries ?? []).map(e => ({
      key: e.key,
      value: e.value,
      rawLine: e.rawLine,
      superseded: false,
    })),
    path: dto.path ?? '',
    children: (dto.children ?? []).map(dtoToBlock),
    conditions: (dto.conditions ?? []).map(c => ({ kind: c.kind, value: c.value })),
    file: dto.file ? dtoToBlock(dto.file) : null,
  }
}

/** Convert a client Block back to a plain DTO object for SaveConfig.
 *
 * Wails serialises the argument to JSON before sending to Go, so the plain
 * object shape is sufficient — the `convertValues` method on the generated
 * class is only used for deserialization (Go → JS), never for sending data.
 * We cast through `unknown` to satisfy TypeScript without the class methods.
 */
export function blockToDTO(block: Block): BlockDTO {
  return {
    id: block.id,
    type: block.type,
    schemaName: block.schemaName || undefined,
    subsection: block.subsection || undefined,
    entries: block.entries.length > 0
      ? block.entries.map(e => ({ key: e.key, value: e.value, rawLine: e.rawLine }))
      : undefined,
    path: block.path || undefined,
    children: block.children.length > 0 ? block.children.map(blockToDTO) : undefined,
    conditions: block.conditions.length > 0
      ? block.conditions.map(c => ({ kind: c.kind, value: c.value }))
      : undefined,
    file: block.file ? blockToDTO(block.file) : undefined,
  } as unknown as BlockDTO
}
