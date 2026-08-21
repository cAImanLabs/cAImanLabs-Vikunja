// Fixed-order categorical palette for per-user chart series. Order is the
// CVD-safety mechanism (validated for adjacent pairs, e.g. pie wedges and
// neighboring bars) - never reorder or cycle past its length. A 9th series
// folds into "Other" instead of generating a new hue.
export const CATEGORICAL_PALETTE = [
	'#2a78d6', // blue
	'#eb6834', // orange
	'#1baf7a', // aqua
	'#eda100', // yellow
	'#e87ba4', // magenta
	'#008300', // green
	'#4a3aa7', // violet
	'#e34948', // red
]

export const CATEGORICAL_OTHER_COLOR = '#9a9a9a'

export function getCategoricalColor(index: number): string {
	return index < CATEGORICAL_PALETTE.length ? CATEGORICAL_PALETTE[index] : CATEGORICAL_OTHER_COLOR
}
