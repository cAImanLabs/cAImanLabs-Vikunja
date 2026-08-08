<template>
	<ProjectWrapper
		class="project-sprint"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id="viewId"
	>
		<template #header>
			<XButton
				v-if="canWrite"
				icon="plus"
				@click="toggleCreateForm"
			>
				{{ $t('project.sprint.create') }}
			</XButton>
		</template>

		<div
			class="loader-container"
			:class="{'is-loading': loading}"
		>
			<div
				v-if="(velocity.length > 0 || selectedBurndownSprintId > 0) && !showCreateForm"
				class="sprint-charts mbe-4"
			>
				<BaseButton
					class="sprint-charts-toggle"
					@click="chartsVisible = !chartsVisible"
				>
					<Icon :icon="chartsVisible ? 'chevron-down' : 'chevron-right'" />
					{{ chartsVisible ? $t('project.sprint.hideCharts') : $t('project.sprint.showCharts') }}
				</BaseButton>

				<div
					v-if="chartsVisible"
					class="sprint-charts-grid"
				>
					<Card
						v-if="velocity.length > 0"
						:title="$t('project.sprint.velocity')"
					>
						<VelocityChart :velocity="velocity" />
					</Card>

					<Card
						v-if="selectedBurndownSprintId > 0"
						:title="$t('project.sprint.burndownOf', {title: selectedBurndownSprintTitle})"
						:show-close="true"
						@close="clearBurndown"
					>
						<p
							v-if="burndownMissingDates"
							class="has-text-grey is-size-7"
						>
							{{ $t('project.sprint.burndownMissingDates') }}
						</p>
						<BurndownChart
							v-else-if="burndownData"
							:burndown="burndownData"
						/>
					</Card>
				</div>
			</div>

			<Card
				v-if="showCreateForm"
				:title="$t('project.sprint.create')"
				:show-close="true"
				class="mbe-4"
				@close="showCreateForm = false"
			>
				<SprintForm
					v-model="newSprint"
					:loading="saving"
					@cancel="showCreateForm = false"
					@submit="createSprint"
				/>
			</Card>

			<Nothing v-if="!loading && sprints.length === 0 && !showCreateForm">
				{{ $t('project.sprint.empty') }}
				<ButtonLink
					v-if="canWrite"
					@click="toggleCreateForm"
				>
					{{ $t('project.sprint.create') }}
				</ButtonLink>
			</Nothing>

			<div class="sprint-grid">
				<div
					v-for="sprint in sprints"
					:key="sprint.id"
					class="sprint-card"
					:class="`is-status-${sprint.status}`"
				>
					<template v-if="editingId === sprint.id">
						<SprintForm
							v-model="editSprint"
							:loading="saving"
							@cancel="editingId = 0"
							@submit="saveSprint"
						/>
					</template>
					<template v-else>
						<div class="sprint-card-header">
							<h3 class="title is-5 mb-0">
								{{ sprint.title }}
							</h3>
							<div class="sprint-card-actions">
								<BaseButton
									:aria-label="selectedBurndownSprintId === sprint.id ? $t('project.sprint.hideBurndown') : $t('project.sprint.viewBurndown')"
									:class="{'is-active': selectedBurndownSprintId === sprint.id}"
									@click="toggleBurndown(sprint)"
								>
									<Icon icon="chart-line" />
								</BaseButton>
								<BaseButton
									v-if="canWrite"
									:aria-label="$t('project.sprint.edit')"
									@click="startEdit(sprint)"
								>
									<Icon icon="pen" />
								</BaseButton>
								<BaseButton
									v-if="canWrite"
									class="has-text-danger"
									:aria-label="$t('project.sprint.delete')"
									@click="confirmDelete(sprint)"
								>
									<Icon icon="trash-alt" />
								</BaseButton>
							</div>
						</div>

						<div class="select sprint-status">
							<select
								:value="sprint.status"
								:disabled="!canWrite"
								:aria-label="$t('project.sprint.status')"
								@change="updateStatus(sprint, ($event.target as HTMLSelectElement).value as SprintStatus)"
							>
								<option
									v-for="s in SPRINT_STATUSES"
									:key="s"
									:value="s"
								>
									{{ $t(`project.sprint.status_${s}`) }}
								</option>
							</select>
						</div>

						<p
							v-if="sprint.goal"
							class="sprint-goal"
						>
							{{ sprint.goal }}
						</p>

						<p
							v-if="sprint.startDate || sprint.endDate"
							class="sprint-dates has-text-grey"
						>
							<Icon
								icon="calendar"
								class="mie-1"
							/>
							{{ formatDateRange(sprint) }}
						</p>
					</template>
				</div>
			</div>
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
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import SprintForm from '@/components/project/partials/SprintForm.vue'
import VelocityChart from '@/components/project/partials/VelocityChart.vue'
import BurndownChart from '@/components/project/partials/BurndownChart.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import ButtonLink from '@/components/misc/ButtonLink.vue'
import Nothing from '@/components/misc/Nothing.vue'

