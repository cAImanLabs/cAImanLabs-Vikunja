import {PRIORITIES, type Priority} from '@/constants/priorities'

// Single source of truth for how each priority level looks - icon, color and
// i18n key - shared between PriorityLabel (read-only badges) and
// PrioritySelect (the picker), so the trigger/options and any badge always
// show the exact same icon+color as each other.
export const PRIORITY_ICONS = {
	[PRIORITIES.UNSET]: 'circle',
	[PRIORITIES.LOW]: 'arrow-down',
	[PRIORITIES.MEDIUM]: 'equals',
	[PRIORITIES.HIGH]: 'arrow-up',
	[PRIORITIES.URGENT]: 'angles-up',
	[PRIORITIES.DO_NOW]: 'fire',
} as const satisfies Record<Priority, string>

// CSS custom property (with var()) for each level's color, escalating from
// calm to critical. undefined means "no color override, use the default
// text color" (used for Unset).
export const PRIORITY_COLORS: Record<Priority, string | undefined> = {
	[PRIORITIES.UNSET]: undefined,
	[PRIORITIES.LOW]: 'var(--success)',
	[PRIORITIES.MEDIUM]: 'var(--warning)',
	[PRIORITIES.HIGH]: 'var(--warning-dark)',
	[PRIORITIES.URGENT]: 'var(--danger-text)',
	[PRIORITIES.DO_NOW]: 'var(--danger-dark)',
}

export const PRIORITY_I18N_KEYS: Record<Priority, string> = {
	[PRIORITIES.UNSET]: 'task.priority.unset',
	[PRIORITIES.LOW]: 'task.priority.low',
	[PRIORITIES.MEDIUM]: 'task.priority.medium',
	[PRIORITIES.HIGH]: 'task.priority.high',
	[PRIORITIES.URGENT]: 'task.priority.urgent',
	[PRIORITIES.DO_NOW]: 'task.priority.doNow',
}

export const PRIORITY_ORDER: Priority[] = [
	PRIORITIES.UNSET,
	PRIORITIES.LOW,
	PRIORITIES.MEDIUM,
	PRIORITIES.HIGH,
	PRIORITIES.URGENT,
	PRIORITIES.DO_NOW,
]
