<template>
	<div class="burndown-chart">
		<Nothing v-if="burndown.series.length === 0">
			{{ $t('project.sprint.burndownEmpty') }}
		</Nothing>
		<template v-else>
			<p
				v-if="burndown.totalPoints === 0"
				class="burndown-chart-hint has-text-grey is-size-7"
			>
				{{ $t('project.sprint.burndownEmpty') }}
			</p>
			<div class="burndown-chart-legend">
				<span class="legend-item">
					<span class="legend-swatch is-ideal" />
					{{ $t('project.sprint.burndownIdeal') }}
				</span>
				<span class="legend-item">
					<span class="legend-swatch is-actual" />
					{{ $t('project.sprint.burndownActual') }}
				</span>
			</div>

			<svg
				:viewBox="`0 0 ${WIDTH} ${HEIGHT}`"
				class="burndown-chart-svg"
				role="img"
				:aria-label="$t('project.sprint.burndownAriaLabel', {points: burndown.totalPoints})"
				@mousemove="onMouseMove"
				@mouseleave="hoverIndex = null"
			>
				<!-- Y gridlines / labels -->
				<g class="axis-y">
					<template
						v-for="tick in yTicks"
						:key="tick"
					>
						<line
							:x1="MARGIN.left"
							:x2="WIDTH - MARGIN.right"
							:y1="yScale(tick)"
							:y2="yScale(tick)"
							class="gridline"
						/>
						<text
							:x="MARGIN.left - 8"
							:y="yScale(tick)"
							class="axis-label"
							text-anchor="end"
							dominant-baseline="middle"
						>{{ tick }}</text>
					</template>
				</g>

				<!-- X labels: sprint start and end -->
				<text
					:x="MARGIN.left"
					:y="HEIGHT - 8"
					class="axis-label"
					text-anchor="start"
				>{{ formatDisplayDate(burndown.startDate) }}</text>
				<text
					:x="WIDTH - MARGIN.right"
					:y="HEIGHT - 8"
					class="axis-label"
					text-anchor="end"
				>{{ formatDisplayDate(burndown.endDate) }}</text>

				<!-- Ideal line -->
				<line
					:x1="xScale(burndown.startDate)"
					:y1="yScale(burndown.totalPoints)"
					:x2="xScale(burndown.endDate)"
					:y2="yScale(0)"
					class="ideal-line"
				/>

				<!-- Actual line -->
				<polyline
					:points="actualLinePoints"
					class="actual-line"
				/>
				<circle
					v-for="(p, i) in burndown.series"
					:key="i"
					:cx="xScale(p.date)"
					:cy="yScale(p.remaining)"
					:r="hoverIndex === i ? 5 : 3"
					class="actual-point"
				/>

				<!-- Hover crosshair -->
				<line
					v-if="hoverIndex !== null"
					:x1="xScale(burndown.series[hoverIndex].date)"
					:x2="xScale(burndown.series[hoverIndex].date)"
					:y1="MARGIN.top"
					:y2="HEIGHT - MARGIN.bottom"
					class="crosshair"
				/>

				<!-- Hover hit area -->
				<rect
					:x="MARGIN.left"
					:y="MARGIN.top"
					:width="plotWidth"
					:height="plotHeight"
					fill="transparent"
				/>
			</svg>

			<div
				v-if="hoverIndex !== null"
				class="burndown-chart-tooltip"
			>
				<strong>{{ formatDisplayDate(burndown.series[hoverIndex].date) }}</strong>
				{{ $t('project.sprint.burndownTooltip', {points: burndown.series[hoverIndex].remaining}) }}
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'

import Nothing from '@/components/misc/Nothing.vue'
import {formatDisplayDate} from '@/helpers/time/formatDate'

import type {ISprintBurndown} from '@/modelTypes/ISprint'

const props = defineProps<{
	burndown: ISprintBurndown,
}>()

