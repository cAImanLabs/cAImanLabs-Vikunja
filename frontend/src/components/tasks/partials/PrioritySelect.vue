<template>
	<IconSelect
		v-model="priority"
		:options="options"
		:disabled="disabled"
		:aria-label="$t('task.attributes.priority')"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import IconSelect, {type IconSelectOption} from '@/components/input/IconSelect.vue'
import {PRIORITY_ORDER, PRIORITY_ICONS, PRIORITY_COLORS, PRIORITY_I18N_KEYS} from '@/helpers/priorityMeta'

withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const priority = defineModel<number>({
	required: true,
	default: 0,
})

const {t} = useI18n({useScope: 'global'})

const options = computed<IconSelectOption[]>(() => PRIORITY_ORDER.map(p => ({
	value: p,
	label: t(PRIORITY_I18N_KEYS[p]),
	icon: PRIORITY_ICONS[p],
	color: PRIORITY_COLORS[p],
})))
</script>
