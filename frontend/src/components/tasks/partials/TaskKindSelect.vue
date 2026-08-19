<template>
	<div class="select">
		<select
			v-model="kind"
			:disabled="disabled || undefined"
			:aria-label="$t('task.attributes.kind')"
		>
			<option
				v-for="k in TASK_KIND_ORDER"
				:key="k"
				:value="k"
			>
				{{ $t(`task.kind.${k}`) }}
			</option>
		</select>
	</div>
</template>

<script setup lang="ts">
import {TASK_KINDS, TASK_KIND_ORDER} from '@/modelTypes/ITaskKind'

withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const kind = defineModel<string>({
	default: TASK_KINDS.TASK,
})
</script>
