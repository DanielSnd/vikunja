<template>
	<Multiselect
		v-model="milestone"
		:disabled="disabled"
		:loading="milestoneService.loading"
		:placeholder="$t('task.milestone.placeholder')"
		:search-results="foundMilestones"
		label="name"
		:show-empty="true"
		@search="findMilestones"
		@select="selectMilestone"
	>
		<template #tag="{item}">
			<span
				class="tag milestone-tag"
				:style="getMilestoneStyles(item)"
			>
				<span>{{ item.name }}</span>
			</span>
		</template>
		<template #searchResult="{option}">
			<span
				class="tag search-result"
				:style="getMilestoneStyles(option)"
			>
				<span>{{ option.name }}</span>
			</span>
		</template>
	</Multiselect>
</template>

<script setup lang="ts">
import {computed, ref, shallowReactive, watch} from 'vue'

import Multiselect from '@/components/input/Multiselect.vue'

import type {IMilestone} from '@/modelTypes/IMilestone'
import MilestoneService from '@/services/milestone'

const props = withDefaults(defineProps<{
	modelValue: IMilestone | null,
	projectId: number,
	disabled?: boolean,
}>(), {
	disabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: IMilestone | null],
}>()

const milestoneService = shallowReactive(new MilestoneService())
const milestone = ref<IMilestone | null>(null)
const searchQuery = ref('')
const loadedMilestones = ref<IMilestone[]>([])

watch(
	() => props.modelValue,
	value => {
		milestone.value = value
	},
	{
		immediate: true,
		deep: true,
	},
)

const foundMilestones = computed(() => {
	const query = searchQuery.value.trim().toLowerCase()
	const seen = new Set<number>()
	const milestones = [
		...(milestone.value ? [milestone.value] : []),
		...loadedMilestones.value,
	]

	return milestones.filter(currentMilestone => {
		if (!currentMilestone || seen.has(currentMilestone.id)) {
			return false
		}
		seen.add(currentMilestone.id)

		if (query === '') {
			return true
		}

		return currentMilestone.name.toLowerCase().includes(query)
	})
})

function getMilestoneStyles(milestone: IMilestone) {
	if (!milestone?.hexColor) {
		return {}
	}

	return {
		backgroundColor: milestone.hexColor,
		color: 'var(--white)',
	}
}

async function findMilestones(query: string) {
	searchQuery.value = query
	loadedMilestones.value = await milestoneService.getAll(
		{projectId: props.projectId},
		{
			s: query,
			includeParents: true,
		},
	)
}

function selectMilestone(selectedMilestone: IMilestone) {
	emit('update:modelValue', selectedMilestone)
}
</script>

<style lang="scss" scoped>
.tag {
	margin: .25rem !important;
}

.search-result {
	margin: 0 !important;
}

:deep(.input-wrapper) {
	padding: .25rem !important;
}

:deep(input.input) {
	padding: 0 .5rem;
}
</style>
