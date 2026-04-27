<template>
	<div class="task-content">
		<div class="task-content__status">
			<CustomTransition name="fade">
				<span
					v-if="loading && saving"
					class="is-small is-inline-flex"
				>
					<span class="loader is-inline-block mie-2" />
					{{ $t('misc.saving') }}
				</span>
				<span
					v-else-if="!loading && saved"
					class="is-small has-text-success"
				>
					<Icon icon="check" />
					{{ $t('misc.saved') }}
				</span>
			</CustomTransition>
		</div>
		<Editor
			ref="editorRef"
			v-model="content"
			class="tiptap__task-description"
			:is-edit-enabled="canWrite"
			:upload-callback="uploadCallback"
			:placeholder="$t('task.description.placeholder')"
			:show-save="true"
			edit-shortcut="KeyE"
			:enable-discard-shortcut="true"
			:enable-mentions="true"
			:mention-project-id="modelValue.projectId"
			:storage-key="contentStorageKey"
			:hide-edit-button="true"
			edit-trigger="click"
			@update:modelValue="saveWithDelay"
			@save="save"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount, ref, watch} from 'vue'
import {onBeforeRouteLeave} from 'vue-router'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import Editor from '@/components/input/AsyncEditor'

import {clearEditorDraft} from '@/helpers/editorDraftStorage'
import type {ITask} from '@/modelTypes/ITask'
import {useTaskStore} from '@/stores/tasks'
import {buildTaskEditorContent, splitTaskEditorContent} from './taskContent'

export type AttachmentUploadFunction = (file: File, onSuccess: (attachmentUrl: string) => void) => Promise<string>

const props = defineProps<{
	modelValue: ITask,
	attachmentUpload: AttachmentUploadFunction,
	canWrite: boolean,
}>()

const emit = defineEmits<{
	'update:modelValue': [value: ITask]
}>()

const content = ref('')
const editorRef = ref<InstanceType<typeof Editor> | null>(null)
const hasChanges = ref(false)
watch(() => [props.modelValue.id, props.modelValue.title, props.modelValue.description], () => {
	if (hasChanges.value) {
		return
	}

	content.value = buildTaskEditorContent({
		title: props.modelValue.title,
		description: props.modelValue.description,
	})
}, {immediate: true})

watch(() => props.modelValue.id, () => {
	hasChanges.value = false
})

const saved = ref(false)

// Since loading is global state, this variable ensures we're only showing the saving icon when saving the description.
const saving = ref(false)

const taskStore = useTaskStore()
const loading = computed(() => taskStore.isLoading)

const changeTimeout = ref<ReturnType<typeof setTimeout> | null>(null)

const contentStorageKey = computed(() => `task-content-${props.modelValue.id}`)

async function saveWithDelay() {
	if (content.value === buildTaskEditorContent(props.modelValue)) {
		hasChanges.value = false
		if (changeTimeout.value !== null) {
			clearTimeout(changeTimeout.value)
		}
		return
	}

	hasChanges.value = true
	if (changeTimeout.value !== null) {
		clearTimeout(changeTimeout.value)
	}

	changeTimeout.value = setTimeout(async () => {
		await save()
	}, 5000)
}

onBeforeUnmount(async () => {
	await save() // Save before unmounting to handle modal race condition
	if (changeTimeout.value !== null) {
		clearTimeout(changeTimeout.value)
	}
})

onBeforeRouteLeave(() => save())

async function save() {
	if (!hasChanges.value) {
		return
	}

	hasChanges.value = false
	if (changeTimeout.value !== null) {
		clearTimeout(changeTimeout.value)
	}
	saved.value = false
	saving.value = true

	try {
		const {title, description} = splitTaskEditorContent(content.value)

		const updated = await taskStore.update({
			...props.modelValue,
			title,
			description,
		})
		emit('update:modelValue', updated)

		// Clear draft from localStorage when saved successfully
		clearEditorDraft(contentStorageKey.value)
		content.value = buildTaskEditorContent(updated)

		saved.value = true
		setTimeout(() => {
			saved.value = false
		}, 2000)
	} catch (error) {
		// If the task was deleted (404), silently skip saving
		if (error?.response?.status === 404) {
			return
		}
		hasChanges.value = true
		// Re-throw other errors
		throw error
	} finally {
		saving.value = false
	}
}

async function uploadCallback(files: File[] | FileList): Promise<string[]> {
	const uploadPromises: Promise<string>[] = []

	files.forEach((file: File) => {
		const promise = new Promise<string>((resolve) => {
			props.attachmentUpload(file, (uploadedFileUrl: string) => resolve(uploadedFileUrl))
		})

		uploadPromises.push(promise)
	})

	return await Promise.all(uploadPromises)
}

function insertImage(url: string) {
	editorRef.value?.insertImage(url)
}

defineExpose({
	insertImage,
})
</script>

<style lang="scss" scoped>
.task-content__status {
	display: flex;
	justify-content: flex-end;
	min-block-size: 1.5rem;
}

.tiptap__task-description {
	:deep(.tiptap__editor) {
		min-block-size: auto;
	}
	:deep(.ProseMirror) {
		min-height: 500px;
	}
	:deep(.ProseMirror > h2:first-child) {
		font-size: 1.75rem;
		line-height: 1.2;
		font-weight: 700;
		margin-block-start: 0;
	}
}
</style>
