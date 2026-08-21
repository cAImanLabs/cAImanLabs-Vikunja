<template>
	<Card>
		<div class="admin-tasks">
			<div class="admin-tasks__toolbar">
				<FormInput
					v-model="searchTerm"
					type="text"
					:placeholder="$t('admin.searchTasksPlaceholder')"
					@input="onSearch"
				/>
			</div>

			<p v-if="loading">
				{{ $t('misc.loading') }}
			</p>
			<template v-else>
				<div class="admin-tasks__table-scroll">
					<table class="table has-actions is-striped is-hoverable is-fullwidth">
						<thead>
							<tr>
								<th>{{ $t('misc.id') }}</th>
								<th>{{ $t('task.attributes.title') }}</th>
								<th>{{ $t('task.attributes.project') }}</th>
								<th>{{ $t('task.attributes.priority') }}</th>
								<th>{{ $t('task.attributes.dueDate') }}</th>
								<th>{{ $t('task.attributes.done') }}</th>
								<th>{{ $t('admin.tasks.completedByLabel') }}</th>
								<th>{{ $t('task.attributes.created') }}</th>
								<th />
							</tr>
						</thead>
						<tbody>
							<tr
								v-for="task in tasks"
								:key="task.id"
							>
								<td>{{ task.id }}</td>
								<td>
									<RouterLink :to="{name: 'task.detail', params: {id: task.id}, query: {from: 'admin'}}">
										{{ task.title }}
									</RouterLink>
									<TaskKindLabel
										:kind="task.kind"
										class="admin-tasks__kind"
									/>
								</td>
								<td>{{ task.projectTitle }}</td>
								<td>
									<PriorityLabel
										:priority="task.priority"
										:done="task.done"
									/>
								</td>
								<td>
									<TimeDisplay
										v-if="task.dueDate"
										:date="task.dueDate"
									/>
								</td>
								<td>{{ task.done ? $t('task.attributes.done') : '' }}</td>
								<td>{{ task.completedBy?.username ?? '' }}</td>
								<td>
									<TimeDisplay :date="task.created" />
								</td>
								<td class="actions">
									<XButton
										v-if="task.done"
										variant="secondary"
										@click="openReassign(task)"
									>
										{{ $t('admin.tasks.reassignCompletedBy') }}
									</XButton>
								</td>
							</tr>
						</tbody>
					</table>
				</div>
				<PaginationEmit
					v-if="totalPages > 1"
					:total-pages="totalPages"
					:current-page="currentPage"
					@pageChanged="goToPage"
				/>
			</template>

			<Modal
				v-if="reassignTarget"
				variant="hint-modal"
				@close="reassignTarget = null"
			>
				<Card
					class="has-no-shadow"
					:title="$t('admin.tasks.reassignTitle', {title: reassignTarget.title})"
				>
					<FormField :label="$t('admin.tasks.newCompletedByLabel')">
						<Multiselect
							v-model="selectedUser"
							:loading="userSearchLoading"
							:placeholder="$t('admin.searchUsersPlaceholder')"
							:search-results="userResults"
							:show-empty="true"
							label="username"
							@search="searchUsers"
						>
							<template #searchResult="{option}">
								<User
									v-if="typeof option !== 'string'"
									:avatar-size="24"
									:show-username="true"
									:user="option"
								/>
							</template>
						</Multiselect>
					</FormField>

					<template #footer>
						<XButton
							variant="tertiary"
							@click="reassignTarget = null"
						>
							{{ $t('misc.cancel') }}
						</XButton>
						<XButton
							variant="primary"
							:disabled="!selectedUser"
							@click="doReassign()"
						>
							{{ $t('admin.tasks.reassignCompletedBy') }}
						</XButton>
					</template>
				</Card>
			</Modal>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {useDebounceFn} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import type {IAdminTask} from '@/modelTypes/IAdminTask'
import type {IAdminUser} from '@/modelTypes/IAdminUser'
import AdminTaskService from '@/services/admin/taskService'
import AdminUserService from '@/services/admin/userService'
import AdminTaskModel from '@/models/adminTask'
import AdminUserModel from '@/models/adminUser'
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import XButton from '@/components/input/Button.vue'
import FormInput from '@/components/input/FormInput.vue'
import FormField from '@/components/input/FormField.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'
import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import TaskKindLabel from '@/components/tasks/partials/TaskKindLabel.vue'
import {error, success} from '@/message'

const {t} = useI18n({useScope: 'global'})

const adminTaskService = new AdminTaskService()
const adminUserService = new AdminUserService()

const tasks = ref<IAdminTask[]>([])
const loading = ref(false)
const searchTerm = ref('')
const currentPage = ref(1)
const totalPages = ref(1)

const reassignTarget = ref<IAdminTask | null>(null)
const userResults = ref<IAdminUser[]>([])
const userSearchLoading = ref(false)
const selectedUser = ref<IAdminUser | null>(null)

async function load() {
	loading.value = true
	try {
		const params = searchTerm.value ? {s: searchTerm.value} : {}
		tasks.value = await adminTaskService.getAll(new AdminTaskModel(), params, currentPage.value)
		totalPages.value = adminTaskService.totalPages || 1
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

function goToPage(page: number) {
	currentPage.value = page
	load()
}

const onSearch = useDebounceFn(() => {
	// Reset to page 1 so a narrower search doesn't strand the UI on an empty page.
	currentPage.value = 1
	load()
}, 300)

function openReassign(task: IAdminTask) {
	reassignTarget.value = task
	userResults.value = []
	selectedUser.value = null
	// Loads the full (unfiltered) user list as soon as the modal opens, so
	// admins who don't know the exact username get a browsable dropdown
	// instead of a blank box that only responds once you start typing. Not
	// wired to a Multiselect @focus listener: 'focus' doesn't bubble and Vue's
	// attrs fallthrough attaches to the component root, not the inner input,
	// so it never actually fires.
	searchUsers('')
}

async function searchUsers(query = '') {
	userSearchLoading.value = true
	try {
		userResults.value = await adminUserService.getAll(new AdminUserModel(), {s: query})
	} catch (e) {
		error(e)
	} finally {
		userSearchLoading.value = false
	}
}

async function doReassign() {
	if (!reassignTarget.value || !selectedUser.value) return
	const target = reassignTarget.value
	const newCompletedById = selectedUser.value.id
	reassignTarget.value = null
	try {
		// The completed-by endpoint returns a plain task, not the admin list
		// shape, so project_title comes back empty - carry the old value over.
		const updated = await adminTaskService.reassignCompletedBy(target.id, newCompletedById)
		updated.projectTitle = target.projectTitle
		const idx = tasks.value.findIndex(x => x.id === target.id)
		if (idx !== -1) tasks.value[idx] = updated
		success({message: t('admin.tasks.reassignedSuccess')})
	} catch (e) {
		error(e)
	}
}

onMounted(load)
</script>

<style lang="scss" scoped>
// `.table.has-actions` sets overflow: hidden which clips the dropdown menu.
.admin-tasks :deep(.table.has-actions) {
	overflow: visible;
}

// The table is wider than the viewport on small/medium screens (project,
// completed-by and the reassign button all need room) - scroll it
// horizontally instead of letting the action column run off-screen.
.admin-tasks__table-scroll {
	overflow-x: auto;
}

.admin-tasks__toolbar {
	display: flex;
	gap: 0.5rem;
	margin-block-end: 1rem;
}

.admin-tasks__kind {
	margin-inline-start: 0.5rem;
}
</style>
