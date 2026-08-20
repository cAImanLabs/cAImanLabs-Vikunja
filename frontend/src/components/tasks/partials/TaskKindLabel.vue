<template>
	<span
		class="task-kind-label"
		:style="{color: TASK_KIND_COLORS[kind]}"
	>
		<span class="icon">
			<Icon :icon="TASK_KIND_ICONS[kind]" />
		</span>
		<span>{{ $t(TASK_KIND_I18N_KEYS[kind]) }}</span>
	</span>
</template>

<script setup lang="ts">
import type {TaskKind} from '@/modelTypes/ITaskKind'
import {TASK_KIND_ICONS, TASK_KIND_COLORS, TASK_KIND_I18N_KEYS} from '@/helpers/taskKindMeta'

defineProps<{
	kind: TaskKind,
}>()
</script>

<style lang="scss" scoped>
.task-kind-label {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	font-size: .9rem;
	white-space: nowrap;

	.icon {
		inline-size: auto !important;
	}
}
</style>
