<template>
	<span
		v-if="!done && (showAll || priority >= minimumPriority)"
		class="priority-label"
		:style="{color: PRIORITY_COLORS[priority as Priority]}"
	>
		<span class="icon">
			<Icon :icon="PRIORITY_ICONS[priority as Priority]" />
		</span>
		<span>{{ $t(PRIORITY_I18N_KEYS[priority as Priority]) }}</span>
	</span>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {PRIORITIES as priorities, type Priority} from '@/constants/priorities'
import {PRIORITY_ICONS, PRIORITY_COLORS, PRIORITY_I18N_KEYS} from '@/helpers/priorityMeta'
import {useAuthStore} from '@/stores/auth'

withDefaults(defineProps<{
	priority: number,
	showAll?: boolean,
	done?: boolean
}>(), {
	showAll: false,
	done: false,
})

const authStore = useAuthStore()

const minimumPriority = computed(() => {
	return authStore.settings.frontendSettings.minimumPriority || priorities.MEDIUM
})
</script>

<style lang="scss" scoped>
.priority-label {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	font-size: .9rem;
	white-space: nowrap;

	.icon {
		vertical-align: top;
		inline-size: auto !important;
	}
}
</style>
