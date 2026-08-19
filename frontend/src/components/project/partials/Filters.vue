<template>
	<Card
		class="filters has-overflow"
		:title="hasTitle ? $t('filters.title') : ''"
		role="search"
		:show-close="showClose"
		@close="$emit('close')"
	>
		<div class="quick-filters mbe-2">
			<XButton
				v-for="qf in quickFilters"
				:key="qf.key"
				v-tooltip="qf.key === 'done' && viewHidesDone ? $t('filters.quick.doneDisabledHint') : ''"
				variant="secondary"
				:disabled="qf.key === 'done' && viewHidesDone"
				@click="applyQuickFilter(qf.query)"
			>
				{{ qf.label }}
			</XButton>
			<Multiselect
				class="assignee-quick-filter"
				:model-value="assigneeQuickFilterValue"
				:placeholder="$t('filters.quick.assignedToPlaceholder')"
				:search-results="foundAssignees"
				:show-empty="true"
				label="name"
				:autocomplete-enabled="false"
				@search="findAssignee"
				@select="applyAssigneeQuickFilter"
			>
				<template #searchResult="{option: user}">
					<User
						:avatar-size="24"
						:show-username="true"
						:user="user"
					/>
				</template>
			</Multiselect>
		</div>

		<FilterInput
			ref="filterInputRef"
			v-model="filterQuery"
			:project-id="projectId"
			class="mbe-2"
			@update:modelValue="() => change('modelValue')"
		/>
		<div
			v-if="filterFromView"
			class="tw:text-sm mbe-2"
		>
			{{ $t('filters.fromView') }}
			<code>{{ filterFromView }}</code><br>
			{{ $t('filters.fromViewBoth') }}
		</div>

		<div class="field is-flex is-flex-direction-column">
			<FancyCheckbox
				v-model="params.filter_include_nulls"
				@change="() => change('always')"
			>
				{{ $t('filters.attributes.includeNulls') }}
			</FancyCheckbox>
		</div>

		<FilterInputDocs />

		<template
			v-if="hasFooter"
			#footer
		>
			<XButton
				variant="secondary"
				class="mie-2"
				:disabled="filterQuery === ''"
				@click.prevent.stop="clearFiltersAndEmit"
			>
				{{ $t('filters.clear') }}
			</XButton>
			<XButton
				variant="primary"
				@click.prevent.stop="changeAndEmitButton"
			>
				{{ $t('filters.showResults') }}
			</XButton>
		</template>
	</Card>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import {useRoute} from 'vue-router'
import type {TaskFilterParams} from '@/services/taskCollection'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import {
	hasFilterQuery,
	transformFilterStringForApi,
} from '@/helpers/filters'
import FilterInputDocs from '@/components/input/filter/FilterInputDocs.vue'
import FilterInput from '@/components/input/filter/FilterInput.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'
import ProjectUserService from '@/services/projectUsers'
import UserService from '@/services/user'
import type {IUser} from '@/modelTypes/IUser'
import {getDisplayName} from '@/models/user'

const props = withDefaults(defineProps<{
	modelValue: TaskFilterParams,
	hasTitle?: boolean,
	hasFooter?: boolean,
	changeImmediately?: boolean,
	filterFromView?: string,
	showClose?: boolean,
}>(), {
	hasTitle: false,
	hasFooter: true,
	changeImmediately: false,
	filterFromView: undefined,
	showClose: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: TaskFilterParams],
	'showResults': [],
	'close': [],
}>()

const route = useRoute()
const projectId = computed(() => {
	if (route.name?.startsWith('project.')) {
		return Number(route.params.projectId)
	}

	return undefined
})

const params = ref<TaskFilterParams>({
	sort_by: [],
	order_by: [],
	filter: '',
	filter_include_nulls: false,
	s: '',
})

const filterQuery = ref('')
watch(
	() => [params.value.filter, params.value.s],
	() => {
		const filter = params.value.filter || ''
		const s = params.value.s || ''
		filterQuery.value = filter || s
	},
)

const labelStore = useLabelStore()
const projectStore = useProjectStore()

const filterInputRef = ref()

