<template>
	<span
		v-if="title"
		v-tooltip="$t('task.attributes.sprint')"
		class="sprint-label"
	>
		<span class="icon">
			<Icon icon="layer-group" />
		</span>
		<span>{{ title }}</span>
	</span>
</template>

<script setup lang="ts">
defineProps<{
	title?: string,
}>()
</script>

<style lang="scss" scoped>
.sprint-label {
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