import {PERMISSIONS as Permissions} from '@/constants/permissions'
import {formatDisplayDate} from '@/helpers/time/formatDate'
import {success, error} from '@/message'

import {useProjectStore} from '@/stores/projects'
import {useSprintStore} from '@/stores/sprints'
import {useSprintService} from '@/services/sprint'
import {SPRINT_STATUSES} from '@/modelTypes/ISprint'
import type {ISprint, ISprintFormData, ISprintBurndown, ISprintVelocityPoint, SprintStatus} from '@/modelTypes/ISprint'
import type {IProjectView} from '@/modelTypes/IProjectView'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: number,
	viewId: IProjectView['id'],
}>()

const {t} = useI18n({useScope: 'global'})

const projectStore = useProjectStore()
const sprintStore = useSprintStore()
const sprintService = useSprintService()

const project = computed(() => projectStore.projects[props.projectId])
const canWrite = computed(() => project.value?.maxPermission > Permissions.READ && project.value?.id > 0)

const sprints = ref<ISprint[]>([])
const loading = ref(false)
const saving = ref(false)

const velocity = ref<ISprintVelocityPoint[]>([])

const chartsVisible = ref(true)
const selectedBurndownSprintId = ref(0)
const selectedBurndownSprintTitle = computed(() => sprints.value.find(s => s.id === selectedBurndownSprintId.value)?.title ?? '')
const burndownData = ref<ISprintBurndown | null>(null)
const burndownMissingDates = ref(false)

async function loadVelocity() {
	if (!props.projectId) {
		velocity.value = []
		return
	}
	try {
		velocity.value = await sprintService.getVelocity(props.projectId)
	} catch (e) {
		error(e)
	}
}

function clearBurndown() {
	selectedBurndownSprintId.value = 0
	burndownData.value = null
	burndownMissingDates.value = false
}

async function toggleBurndown(sprint: ISprint) {
	if (selectedBurndownSprintId.value === sprint.id) {
		selectedBurndownSprintId.value = 0
		burndownData.value = null
		return
	}

	chartsVisible.value = true
	selectedBurndownSprintId.value = sprint.id
	burndownData.value = null
	burndownMissingDates.value = false

	try {
		burndownData.value = await sprintService.getBurndown(props.projectId, sprint.id)
	} catch (e) {
		if (e?.response?.status === 412) {
			burndownMissingDates.value = true
		} else {
			error(e)
		}
	}
}

function emptySprint(): ISprintFormData {
	return {
		title: '',
		goal: '',
		startDate: null,
		endDate: null,
		status: SPRINT_STATUSES.PLANNING,
	}
}

const showCreateForm = ref(false)
const newSprint = ref<ISprintFormData>(emptySprint())

const editingId = ref(0)
const editSprint = ref<ISprintFormData>(emptySprint())

const showDeleteModal = ref(false)
const sprintToDelete = ref<ISprint | null>(null)

