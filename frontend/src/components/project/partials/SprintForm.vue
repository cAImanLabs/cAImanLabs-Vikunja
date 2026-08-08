<template>
	<form @submit.prevent="emit('submit')">
		<FormField
			id="sprint-title"
			v-model="form.title"
			v-focus
			:label="$t('project.views.title')"
			:placeholder="$t('project.share.links.namePlaceholder')"
			:error="titleValid ? null : $t('project.sprint.titleRequired')"
			@blur="validateTitle"
		/>

		<FormField :label="$t('project.sprint.goal')">
			<template #default="{id}">
				<textarea
					:id="id"
					v-model="form.goal"
					class="textarea"
					rows="2"
					:placeholder="$t('project.sprint.goalPlaceholder')"
				/>
			</template>
		</FormField>

		<div class="columns">
			<div class="column">
				<label class="label">{{ $t('project.sprint.startDate') }}</label>
				<Datepicker
					v-model="form.startDate"
					:choose-date-label="$t('project.sprint.startDate')"
				/>
			</div>
			<div class="column">
				<label class="label">{{ $t('project.sprint.endDate') }}</label>
				<Datepicker
					v-model="form.endDate"
					:choose-date-label="$t('project.sprint.endDate')"
				/>
			</div>
		</div>

		<FormField :label="$t('project.sprint.status')">
			<template #default="{id}">
				<div class="select">
					<select
						:id="id"
						v-model="form.status"
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
			</template>
		</FormField>

		<slot name="extra-fields" />

		<div class="is-flex is-justify-content-end">
			<XButton
				variant="tertiary"
				class="mie-2"
				@click="emit('cancel')"
			>
				{{ $t('misc.cancel') }}
			</XButton>
			<XButton
				:loading="loading"
				type="submit"
			>
				{{ $t('misc.save') }}
			</XButton>
		</div>
	</form>
</template>

<script setup lang="ts">
import {ref} from 'vue'

import FormField from '@/components/input/FormField.vue'
import Datepicker from '@/components/input/Datepicker.vue'

import {SPRINT_STATUSES} from '@/modelTypes/ISprint'
import type {ISprintFormData} from '@/modelTypes/ISprint'

withDefaults(defineProps<{
	loading?: boolean,
}>(), {
	loading: false,
})

const emit = defineEmits<{
	submit: [],
	cancel: [],
}>()

const form = defineModel<ISprintFormData>({required: true})

const titleValid = ref(true)

function validateTitle() {
	titleValid.value = !!form.value.title
}
</script>
