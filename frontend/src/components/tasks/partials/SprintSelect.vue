<template>
	<Multiselect
		v-model="selectedSprint"
		:placeholder="$t('task.detail.noSprint')"
		:loading="loading"
		:search-results="filteredSprints"
		:show-empty="true"
		label="title"
		:disabled="disabled"
		:aria-label="$t('task.attributes.sprint')"
		@search="findSprints"
	/>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'

import type {ISprint} from '@/modelTypes/ISprint'
import {useSprintService} from '@/services/sprint'
import {useSprintStore} from '@/stores/sprints'
import Multiselect from '@/components/input/Multiselect.vue'
import {error} from '@/message'

const props = withDefaults(defineProps<{
	projectId: number
	disabled?: boolean
}>(), {
	disabled: false,
})

const sprintId = defineModel<number>({
	default: 0,
})

const sprints = ref<ISprint[]>([])
const loading = ref(false)
const sprintService = useSprintService()
const sprintStore = useSprintStore()

watch(() => props.projectId, async projectId => {
	if (!projectId) {
		sprints.value = []
		return
	}

	loading.value = true
	try {
		sprints.value = await sprintService.getAll(projectId)
		sprintStore.setSprints(sprints.value)
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}, {immediate: true})

const selectedSprint = computed<ISprint | null>({
	get: () => sprints.value.find(s => s.id === sprintId.value) ?? null,
	set: value => {
		sprintId.value = value?.id ?? 0
	},
})

const query = ref('')
const filteredSprints = computed(() => {
	if (query.value === '') {
		return sprints.value
	}

	const q = query.value.toLowerCase()
	return sprints.value.filter(s => s.title.toLowerCase().includes(q))
})

function findSprints(newQuery: string) {
	query.value = newQuery
}
</script>
