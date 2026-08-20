<template>
	<div
		class="loader-container"
		:class="{'is-loading': loading}"
	>
		<div class="content sprint-picker-header">
			<div>
				<h1>{{ $t('project.sprint.title') }}</h1>
				<p v-if="projects.length > 0">
					{{ $t('project.sprint.pickerDescription') }}
				</p>
				<p
					v-else-if="!loading"
					class="has-text-centered has-text-grey is-italic"
				>
					{{ $t('project.sprint.pickerEmpty') }}
				</p>
			</div>
			<XButton
				v-if="projects.length > 0"
				icon="plus"
				@click="toggleCreateForm"
			>
				{{ $t('project.sprint.create') }}
			</XButton>
		</div>

		<Card
			v-if="showCreateForm"
			:title="$t('project.sprint.create')"
			:show-close="true"
			class="mbe-4"
			@close="toggleCreateForm"
		>
			<div class="field">
				<label class="label">{{ $t('project.sprint.project') }}</label>
				<div class="select">
					<select
						v-model.number="newSprintProjectId"
						:aria-label="$t('project.sprint.project')"
					>
						<option :value="0">
							{{ $t('project.sprint.selectProject') }}
						</option>
						<option
							v-for="p in projects"
							:key="p.id"
							:value="p.id"
						>
							{{ getProjectTitle(p) }}
						</option>
					</select>
				</div>
			</div>

			<SprintForm
				v-if="newSprintProjectId > 0"
				v-model="newSprint"
				:loading="saving"
				@cancel="toggleCreateForm"
				@submit="createSprint"
			>
				<template #extra-fields>
					<div class="field">
						<label class="label">{{ $t('project.sprint.assignTasks') }}</label>
						<Multiselect
							v-model="selectedTasks"
							:placeholder="$t('project.sprint.assignTasksPlaceholder')"
							:loading="taskService.loading"
							:multiple="true"
							:search-results="foundTasks"
							:show-empty="true"
							label="title"
							@search="findTasks"
						/>
					</div>
				</template>
			</SprintForm>
		</Card>

		<Card
			v-if="projects.length > 0"
			:title="$t('project.sprint.allSprints')"
			class="mbe-4"
		>
			<p
				v-if="loadingSprints"
				class="has-text-grey is-size-7"
			>
				{{ $t('project.sprint.allSprintsLoading') }}
			</p>
			<Nothing v-else-if="allSprints.length === 0">
				{{ $t('project.sprint.allSprintsEmpty') }}
			</Nothing>
			<div
				v-else
				class="all-sprints-list"
			>
				<div
					v-for="row in allSprints"
					:key="row.sprint.id"
					class="all-sprints-row"
					:class="[`is-status-${row.sprint.status}`, {'is-loading-card': loadingProjectId === row.project.id}]"
					:style="row.sprint.hexColor ? {borderInlineStartColor: `#${row.sprint.hexColor}`} : undefined"
				>
					<BaseButton
						class="all-sprints-row-content"
						@click="openProjectSprints(row.project)"
					>
						<span class="all-sprints-row-title">{{ row.sprint.title }}</span>
						<span class="all-sprints-row-project">{{ getProjectTitle(row.project) }}</span>
						<span class="all-sprints-row-status">{{ $t(`project.sprint.status_${row.sprint.status}`) }}</span>
					</BaseButton>
					<BaseButton
						v-if="(row.project.maxPermission ?? 0) > Permissions.READ"
						class="has-text-danger"
						:aria-label="$t('project.sprint.delete')"
						@click.stop="confirmDelete(row)"
					>
						<Icon icon="trash-alt" />
					</BaseButton>
				</div>
			</div>
		</Card>

		<div class="sprint-project-list">
			<BaseButton
				v-for="project in projects"
				:key="project.id"
				class="sprint-project-card"
				:class="{'is-loading-card': loadingProjectId === project.id}"
				@click="openProjectSprints(project)"
			>
				<div class="sprint-project-card-title">
					<Icon
						icon="chart-line"
						class="mie-2"
					/>
					{{ getProjectTitle(project) }}
				</div>
				<p class="sprint-project-card-hint">
					{{ $t('project.sprint.pickerCardHint') }}
				</p>
				<span class="sprint-project-card-action">
					{{ $t('project.sprint.viewSprintsAndCharts') }}
					<Icon icon="angle-right" />
				</span>
			</BaseButton>
		</div>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteSprint"
		>
			<template #header>
				<span>{{ $t('project.sprint.delete') }}</span>
			</template>
			<template #text>
				<p>{{ $t('project.sprint.deleteText') }}</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, shallowReactive, watch} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import Nothing from '@/components/misc/Nothing.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import SprintForm from '@/components/project/partials/SprintForm.vue'

