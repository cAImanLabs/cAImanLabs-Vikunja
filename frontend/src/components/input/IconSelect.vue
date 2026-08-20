<template>
	<div class="icon-select">
		<Popup :ignore-click-classes="['icon-select-trigger']">
			<template #trigger="{toggle, isOpen}">
				<button
					type="button"
					class="icon-select-trigger"
					:class="{'is-open': isOpen}"
					:disabled="disabled || undefined"
					:aria-label="ariaLabel"
					:aria-expanded="isOpen"
					aria-haspopup="listbox"
					@click="toggle()"
				>
					<span
						v-if="selected"
						class="icon-select-icon"
						:style="{color: selected.color}"
					>
						<Icon :icon="selected.icon" />
					</span>
					<span class="icon-select-label">{{ selected?.label }}</span>
					<Icon
						icon="chevron-down"
						class="icon-select-chevron"
					/>
				</button>
			</template>
			<template #content="{close}">
				<ul
					class="icon-select-list"
					role="listbox"
					:aria-label="ariaLabel"
				>
					<li
						v-for="option in options"
						:key="option.value"
					>
						<button
							type="button"
							class="icon-select-option"
							role="option"
							:aria-selected="option.value === modelValue"
							:class="{'is-selected': option.value === modelValue}"
							@click="select(option.value); close()"
						>
							<span
								class="icon-select-icon"
								:style="{color: option.color}"
							>
								<Icon :icon="option.icon" />
							</span>
							<span class="icon-select-label">{{ option.label }}</span>
							<Icon
								v-if="option.value === modelValue"
								icon="check"
								class="icon-select-check"
							/>
						</button>
					</li>
				</ul>
			</template>
		</Popup>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'
import Popup from '@/components/misc/Popup.vue'

export interface IconSelectOption {
	value: string | number
	label: string
	icon: IconProp
	color?: string
}

const props = withDefaults(defineProps<{
	modelValue: string | number,
	options: IconSelectOption[],
	disabled?: boolean,
	ariaLabel?: string,
}>(), {
	disabled: false,
	ariaLabel: undefined,
})

const emit = defineEmits<{
	'update:modelValue': [value: string | number],
}>()

const selected = computed(() => props.options.find(o => o.value === props.modelValue))

function select(value: string | number) {
	if (value === props.modelValue) {
		return
	}
	emit('update:modelValue', value)
}
</script>

<style lang="scss" scoped>
.icon-select {
	position: relative;
}

.icon-select-trigger {
	display: flex;
	align-items: center;
	gap: .5rem;
	inline-size: 100%;
	padding: .5rem .75rem;
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	color: var(--text);
	cursor: pointer;
	transition: border-color $transition;
	text-align: start;

	&:hover {
		border-color: var(--grey-300);
	}

	&:focus-visible,
	&.is-open {
		border-color: var(--primary);
		outline: none;
	}

	&:disabled {
		cursor: not-allowed;
		opacity: .5;
	}
}

.icon-select-label {
	flex: 1;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.icon-select-icon {
	display: inline-flex;
	inline-size: 1rem;
	flex-shrink: 0;
}

.icon-select-chevron {
	flex-shrink: 0;
	color: var(--grey-400);
	inline-size: auto !important;
}

.icon-select-list {
	min-inline-size: 12rem;
	max-block-size: 16rem;
	overflow-y: auto;
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	box-shadow: var(--shadow-md);
	padding: .35rem;
}

.icon-select-option {
	display: flex;
	align-items: center;
	gap: .5rem;
	inline-size: 100%;
	padding: .5rem .6rem;
	background: transparent;
	border: none;
	border-radius: $radius;
	color: var(--text);
	cursor: pointer;
	text-align: start;
	font-family: $family-sans-serif;

	&:hover,
	&:focus-visible {
		background: var(--grey-100);
		outline: none;
	}

	&.is-selected {
		font-weight: bold;
	}
}

.icon-select-check {
	flex-shrink: 0;
	color: var(--primary);
	inline-size: auto !important;
}
</style>
