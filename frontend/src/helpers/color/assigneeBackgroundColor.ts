import type {IUser} from '@/modelTypes/IUser'

/**
 * The color of the first assignee that has one set, as a plain hex string
 * (e.g. "#ff7a00"). Returns undefined when no assignee has a color.
 */
export function getFirstAssigneeColor(assignees: IUser[] | undefined): string | undefined {
	return assignees?.find(a => a.color)?.color || undefined
}

// Translucent so task text/badges stay readable regardless of how light or
// dark the assignee's chosen color is, and so it works in both light and dark theme.
const ASSIGNEE_TINT_ALPHA = 0.18

/**
 * The background tint for a task row based on its first colored assignee.
 * Returns undefined when no assignee has a color set (row keeps its default background).
 */
export function getAssigneeBackgroundColor(assignees: IUser[] | undefined): string | undefined {
	const color = getFirstAssigneeColor(assignees)
	if (!color) {
		return undefined
	}

	const hex = color.startsWith('#') ? color.slice(1) : color
	if (hex.length !== 6) {
		return undefined
	}

	const r = parseInt(hex.slice(0, 2), 16)
	const g = parseInt(hex.slice(2, 4), 16)
	const b = parseInt(hex.slice(4, 6), 16)
	if ([r, g, b].some(Number.isNaN)) {
		return undefined
	}

	return `rgba(${r}, ${g}, ${b}, ${ASSIGNEE_TINT_ALPHA})`
}