import {useProjectStore} from '@/stores/projects'
import {useSprintStore} from '@/stores/sprints'
import {useTaskStore} from '@/stores/tasks'
import {useSprintService} from '@/services/sprint'
import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import TaskService from '@/services/task'
import type {IProject} from '@/modelTypes/IProject'
import type {ISprint, ISprintFormData} from '@/modelTypes/ISprint'
import type {ITask} from '@/modelTypes/ITask'
import {SPRINT_STATUSES} from '@/modelTypes/ISprint'

import {PERMISSIONS as Permissions} from '@/constants/permissions'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import {useTitle} from '@/composables/useTitle'
import {error, success} from '@/message'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('project.sprint.title'))

const projectStore = useProjectStore()
projectStore.loadAllProjects()

const loading = computed(() => projectStore.isLoading)
const projects = computed(() => projectStore.projectsArray.filter(p => p.id > 0 && !p.isArchived))

const router = useRouter()
const loadingProjectId = ref(0)

async function openProjectSprints(project: IProject) {
	loadingProjectId.value = project.id
	try {
		const projectService = new ProjectService()
		const loaded = await projectService.get(new ProjectModel({id: project.id}))
		const sprintView = loaded.views.find(v => v.viewKind === 'sprint')
		if (sprintView) {
			router.push({name: 'project.view', params: {projectId: project.id, viewId: sprintView.id}})
		} else {
			// Very old data the backfill migration hasn't reached (or the user deleted their
			// sprint view) — land on the project itself rather than a dead link.
			router.push({name: 'project.index', params: {projectId: project.id}})
		}
	} catch (e) {
		error(e)
	} finally {
		loadingProjectId.value = 0
	}
}

const sprintService = useSprintService()
const sprintStore = useSprintStore()
const loadingSprints = ref(false)
const allSprints = ref<{sprint: ISprint, project: IProject}[]>([])

const STATUS_ORDER = {active: 0, planning: 1, completed: 2}

async function loadAllSprints(forProjects: readonly IProject[]) {
	if (forProjects.length === 0) {
		allSprints.value = []
		return
	}

	loadingSprints.value = true
	try {
		// projects.value is readonly-wrapped store state; cast to plain IProject here
		// (same pattern the existing openProjectSprints(project) call already relies on)
		// so it flows into allSprints and every consumer without repeating the cast.
		const perProject = await Promise.all(forProjects.map(async project => {
			const plainProject = project as IProject
			const sprints = await sprintService.getAll(plainProject.id)
			sprintStore.setSprints(sprints)
			return sprints.map(sprint => ({sprint, project: plainProject}))
		}))
		allSprints.value = perProject.flat().sort((a, b) => {
			const statusDiff = STATUS_ORDER[a.sprint.status] - STATUS_ORDER[b.sprint.status]
			if (statusDiff !== 0) {
				return statusDiff
			}
			return (b.sprint.startDate?.getTime() ?? 0) - (a.sprint.startDate?.getTime() ?? 0)
		})
	} catch (e) {
		error(e)
	} finally {
		loadingSprints.value = false
	}
}

watch(projects, newProjects => loadAllSprints(newProjects as IProject[]), {immediate: true})

const showDeleteModal = ref(false)
const rowToDelete = ref<{sprint: ISprint, project: IProject} | null>(null)

function confirmDelete(row: {sprint: ISprint, project: IProject}) {
	rowToDelete.value = row
	showDeleteModal.value = true
}

