import type {IUserTaskCompletionStat} from '@/modelTypes/IUserTaskCompletionStat'
import {getCategoricalColor, CATEGORICAL_OTHER_COLOR} from '@/helpers/categoricalPalette'

export interface FoldedUserSeriesEntry {
	key: string
	label: string
	value: number
	color: string
}

// Beyond this many series the categorical palette can no longer guarantee
// adjacent-pair CVD separation, so the tail folds into a single "Other" entry
// instead of generating a new hue.
const MAX_SERIES = 8

/**
 * Folds a per-user stat list (already ordered by the backend) into at most
 * MAX_SERIES named series plus one "Other" bucket. Every caller passes the
 * same `stats` order, so a given user gets the same color across every chart
 * reading it - even a chart keyed on a different value (e.g. story points)
 * stays in completed-count order rather than re-sorting and reassigning colors.
 */
export function foldUserSeries(
	stats: IUserTaskCompletionStat[],
	valueOf: (s: IUserTaskCompletionStat) => number,
	otherLabel: string,
): FoldedUserSeriesEntry[] {
	const shown = stats.slice(0, MAX_SERIES)
	const rest = stats.slice(MAX_SERIES)
	const restValue = rest.reduce((sum, s) => sum + valueOf(s), 0)

	const entries: FoldedUserSeriesEntry[] = shown.map((s, i) => ({
		key: `user-${s.userId}`,
		label: s.username,
		value: valueOf(s),
		color: getCategoricalColor(i),
	}))
	if (restValue > 0) {
		entries.push({key: 'other', label: otherLabel, value: restValue, color: CATEGORICAL_OTHER_COLOR})
	}
	return entries
}
