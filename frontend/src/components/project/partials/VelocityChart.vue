<template>
	<div class="velocity-chart">
		<Nothing v-if="velocity.length === 0">
			{{ $t('project.sprint.velocityEmpty') }}
		</Nothing>
		<template v-else>
			<div class="velocity-chart-legend">
				<span class="legend-item">
					<span class="legend-swatch is-total" />
					{{ $t('project.sprint.velocityTotal') }}
				</span>
				<span class="legend-item">
					<span class="legend-swatch is-completed" />
					{{ $t('project.sprint.velocityCompleted') }}
				</span>
			</div>

			<svg
				:viewBox="`0 0 ${WIDTH} ${HEIGHT}`"
				class="velocity-chart-svg"
				role="img"
				:aria-label="$t('project.sprint.velocityAriaLabel')"
			>
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

				<g
					v-for="(v, i) in velocity"
					:key="v.sprintId"
					@mouseenter="hoverIndex = i"
					@mouseleave="hoverIndex = null"
				>
					<rect
						:x="barX(i)"
						:y="yScale(v.totalPoints)"
						:width="barWidth"
						:height="Math.max(plotHeight - (yScale(v.totalPoints) - MARGIN.top), 0)"
						class="bar-total"
						:class="{'is-hovered': hoverIndex === i}"
					/>
					<rect
						:x="barX(i)"
						:y="yScale(v.completedPoints)"
						:width="barWidth"
						:height="Math.max(plotHeight - (yScale(v.completedPoints) - MARGIN.top), 0)"
						class="bar-completed"
						:class="{'is-hovered': hoverIndex === i}"
					/>
					<text
						:x="barX(i) + barWidth / 2"
						:y="HEIGHT - 8"
						class="axis-label"
						text-anchor="middle"
					>{{ truncateTitle(v.title) }}</text>
				</g>
			</svg>

			<div
				v-if="hoverIndex !== null"
				class="velocity-chart-tooltip"
			>
				<strong>{{ velocity[hoverIndex].title }}</strong>
				{{ $t('project.sprint.velocityTooltip', {completed: velocity[hoverIndex].completedPoints, total: velocity[hoverIndex].totalPoints}) }}
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'

import Nothing from '@/components/misc/Nothing.vue'

import type {ISprintVelocityPoint} from '@/modelTypes/ISprint'

const props = defineProps<{
	velocity: ISprintVelocityPoint[],
}>()

const WIDTH = 640
const HEIGHT = 280
const MARGIN = {top: 16, right: 16, bottom: 32, left: 40}

const plotWidth = WIDTH - MARGIN.left - MARGIN.right
const plotHeight = HEIGHT - MARGIN.top - MARGIN.bottom

const maxY = computed(() => Math.max(...props.velocity.map(v => v.totalPoints), 1))

const yTicks = computed(() => {
	const max = maxY.value
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

function yScale(points: number): number {
	const ratio = points / maxY.value
	return MARGIN.top + (1 - Math.min(Math.max(ratio, 0), 1)) * plotHeight
}

const barGap = 16
const barWidth = computed(() => {
	const count = props.velocity.length || 1
	return Math.max((plotWidth - barGap * (count - 1)) / count, 8)
})

function barX(index: number): number {
	return MARGIN.left + index * (barWidth.value + barGap)
}

function truncateTitle(title: string): string {
	return title.length > 12 ? title.slice(0, 11) + '…' : title
}

const hoverIndex = ref<number | null>(null)
</script>

<style lang="scss" scoped>
.velocity-chart {
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
	inline-size: .75rem;
	block-size: .75rem;
	border-radius: 2px;

	&.is-total {
		background: var(--grey-300);
	}

	&.is-completed {
		background: var(--success);
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

.bar-total {
	fill: var(--grey-300);
	transition: opacity 100ms;

	&.is-hovered {
		opacity: .8;
	}
}

.bar-completed {
	fill: var(--success);
	transition: opacity 100ms;

	&.is-hovered {
		opacity: .85;
	}
}
</style>
