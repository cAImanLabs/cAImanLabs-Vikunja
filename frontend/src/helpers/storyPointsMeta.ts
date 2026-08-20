// Single source of truth for how each story point value looks - options and
// an escalating color ramp by effort size - shared with StoryPointsLabel so
// the picker and any badge always show the exact same color.
export const STORY_POINTS_VALUES = [0, 1, 2, 3, 5, 8, 13, 21] as const

// CSS custom property (with var()) for each value's color, escalating from
// calm to critical as the estimated effort grows. undefined means "no color
// override, use the default text color" (used for 0/unestimated).
export const STORY_POINTS_COLORS: Record<number, string | undefined> = {
	0: undefined,
	1: 'var(--success)',
	2: 'var(--success)',
	3: 'var(--warning)',
	5: 'var(--warning)',
	8: 'var(--warning-dark)',
	13: 'var(--danger-text)',
	21: 'var(--danger-dark)',
}
