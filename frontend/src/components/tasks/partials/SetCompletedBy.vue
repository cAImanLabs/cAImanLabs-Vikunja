<template>
	<Multiselect
		v-model="selected"
		class="set-completed-by"
		:loading="userSearchLoading"
		:placeholder="placeholder"
		:search-results="userResults"
		label="username"
		@search="searchUsers"
		@select="reassign"
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
</template>

<script setup lang="ts">
import {ref, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'
import AdminUserService from '@/services/admin/userService'
import AdminTaskService from '@/services/admin/taskService'
import AdminUserModel from '@/models/adminUser'
import type {IAdminUser} from '@/modelTypes/IAdminUser'
import type {IUser} from '@/modelTypes/IUser'
import {error, success} from '@/message'

const props = defineProps<{
	modelValue: IUser | null,
	taskId: number,
}>()

const emit = defineEmits<{
	'update:modelValue': [value: IUser | null],
}>()

const {t} = useI18n({useScope: 'global'})

const adminUserService = new AdminUserService()
const adminTaskService = new AdminTaskService()

// Deliberately not seeded from props.modelValue: pre-filling the search box
// with the current user's name looks like a completed search that found
// nothing (there's no query behind it yet), which reads as "that user
// doesn't exist". Starting empty makes it unambiguously a fresh search field;
// the current value is still shown via the placeholder and elsewhere on the page.
const selected = ref<IUser | null>(null)

const placeholder = computed(() => props.modelValue
	? t('admin.tasks.currentlyCompletedByPlaceholder', {username: props.modelValue.username})
	: t('admin.searchUsersPlaceholder'))

const userResults = ref<IAdminUser[]>([])
const userSearchLoading = ref(false)

async function searchUsers(query: string) {
	if (!query || query.length < 2) {
		userResults.value = []
		return
	}
	userSearchLoading.value = true
	try {
		userResults.value = await adminUserService.getAll(new AdminUserModel(), {s: query})
	} catch (e) {
		error(e)
	} finally {
		userSearchLoading.value = false
	}
}

async function reassign(user: IUser) {
	try {
		const updated = await adminTaskService.reassignCompletedBy(props.taskId, user.id)
		emit('update:modelValue', updated.completedBy)
		success({message: t('admin.tasks.reassignedSuccess')})
	} catch (e) {
		error(e)
	}
}
</script>