// Using watchDebounced to prevent the filter re-triggering itself.
watch(
	() => props.modelValue,
	(value: TaskFilterParams) => {
		params.value = {...value}
	},
	{
		immediate: true,
		deep: true,
	},
)

function change(event: 'blur' | 'modelValue' | 'always') {
	if (event !== 'always') {
		// The filter edit setting needs to save immediately, but the filter query edit in project views should 
		// only change on blur, or it will show the filter replaced for api when the query is not yet complete. 
		// This is highly confusing UX, hence we want to avoid that.
		// The approach taken here allows us to either toggle on blur or immediately, depending on the prop
		// value provided. This probably is a hacky way to do this, but it is also the most effective.
		if (props.changeImmediately && event === 'blur') {
			return
		}

		if (!props.changeImmediately && event === 'modelValue') {
			return
		}
	}

	const filter = transformFilterStringForApi(
		filterQuery.value,
		labelTitle => labelStore.getLabelByExactTitle(labelTitle)?.id || null,
		projectTitle => {
			const found = projectStore.findProjectByExactname(projectTitle)
			return found?.id || null
		},
	)

	let s = ''

	// When the filter does not contain any filter tokens, assume a simple search and redirect the input
	if (!hasFilterQuery(filter)) {
		s = filter
	}

	const newParams = {
		...params.value,
		filter: s === '' ? filter : '',
		s,
	}

	if (JSON.stringify(props.modelValue) === JSON.stringify(newParams)) {
		return
	}

	emit('update:modelValue', newParams)
}

function changeAndEmitButton() {
	change()
	emit('showResults')
}

function clearFiltersAndEmit() {
	filterQuery.value = ''
	changeAndEmitButton()
}

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()

// Only the List view ships a `done = false` base filter today (see project_view.go).
// Combining that with a user-entered `done = true` query ANDs into a permanent
// contradiction (zero results, no error) - the exact trap that makes "done" tasks
// seem to have vanished. Disable the quick filter there instead of letting it lie.
const viewHidesDone = computed(() => (props.filterFromView ?? '').includes('done'))

const quickFilters = computed(() => {
	const filters = [
		{key: 'done', label: t('filters.quick.done'), query: 'done = true'},
		{key: 'overdue', label: t('filters.quick.overdue'), query: 'dueDate < now/d && done = false'},
		{key: 'dueToday', label: t('filters.quick.dueToday'), query: 'dueDate >= now/d && dueDate < now/d+1d'},
		{key: 'dueThisWeek', label: t('filters.quick.dueThisWeek'), query: 'dueDate >= now/w && dueDate < now/w+1w'},
		{key: 'highPriority', label: t('filters.quick.highPriority'), query: 'priority >= 3'},
	]

	if (authStore.info?.username) {
		filters.push({
			key: 'assignedToMe',
			label: t('filters.quick.assignedToMe'),
			query: `assignees in ${authStore.info.username}`,
		})
	}

	return filters
})

function applyQuickFilter(query: string) {
	filterQuery.value = query
	change('always')
	emit('showResults')
}

const foundAssignees = ref<IUser[]>([])
const assigneeQuickFilterValue = ref<IUser | null>(null)
const projectUserService = new ProjectUserService()
const userService = new UserService()

async function findAssignee(query = '') {
	const response = projectId.value
		// @ts-expect-error - projectId is used for URL replacement but not part of IAbstract
		? await projectUserService.getAll({projectId: projectId.value}, {s: query}) as IUser[]
		: await userService.getAll({} as IUser, {s: query}) as IUser[]

	foundAssignees.value = response.map(u => {
		u.name = getDisplayName(u)
		return u
	})
}

function applyAssigneeQuickFilter(user: IUser) {
	assigneeQuickFilterValue.value = user
	applyQuickFilter(`assignees in ${user.username}`)
}

function focusFilterInput() {
	filterInputRef.value?.focus()
}

defineExpose({
	focusFilterInput,
})
</script>

<style lang="scss" scoped>
.quick-filters {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem;

	.assignee-quick-filter {
		max-inline-size: 220px;
	}
}
</style>