async function loadSprints() {
	if (!props.projectId) {
		sprints.value = []
		return
	}

	loading.value = true
	try {
		sprints.value = await sprintService.getAll(props.projectId)
		sprintStore.setSprints(sprints.value)
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

watch(() => props.projectId, () => {
	loadSprints()
	loadVelocity()
	selectedBurndownSprintId.value = 0
	burndownData.value = null
}, {immediate: true})

function toggleCreateForm() {
	showCreateForm.value = !showCreateForm.value
	if (showCreateForm.value) {
		editingId.value = 0
		newSprint.value = emptySprint()
	}
}

async function createSprint() {
	if (!newSprint.value.title || saving.value) {
		return
	}

	saving.value = true
	try {
		const created = await sprintService.create(props.projectId, newSprint.value)
		sprints.value = [...sprints.value, created]
		sprintStore.setSprints([created])
		showCreateForm.value = false
		success({message: t('project.sprint.createSuccess')})
		loadVelocity()
	} finally {
		saving.value = false
	}
}

function startEdit(sprint: ISprint) {
	showCreateForm.value = false
	editingId.value = sprint.id
	editSprint.value = {
		title: sprint.title,
		goal: sprint.goal,
		startDate: sprint.startDate,
		endDate: sprint.endDate,
		status: sprint.status,
	}
}

async function saveSprint() {
	if (!editSprint.value.title || !editingId.value || saving.value) {
		return
	}

	saving.value = true
	try {
		const updated = await sprintService.update(props.projectId, {
			...editSprint.value,
			id: editingId.value,
		})
		sprints.value = sprints.value.map(s => s.id === updated.id ? updated : s)
		sprintStore.setSprints([updated])
		editingId.value = 0
		success({message: t('project.sprint.updateSuccess')})
	} finally {
		saving.value = false
	}
}

async function updateStatus(sprint: ISprint, status: SprintStatus) {
	const updated = await sprintService.update(props.projectId, {
		id: sprint.id,
		title: sprint.title,
		goal: sprint.goal,
		startDate: sprint.startDate,
		endDate: sprint.endDate,
		status,
	})
	sprints.value = sprints.value.map(s => s.id === updated.id ? updated : s)
	sprintStore.setSprints([updated])
	success({message: t('project.sprint.updateSuccess')})
}

function confirmDelete(sprint: ISprint) {
	sprintToDelete.value = sprint
	showDeleteModal.value = true
}

async function deleteSprint() {
	if (!sprintToDelete.value) {
		return
	}

	await sprintService.remove(props.projectId, sprintToDelete.value.id)
	sprints.value = sprints.value.filter(s => s.id !== sprintToDelete.value?.id)
	if (selectedBurndownSprintId.value === sprintToDelete.value.id) {
		selectedBurndownSprintId.value = 0
		burndownData.value = null
	}
	showDeleteModal.value = false
	success({message: t('project.sprint.deleteSuccess')})
	loadVelocity()
}

function formatDateRange(sprint: ISprint) {
	if (sprint.startDate && sprint.endDate) {
		return `${formatDisplayDate(sprint.startDate)} – ${formatDisplayDate(sprint.endDate)}`
	}
	if (sprint.startDate) {
		return formatDisplayDate(sprint.startDate)
	}
	return formatDisplayDate(sprint.endDate)
}
</script>

<style lang="scss" scoped>
.sprint-charts {
	&-toggle {
		display: inline-flex;
		align-items: center;
		gap: .35rem;
		color: var(--grey-500);
		font-size: .9rem;
		margin-block-end: .5rem;

		&:hover {
			color: var(--text);
		}
	}

	&-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(420px, 1fr));
		gap: 1rem;
		align-items: start;

		:deep(.card) {
			margin-block-end: 0;
		}
	}
}

.sprint-grid {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
	gap: 1rem;
	align-items: start;
}

.sprint-card {
	--sprint-card-padding: 1.5rem;
	background: var(--white);
	padding: var(--sprint-card-padding);
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	border-inline-start: .25rem solid var(--grey-300);
	transition: box-shadow $transition;

	&:hover {
		box-shadow: var(--shadow-md);
	}

	&.is-status-active {
		border-inline-start-color: var(--primary);
	}

	&.is-status-completed {
		border-inline-start-color: var(--success);
	}

	&-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: .5rem;
		margin-block-end: .75rem;
	}

	&-actions {
		display: flex;
		gap: .25rem;
		flex-shrink: 0;

		.icon {
			pointer-events: none;
		}

		.is-active {
			color: var(--primary);
		}
	}
}

.sprint-status {
	margin-block-end: .75rem;
}

.sprint-goal {
	color: var(--text);
	margin-block-end: .75rem;
	white-space: pre-wrap;
	word-break: break-word;
}

.sprint-dates {
	font-size: .9rem;
}
</style>
