<template>
	<IconSelect
		v-model="kind"
		:options="options"
		:disabled="disabled"
		:aria-label="$t('task.attributes.kind')"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import IconSelect, {type IconSelectOption} from '@/components/input/IconSelect.vue'
import {TASK_KINDS, TASK_KIND_ORDER} from '@/modelTypes/ITaskKind'
import {TASK_KIND_ICONS, TASK_KIND_COLORS, TASK_KIND_I18N_KEYS} from '@/helpers/taskKindMeta'

withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const kind = defineModel<string>({
	default: TASK_KINDS.TASK,
})

const {t} = useI18n({useScope: 'global'})

const options = computed<IconSelectOption[]>(() => TASK_KIND_ORDER.map(k => ({
	value: k,
	label: t(TASK_KIND_I18N_KEYS[k]),
	icon: TASK_KIND_ICONS[k],
	color: TASK_KIND_COLORS[k],
})))
</script>