async function deleteSprint() {
	if (!rowToDelete.value) {
		return
	}

	const {sprint, project} = rowToDelete.value
	try {
		await sprintService.remove(project.id, sprint.id)
		allSprints.value = allSprints.value.filter(row => row.sprint.id !== sprint.id)
		showDeleteModal.value = false
		success({message: t('project.sprint.deleteSuccess')})
	} catch (e) {
		error(e)
	}
}

function emptySprint(): ISprintFormData {
	return {
		title: '',
		goal: '',
		startDate: null,
		endDate: null,
		status: SPRINT_STATUSES.PLANNING,
		hexColor: '',
	}
}

const showCreateForm = ref(false)
const newSprintProjectId = ref(0)
const newSprint = ref<ISprintFormData>(emptySprint())
const saving = ref(false)

const taskStore = useTaskStore()
const taskService = shallowReactive(new TaskService())
const selectedTasks = ref<ITask[]>([])
const foundTasks = ref<ITask[]>([])

function toggleCreateForm() {
	showCreateForm.value = !showCreateForm.value
	if (showCreateForm.value) {
		newSprintProjectId.value = 0
		newSprint.value = emptySprint()
		selectedTasks.value = []
		foundTasks.value = []
	}
}

watch(newSprintProjectId, () => {
	selectedTasks.value = []
	foundTasks.value = []
})

async function findTasks(query: string) {
	if (!query || !newSprintProjectId.value) {
		foundTasks.value = []
		return
	}

	const result = await taskService.getAll({}, {s: query, sort_by: 'done'}) as ITask[]
	const alreadySelected = new Set(selectedTasks.value.map(t => t.id))
	foundTasks.value = result.filter(t => t.projectId === newSprintProjectId.value && !alreadySelected.has(t.id))
}

async function createSprint() {
	if (!newSprint.value.title || !newSprintProjectId.value || saving.value) {
		return
	}

	saving.value = true
	try {
		const created = await sprintService.create(newSprintProjectId.value, newSprint.value)
		sprintStore.setSprints([created])

		await Promise.all(selectedTasks.value.map(task =>
			taskStore.update({...task, sprintId: created.id}),
		))

		showCreateForm.value = false
		success({message: t('project.sprint.createSuccess')})
		await loadAllSprints(projects.value as IProject[])
	} catch (e) {
		error(e)
	} finally {
		saving.value = false
	}
}
</script>

<style lang="scss" scoped>
.sprint-picker-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	gap: 1rem;
	flex-wrap: wrap;
}

.all-sprints-list {
	display: flex;
	flex-direction: column;
	gap: .5rem;
}

.all-sprints-row {
	display: flex;
	align-items: center;
	gap: .5rem;
	border-radius: $radius;
	border-inline-start: .25rem solid var(--grey-300);
	transition: background-color $transition;

	&:hover,
	&:focus-within {
		background-color: var(--grey-100);
	}

	&.is-loading-card {
		opacity: .6;
	}

	&.is-status-active {
		border-inline-start-color: var(--primary);
	}

	&.is-status-completed {
		border-inline-start-color: var(--success);
	}

	&-content {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex: 1;
		min-inline-size: 0;
		text-align: start;
		padding: .75rem 1rem;
	}

	&-title {
		font-weight: bold;
		color: var(--text);
	}

	&-project {
		color: var(--grey-500);
		font-size: .9rem;
		flex: 1;
	}

	&-status {
		color: var(--grey-500);
		font-size: .85rem;
		text-transform: capitalize;
	}
}

.sprint-project-list {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	gap: 1rem;
}

.sprint-project-card {
	display: block;
	inline-size: 100%;
	text-align: start;
	background: var(--white);
	padding: 1.5rem;
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	transition: box-shadow $transition;

	&:hover,
	&:focus {
		box-shadow: var(--shadow-md);
	}

	&:active,
	&.is-loading-card {
		box-shadow: var(--shadow-xs) !important;
	}

	&-title {
		font-family: $vikunja-font;
		font-weight: 400;
		font-size: 1.25rem;
		color: var(--text);
	}

	&-hint {
		color: var(--grey-500);
		font-size: .9rem;
		margin-block: .5rem 1rem;
	}

	&-action {
		display: inline-flex;
		align-items: center;
		gap: .5rem;
		color: var(--primary);
		font-weight: bold;
	}
}
</style>
