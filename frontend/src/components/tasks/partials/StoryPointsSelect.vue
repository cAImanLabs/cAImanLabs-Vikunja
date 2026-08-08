<template>
	<div class="select">
		<select
			v-model="storyPoints"
			:disabled="disabled || undefined"
			:aria-label="$t('task.attributes.storyPoints')"
		>
			<option :value="0">
				0 pts
			</option>
			<option :value="1">
				1 pt
			</option>
			<option :value="2">
				2 pts
			</option>
			<option :value="3">
				3 pts
			</option>
			<option :value="5">
				5 pts
			</option>
			<option :value="8">
				8 pts
			</option>
			<option :value="13">
				13 pts
			</option>
			<option :value="21">
				21 pts
			</option>
		</select>
	</div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const storyPoints = defineModel<number>({
	default: 0,
})
</script>