const WIDTH = 640
const HEIGHT = 280
const MARGIN = {top: 16, right: 16, bottom: 32, left: 40}

const plotWidth = WIDTH - MARGIN.left - MARGIN.right
const plotHeight = HEIGHT - MARGIN.top - MARGIN.bottom

const maxY = computed(() => Math.max(props.burndown.totalPoints, 1))

const yTicks = computed(() => {
	const max = props.burndown.totalPoints
	if (max <= 0) {
		return [0]
	}
	const step = Math.max(1, Math.round(max / 4))
	const ticks = []
	for (let v = 0; v <= max; v += step) {
		ticks.push(v)
	}
	if (ticks[ticks.length - 1] !== max) {
		ticks.push(max)
	}
	return ticks
})

// Truncated to day boundaries to match the series' points (see truncateToDay
// server-side) - the sprint's raw start/end carry a time-of-day component
// (whenever the sprint was created/edited), which otherwise bunches the first
// day's point against the axis instead of spacing all days evenly.
function truncateToDay(date: Date): number {
	return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime()
}

function xScale(date: Date): number {
	const start = truncateToDay(props.burndown.startDate)
	const end = truncateToDay(props.burndown.endDate)
	const span = end - start || 1
	const ratio = (truncateToDay(date) - start) / span
	return MARGIN.left + Math.min(Math.max(ratio, 0), 1) * plotWidth
}

function yScale(points: number): number {
	const ratio = points / maxY.value
	return MARGIN.top + (1 - Math.min(Math.max(ratio, 0), 1)) * plotHeight
}

const actualLinePoints = computed(() =>
	props.burndown.series
		.map(p => `${xScale(p.date)},${yScale(p.remaining)}`)
		.join(' '),
)

const hoverIndex = ref<number | null>(null)

function onMouseMove(e: MouseEvent) {
	const target = e.currentTarget as SVGSVGElement
	const rect = target.getBoundingClientRect()
	const svgX = ((e.clientX - rect.left) / rect.width) * WIDTH

	let closest = 0
	let closestDistance = Infinity
	props.burndown.series.forEach((p, i) => {
		const distance = Math.abs(xScale(p.date) - svgX)
		if (distance < closestDistance) {
			closestDistance = distance
			closest = i
		}
	})
	hoverIndex.value = closest
}
</script>

<style lang="scss" scoped>
.burndown-chart {
	&-hint {
		margin-block-end: .5rem;
	}

	&-legend {
		display: flex;
		gap: 1.5rem;
		font-size: .85rem;
		color: var(--grey-500);
		margin-block-end: .5rem;
	}

	&-svg {
		inline-size: 100%;
		block-size: auto;
		max-block-size: 280px;
	}

	&-tooltip {
		font-size: .85rem;
		color: var(--grey-500);
		margin-block-start: .25rem;
	}
}

.legend-item {
	display: inline-flex;
	align-items: center;
	gap: .4rem;
}

.legend-swatch {
	inline-size: 1rem;
	block-size: 2px;
	border-radius: 1px;

	&.is-ideal {
		background: repeating-linear-gradient(90deg, var(--grey-400) 0 4px, transparent 4px 7px);
	}

	&.is-actual {
		background: var(--primary);
	}
}

.gridline {
	stroke: var(--grey-200);
	stroke-width: 1;
}

.axis-label {
	fill: var(--grey-500);
	font-size: .65rem;
}

.ideal-line {
	stroke: var(--grey-400);
	stroke-width: 2;
	stroke-dasharray: 4 3;
	fill: none;
}

.actual-line {
	stroke: var(--primary);
	stroke-width: 2;
	fill: none;
	stroke-linecap: round;
	stroke-linejoin: round;
}

.actual-point {
	fill: var(--primary);
	stroke: var(--white);
	stroke-width: 1;
	transition: r 100ms;
}

.crosshair {
	stroke: var(--grey-300);
	stroke-width: 1;
	pointer-events: none;
}
</style>
