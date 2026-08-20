import {TASK_KINDS, type TaskKind} from '@/modelTypes/ITaskKind'

// Single source of truth for how each task kind looks - icon, color and
// i18n key - shared between TaskKindLabel (read-only badges) and
// TaskKindSelect (the picker), so the trigger/options and any badge always
// show the exact same icon+color as each other.
export const TASK_KIND_ICONS = {
	[TASK_KINDS.TASK]: 'tasks',
	[TASK_KINDS.EPIC]: 'bolt',
	[TASK_KINDS.STORY]: 'bookmark',
	[TASK_KINDS.BUG]: 'bug',
	[TASK_KINDS.SUBTASK]: 'arrow-turn-down',
	[TASK_KINDS.FEATURE]: 'layer-group',
	[TASK_KINDS.INITIATIVE]: 'bullseye',
} as const satisfies Record<TaskKind, string>

export const TASK_KIND_COLORS: Record<TaskKind, string> = {
	[TASK_KINDS.TASK]: 'var(--grey-500)',
	[TASK_KINDS.EPIC]: 'var(--primary)',
	[TASK_KINDS.STORY]: 'var(--success)',
	[TASK_KINDS.BUG]: 'var(--danger-text)',
	[TASK_KINDS.SUBTASK]: 'var(--grey-500)',
	[TASK_KINDS.FEATURE]: 'var(--info)',
	[TASK_KINDS.INITIATIVE]: 'var(--link-visited)',
}

export const TASK_KIND_I18N_KEYS: Record<TaskKind, string> = {
	[TASK_KINDS.TASK]: 'task.kind.task',
	[TASK_KINDS.EPIC]: 'task.kind.epic',
	[TASK_KINDS.STORY]: 'task.kind.story',
	[TASK_KINDS.BUG]: 'task.kind.bug',
	[TASK_KINDS.SUBTASK]: 'task.kind.subtask',
	[TASK_KINDS.FEATURE]: 'task.kind.feature',
	[TASK_KINDS.INITIATIVE]: 'task.kind.initiative',
}
