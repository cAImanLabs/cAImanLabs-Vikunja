<script setup lang="ts">
import { computed } from 'vue'
import { useColorScheme } from '@/composables/useColorScheme'

import LogoIcon from '@/assets/logo-icon.png'

const { isDark } = useColorScheme()

const CustomLogo = computed(() => {
	const lightLogo = window.CUSTOM_LOGO_URL
	const darkLogo = window.CUSTOM_LOGO_URL_DARK

	if (!lightLogo && !darkLogo) return ''
	if (!darkLogo) return lightLogo
	if (!lightLogo) return darkLogo

	return isDark.value ? darkLogo : lightLogo
})
</script>

<template>
	<div class="logo-wrapper">
		<img
			:src="CustomLogo || LogoIcon"
			alt="cAImanDesk"
			class="logo"
		>
		<span
			v-if="!CustomLogo"
			class="logo-text"
		>cAImanDesk</span>
	</div>
</template>

<style lang="scss" scoped>
.logo-wrapper {
	display: flex;
	align-items: center;
	gap: .5rem;
	max-inline-size: 168px;
}

.logo {
	max-block-size: 40px;
	max-inline-size: 40px;
}

.logo-text {
	color: var(--logo-text-color);
	font-weight: 700;
	font-size: 1.1rem;
	white-space: nowrap;
}
</style>
