<template>
	<IconSelect
		v-model="storyPoints"
		:options="options"
		:disabled="disabled"
		:aria-label="$t('task.attributes.storyPoints')"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import IconSelect, {type IconSelectOption} from '@/components/input/IconSelect.vue'
import {STORY_POINTS_VALUES, STORY_POINTS_COLORS} from '@/helpers/storyPointsMeta'

withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const storyPoints = defineModel<number>({
	default: 0,
})

const {t} = useI18n({useScope: 'global'})

const options = computed<IconSelectOption[]>(() => STORY_POINTS_VALUES.map(points => ({
	value: points,
	label: t('task.attributes.storyPointsCount', points),
	icon: 'bolt',
	color: STORY_POINTS_COLORS[points],
})))
</script>
