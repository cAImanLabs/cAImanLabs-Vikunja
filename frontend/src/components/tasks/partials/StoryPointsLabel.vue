<template>
	<span
		v-if="points > 0"
		v-tooltip="$t('task.attributes.storyPoints')"
		class="story-points-label"
	>
		<span class="icon">
			<Icon icon="bolt" />
		</span>
		<span>{{ points }}</span>
	</span>
</template>

<script setup lang="ts">
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
