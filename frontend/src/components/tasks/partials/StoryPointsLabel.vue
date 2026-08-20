<template>
	<span
		v-if="points > 0"
		v-tooltip="$t('task.attributes.storyPoints')"
		class="story-points-label"
		:style="{color: STORY_POINTS_COLORS[points]}"
	>
		<span class="icon">
			<Icon icon="bolt" />
		</span>
		<span>{{ points }}</span>
	</span>
</template>

<script setup lang="ts">
import {STORY_POINTS_COLORS} from '@/helpers/storyPointsMeta'

defineProps<{
	points: number,
}>()
</script>

<style lang="scss" scoped>
.story-points-label {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	color: var(--grey-400);
	font-size: .9rem;
	white-space: nowrap;

	.icon {
		inline-size: auto !important;
	}
}
</style>
