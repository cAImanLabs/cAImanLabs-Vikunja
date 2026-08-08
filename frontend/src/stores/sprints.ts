import {computed, readonly, ref} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'

import {useSprintService} from '@/services/sprint'
import type {ISprint} from '@/modelTypes/ISprint'
import type {IProject} from '@/modelTypes/IProject'

// A small cross-view cache of sprints by id, keyed independently of any
// single project view, so task list rows (List, ShowTasks) can resolve a
// sprint's title for a badge without each row fetching it individually.
export const useSprintStore = defineStore('sprint', () => {
	const sprints = ref<{ [id: ISprint['id']]: ISprint }>({})
	const loadedProjectIds = ref(new Set<IProject['id']>())

	const getSprintById = computed(() => {
		return (sprintId: ISprint['id']) => sprints.value[sprintId]
	})

	function setSprints(newSprints: ISprint[]) {
		newSprints.forEach(s => {
			sprints.value[s.id] = s
		})
	}

	async function ensureProjectLoaded(projectId: IProject['id']) {
		if (!projectId || loadedProjectIds.value.has(projectId)) {
			return
		}
		loadedProjectIds.value.add(projectId)

		const sprintService = useSprintService()
		try {
			setSprints(await sprintService.getAll(projectId))
		} catch {
			// Allow a later call to retry (e.g. after a transient network error).
			loadedProjectIds.value.delete(projectId)
		}
	}

	return {
		sprints: readonly(sprints),
		getSprintById,
		setSprints,
		ensureProjectLoaded,
	}
})

if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useSprintStore, import.meta.hot))
}
