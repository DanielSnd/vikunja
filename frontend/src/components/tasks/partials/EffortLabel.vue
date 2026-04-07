<template>
	<span
		v-if="!done && (showAll || effort >= minimumEffort)"
		:class="{
			'negligible': effort <= 3,
			'not-so-high': effort > 3 && effort < 7,
			'high-effort': effort >= 7
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
.high-priority {
	color: var(--warning);
	inline-size: auto !important; // To override the width set in tasks
	background-color: #434350;
	padding: 2px 8px;
	border-radius: 8px;
}

.not-so-high {
	color: var(--danger);
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
