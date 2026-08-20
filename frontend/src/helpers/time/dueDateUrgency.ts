export type DueDateUrgency = 'overdue' | 'today' | 'soon' | null

const ONE_DAY_MS = 24 * 60 * 60 * 1000
const DUE_SOON_DAYS = 3

// A tiered read on how close a due date is: red once overdue, then two
// lighter warning shades as it approaches, so urgency is visible without
// opening a date picker. Shared between the task detail view's due date
// field and the due date shown on task rows in list views.
export function getDueDateUrgency(dueDate: Date | null, done: boolean, now: Date): DueDateUrgency {
	if (done || !dueDate) {
		return null
	}

	const due = dueDate.getTime()
	if (due === 0) {
		return null
	}

	const nowMs = now.getTime()
	if (due <= nowMs) {
		return 'overdue'
	}
	if (due <= nowMs + ONE_DAY_MS) {
		return 'today'
	}
	if (due <= nowMs + DUE_SOON_DAYS * ONE_DAY_MS) {
		return 'soon'
	}
	return null
}
