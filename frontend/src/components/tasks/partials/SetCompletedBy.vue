<template>
	<Multiselect
		v-model="selected"
		class="set-completed-by"
		:loading="userSearchLoading"
		:placeholder="$t('admin.searchUsersPlaceholder')"
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
import {ref, watch} from 'vue'
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

const selected = ref<IUser | null>(null)
watch(() => props.modelValue, value => {
	selected.value = value
}, {immediate: true})

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
