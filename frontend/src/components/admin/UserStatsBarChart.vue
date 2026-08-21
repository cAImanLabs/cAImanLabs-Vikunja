<template>
	<div class="user-stats-bar-chart">
		<Nothing v-if="entries.length === 0">
			{{ $t('admin.overview.completionChartsEmpty') }}
		</Nothing>
		<ul
			v-else
			class="user-stats-bar-chart-list"
		>
			<li
				v-for="entry in entries"
				:key="entry.key"
				class="bar-row"
			>
				<span
					class="bar-row-label"
					:title="entry.label"
				>{{ entry.label }}</span>
				<span class="bar-row-track">
					<span
						class="bar-row-fill"
						:style="{
							inlineSize: `${maxValue > 0 ? (entry.value / maxValue) * 100 : 0}%`,
							background: entry.color,
						}"
					/>
				</span>
				<span class="bar-row-value">{{ entry.value }}</span>
			</li>
		</ul>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import type {IUserTaskCompletionStat} from '@/modelTypes/IUserTaskCompletionStat'
import {foldUserSeries} from '@/helpers/foldUserSeries'
import Nothing from '@/components/misc/Nothing.vue'

const props = defineProps<{
	stats: IUserTaskCompletionStat[],
	valueKey: 'completed' | 'storyPoints',
}>()

const {t} = useI18n({useScope: 'global'})

const entries = computed(() => foldUserSeries(
	props.stats,
	s => props.valueKey === 'completed' ? s.completed : s.storyPoints,
	t('admin.overview.completionOtherLabel'),
).filter(entry => entry.value > 0))

const maxValue = computed(() => Math.max(...entries.value.map(e => e.value), 1))
</script>

<style lang="scss" scoped>
.user-stats-bar-chart-list {
	list-style: none;
	margin: 0;
	padding: 0;
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.bar-row {
	display: grid;
	grid-template-columns: minmax(4rem, 8rem) 1fr 2.5rem;
	align-items: center;
	gap: 0.6rem;
	font-size: 0.85rem;
}

.bar-row-label {
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.bar-row-track {
	background: var(--grey-100);
	border-radius: 3px;
	block-size: 0.6rem;
	overflow: hidden;
}

.bar-row-fill {
	display: block;
	block-size: 100%;
	border-radius: 3px;
	transition: inline-size 200ms;
}

.bar-row-value {
	text-align: end;
	color: var(--grey-600);
	font-variant-numeric: tabular-nums;
}
</style>
