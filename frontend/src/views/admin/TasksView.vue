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
				<table class="table has-actions is-striped is-hoverable is-fullwidth">
					<thead>
						<tr>
							<th>{{ $t('misc.id') }}</th>
							<th>{{ $t('task.attributes.title') }}</th>
							<th>{{ $t('task.attributes.project') }}</th>
							<th>{{ $t('task.attributes.priority') }}</th>
							<th>{{ $t('task.attributes.dueDate') }}</th>
							<th>{{ $t('task.attributes.done') }}</th>
							<th>{{ $t('task.attributes.created') }}</th>
						</tr>
					</thead>
					<tbody>
						<tr
							v-for="t in tasks"
							:key="t.id"
						>
							<td>{{ t.id }}</td>
							<td>
								<RouterLink :to="{name: 'task.detail', params: {id: t.id}}">
									{{ t.title }}
								</RouterLink>
								<TaskKindLabel
									:kind="t.kind"
									class="admin-tasks__kind"
								/>
							</td>
							<td>{{ t.projectTitle }}</td>
							<td>
								<PriorityLabel
									:priority="t.priority"
									:done="t.done"
								/>
							</td>
							<td>
								<TimeDisplay
									v-if="t.dueDate"
									:date="t.dueDate"
								/>
							</td>
							<td>{{ t.done ? $t('task.attributes.done') : '' }}</td>
							<td>
								<TimeDisplay :date="t.created" />
							</td>
						</tr>
					</tbody>
				</table>
				<PaginationEmit
					v-if="totalPages > 1"
					:total-pages="totalPages"
					:current-page="currentPage"
					@pageChanged="goToPage"
				/>
			</template>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {useDebounceFn} from '@vueuse/core'
import type {IAdminTask} from '@/modelTypes/IAdminTask'
import AdminTaskService from '@/services/admin/taskService'
import AdminTaskModel from '@/models/adminTask'
import Card from '@/components/misc/Card.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import FormInput from '@/components/input/FormInput.vue'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'
import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import TaskKindLabel from '@/components/tasks/partials/TaskKindLabel.vue'
import {error} from '@/message'

const adminTaskService = new AdminTaskService()

const tasks = ref<IAdminTask[]>([])
const loading = ref(false)
const searchTerm = ref('')
const currentPage = ref(1)
const totalPages = ref(1)

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

onMounted(load)
</script>

<style lang="scss" scoped>
.admin-tasks__toolbar {
	display: flex;
	gap: 0.5rem;
	margin-block-end: 1rem;
}

.admin-tasks__kind {
	margin-inline-start: 0.5rem;
}
</style>
