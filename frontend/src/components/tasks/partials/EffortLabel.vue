<template>
	<span
		v-if="!done && (showAll || effort >= minimumEffort)"
		:class="{
			'negligible': effort <= 5,
			'not-so-high': effort > 5 && effort < 8,
			'high-effort': effort >= 8
		}"
		class="effort-label"
	>
		<span>
			{{  effort }}
		</span>
	</span>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useAuthStore} from '@/stores/auth'
	
withDefaults(defineProps<{
	effort: number,
	showAll?: boolean,
	done?: boolean
}>(), {
	showAll: false,
	done: false,
})

const authStore = useAuthStore()

const minimumEffort = computed(() => {
	return 0
})
</script>

<style lang="scss" scoped>
.high-effort {
	color: var(--danger);
	inline-size: auto !important; // To override the width set in tasks
	background-color: #434350;
	padding: 2px 8px;
	border-radius: 8px;
}

.not-so-high {
	color: var(--warning);
	background-color: #434350;
	padding: 2px 8px;
	border-radius: 8px;
}

.negligible {
	color: var(--info);
	background-color: #434350;
	padding: 2px 8px;
	border-radius: 8px;
}

.icon {
	vertical-align: top;
	inline-size: auto !important;
	padding-inline-end: .5rem;
}
</style>
