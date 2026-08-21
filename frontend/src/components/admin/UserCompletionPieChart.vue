<template>
	<div class="completion-pie-chart">
		<Nothing v-if="segments.length === 0">
			{{ $t('admin.overview.completionChartsEmpty') }}
		</Nothing>
		<template v-else>
			<svg
				:viewBox="`0 0 ${SIZE} ${SIZE}`"
				class="completion-pie-chart-svg"
				role="img"
				:aria-label="$t('admin.overview.completionShareAriaLabel')"
			>
				<g :transform="`rotate(-90 ${CENTER} ${CENTER})`">
					<circle
						v-for="seg in segments"
						:key="seg.key"
						:cx="CENTER"
						:cy="CENTER"
						:r="RADIUS"
						fill="none"
						:stroke="seg.color"
						:stroke-width="STROKE_WIDTH"
						:stroke-dasharray="`${seg.length} ${CIRCUMFERENCE - seg.length}`"
						:stroke-dashoffset="-seg.offset"
						class="pie-segment"
						:class="{'is-hovered': hoverKey === seg.key, 'is-dimmed': hoverKey !== null && hoverKey !== seg.key}"
						@mouseenter="hoverKey = seg.key"
						@mouseleave="hoverKey = null"
					/>
				</g>
				<text
					:x="CENTER"
					:y="CENTER"
					text-anchor="middle"
					dominant-baseline="middle"
					class="pie-center-label"
				>
					{{ total }}
				</text>
			</svg>

			<ul class="completion-pie-chart-legend">
				<li
					v-for="seg in segments"
					:key="seg.key"
					class="legend-item"
					:class="{'is-hovered': hoverKey === seg.key, 'is-dimmed': hoverKey !== null && hoverKey !== seg.key}"
					@mouseenter="hoverKey = seg.key"
					@mouseleave="hoverKey = null"
				>
					<span
						class="legend-swatch"
						:style="{background: seg.color}"
					/>
					<span class="legend-label">{{ seg.label }}</span>
					<span class="legend-value">{{ seg.count }} ({{ seg.percentLabel }})</span>
				</li>
			</ul>
		</template>
	</div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import type {IUserTaskCompletionStat} from '@/modelTypes/IUserTaskCompletionStat'
import {foldUserSeries} from '@/helpers/foldUserSeries'
import Nothing from '@/components/misc/Nothing.vue'

const props = defineProps<{
	stats: IUserTaskCompletionStat[],
}>()

const {t} = useI18n({useScope: 'global'})

const SIZE = 200
const CENTER = SIZE / 2
const STROKE_WIDTH = 32
const RADIUS = CENTER - STROKE_WIDTH / 2
const CIRCUMFERENCE = 2 * Math.PI * RADIUS

const total = computed(() => props.stats.reduce((sum, s) => sum + s.completed, 0))

const segments = computed(() => {
	if (total.value === 0) return []

	const entries = foldUserSeries(props.stats, s => s.completed, t('admin.overview.completionOtherLabel'))

	let offset = 0
	return entries.map(entry => {
		const length = (entry.value / total.value) * CIRCUMFERENCE
		const seg = {
			...entry,
			count: entry.value,
			length,
			offset,
			percentLabel: `${Math.round((entry.value / total.value) * 100)}%`,
		}
		offset += length
		return seg
	})
})

const hoverKey = ref<string | null>(null)
</script>

<style lang="scss" scoped>
.completion-pie-chart {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 1.5rem;
}

.completion-pie-chart-svg {
	inline-size: 160px;
	block-size: 160px;
	flex-shrink: 0;
}

.pie-segment {
	transition: opacity 100ms;

	&.is-hovered {
		opacity: 0.85;
	}

	&.is-dimmed {
		opacity: 0.35;
	}
}

.pie-center-label {
	fill: var(--grey-700);
	font-size: 1.5rem;
	font-weight: 600;
	transform: rotate(90deg);
	transform-origin: center;
}

.completion-pie-chart-legend {
	list-style: none;
	margin: 0;
	padding: 0;
	display: flex;
	flex-direction: column;
	gap: 0.4rem;
	font-size: 0.85rem;
	flex: 1 1 auto;
	min-inline-size: 12rem;
}

.legend-item {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	transition: opacity 100ms;

	&.is-dimmed {
		opacity: 0.5;
	}
}

.legend-swatch {
	inline-size: 0.75rem;
	block-size: 0.75rem;
	border-radius: 2px;
	flex-shrink: 0;
}

.legend-label {
	flex: 1 1 auto;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.legend-value {
	color: var(--grey-500);
	flex-shrink: 0;
}
</style>
