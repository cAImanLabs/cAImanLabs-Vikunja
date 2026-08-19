<template>
	<span
		:class="`kind-${kind}`"
		class="task-kind-label"
	>
		<span class="icon">
			<Icon :icon="KIND_ICONS[kind]" />
		</span>
		<span>{{ $t(`task.kind.${kind}`) }}</span>
	</span>
</template>

<script setup lang="ts">
import {TASK_KINDS, type TaskKind} from '@/modelTypes/ITaskKind'

defineProps<{
	kind: TaskKind,
}>()

const KIND_ICONS = {
	[TASK_KINDS.TASK]: 'tasks',
	[TASK_KINDS.EPIC]: 'bolt',
	[TASK_KINDS.STORY]: 'bookmark',
	[TASK_KINDS.BUG]: 'bug',
	[TASK_KINDS.SUBTASK]: 'arrow-turn-down',
	[TASK_KINDS.FEATURE]: 'layer-group',
	[TASK_KINDS.INITIATIVE]: 'bullseye',
} as const satisfies Record<TaskKind, string>
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

	&.kind-task {
		color: var(--grey-500);
	}

	&.kind-epic {
		color: var(--primary);
	}

	&.kind-story {
		color: var(--success);
	}

	&.kind-bug {
		color: var(--danger-text);
	}

	&.kind-subtask {
		color: var(--grey-500);
	}

	&.kind-feature {
		color: var(--info);
	}

	&.kind-initiative {
		color: var(--link-visited);
	}
}
</style>
