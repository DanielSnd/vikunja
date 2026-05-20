<script setup lang="ts">
import {computed, onMounted, reactive, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import Card from '@/components/misc/Card.vue'
import Message from '@/components/misc/Message.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import FormField from '@/components/input/FormField.vue'
import FormInput from '@/components/input/FormInput.vue'
import XButton from '@/components/input/Button.vue'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'
import {formatDisplayDate} from '@/helpers/time/formatDate'
import UserReportService from '@/services/userReport'
import BucketService from '@/services/bucket'
import ProjectViewService from '@/services/projectViews'
import type {IProject} from '@/modelTypes/IProject'
import type {IUserReportApplication, IUserReportToken} from '@/modelTypes/IUserReport'
import type {IBucket} from '@/modelTypes/IBucket'
import type {IProjectView} from '@/modelTypes/IProjectView'
import {PROJECT_VIEW_KINDS} from '@/modelTypes/IProjectView'
import {PRIORITIES} from '@/constants/priorities'

type ProjectField = 'projectId' | 'criticalProjectId' | 'highProjectId' | 'lowProjectId'
type BucketField = 'bucketId' | 'criticalBucketId' | 'highBucketId' | 'lowBucketId'

const bucketFieldByProjectField: Record<ProjectField, BucketField> = {
	projectId: 'bucketId',
	criticalProjectId: 'criticalBucketId',
	highProjectId: 'highBucketId',
	lowProjectId: 'lowBucketId',
}

defineOptions({name: 'UserSettingsUserReports'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.userReports.title')} - ${t('user.settings.title')}`)

const projectStore = useProjectStore()
const service = new UserReportService()
const bucketService = new BucketService()
const projectViewService = new ProjectViewService()

const loading = ref(false)
const applications = ref<IUserReportApplication[]>([])
const tokensByApplication = ref<Record<number, IUserReportToken[]>>({})
const revealedAccessKeys = ref<Record<number, string>>({})
const createdTokenSecrets = ref<Record<number, string>>({})
const newTokenLabels = reactive<Record<number, string>>({})
const bucketOptionsByProject = ref<Record<number, Array<{id: number, title: string}>>>({})
const createForm = reactive({
	name: '',
	projectId: 0,
})

const uploadSizeFactor = 1024 * 1024

const hasApplications = computed(() => applications.value.length > 0)
const priorityOptions = computed(() => [
	{value: PRIORITIES.UNSET, label: t('task.priority.unset')},
	{value: PRIORITIES.LOW, label: t('task.priority.low')},
	{value: PRIORITIES.MEDIUM, label: t('task.priority.medium')},
	{value: PRIORITIES.HIGH, label: t('task.priority.high')},
	{value: PRIORITIES.URGENT, label: t('task.priority.urgent')},
	{value: PRIORITIES.DO_NOW, label: t('task.priority.doNow')},
])

async function load() {
	loading.value = true
	try {
		applications.value = await service.getAll()
		await Promise.all(collectProjectIds(applications.value).map(projectId => ensureBucketsForProject(projectId)))
		for (const app of applications.value) {
			tokensByApplication.value[app.id] = await service.getTokens(app.id)
		}
	} finally {
		loading.value = false
	}
}

function getProject(projectId: number): IProject | null {
	return projectStore.projects[projectId] ?? null
}

function setProject(application: IUserReportApplication, key: ProjectField, project: IProject | null) {
	application[key] = project?.id ?? 0
	application[bucketFieldByProjectField[key]] = 0
	void ensureBucketsForProject(getTargetProjectId(application, bucketFieldByProjectField[key]))
}

function getTargetProjectId(application: IUserReportApplication, bucketKey: BucketField): number {
	switch (bucketKey) {
		case 'criticalBucketId':
			return application.criticalProjectId || application.projectId
		case 'highBucketId':
			return application.highProjectId || application.projectId
		case 'lowBucketId':
			return application.lowProjectId || application.projectId
		default:
			return application.projectId
	}
}

function getBucketOptions(application: IUserReportApplication, bucketKey: BucketField): Array<{id: number, title: string}> {
	const projectId = getTargetProjectId(application, bucketKey)
	return bucketOptionsByProject.value[projectId] ?? []
}

async function ensureBucketsForProject(projectId: number) {
	if (projectId === 0 || bucketOptionsByProject.value[projectId]) {
		return
	}

	const views = await projectViewService.getAll({projectId} as IProjectView)
	const kanbanViews = views.filter(view =>
		view.viewKind === PROJECT_VIEW_KINDS.KANBAN &&
		view.bucketConfigurationMode === 'manual',
	)

	const bucketOptions = (await Promise.all(kanbanViews.map(async view => {
		const buckets = await bucketService.getAll({
			projectId,
			projectViewId: view.id,
		} as IBucket)

		return buckets.map(bucket => ({
			id: bucket.id,
			title: `${view.title} / ${bucket.title}`,
		}))
	}))).flat()

	bucketOptionsByProject.value[projectId] = bucketOptions
}

function collectProjectIds(userReportApplications: IUserReportApplication[]): number[] {
	return [...new Set(userReportApplications.flatMap(application => [
		application.projectId,
		application.criticalProjectId || application.projectId,
		application.highProjectId || application.projectId,
		application.lowProjectId || application.projectId,
	]).filter(projectId => projectId !== 0))]
}

function getUploadSizeMb(application: IUserReportApplication): string {
	if (application.maxUploadSize === 0) {
		return ''
	}

	return String(application.maxUploadSize / uploadSizeFactor)
}

function setUploadSizeMb(application: IUserReportApplication, value: string) {
	const parsed = Number(value)
	application.maxUploadSize = Number.isFinite(parsed) && parsed > 0 ? parsed * uploadSizeFactor : 0
}

async function createApplication() {
	if (createForm.name.trim() === '' || createForm.projectId === 0) {
		return
	}

	const created = await service.createApplication({
		name: createForm.name,
		projectId: createForm.projectId,
	})

	applications.value.push(created.application)
	tokensByApplication.value[created.application.id] = [created.defaultReportToken]
	revealedAccessKeys.value[created.application.id] = created.accessKey
	createdTokenSecrets.value[created.application.id] = created.defaultReportToken.token ?? ''
	createForm.name = ''
	createForm.projectId = 0
	success({message: t('user.settings.userReports.createSuccess')})
}

async function saveApplication(application: IUserReportApplication) {
	const updated = await service.updateApplication(application)
	const index = applications.value.findIndex(app => app.id === updated.id)
	if (index !== -1) {
		applications.value[index] = updated
	}
	success({message: t('user.settings.userReports.saveSuccess')})
}

async function regenerateAccessKey(application: IUserReportApplication) {
	revealedAccessKeys.value[application.id] = await service.regenerateAccessKey(application.id)
	success({message: t('user.settings.userReports.accessKeyRegenerated')})
}

async function createToken(application: IUserReportApplication) {
	const label = (newTokenLabels[application.id] ?? '').trim()
	if (label === '') {
		return
	}

	const token = await service.createToken(application.id, label)
	tokensByApplication.value[application.id] = [
		...(tokensByApplication.value[application.id] ?? []),
		token,
	]
	createdTokenSecrets.value[application.id] = token.token ?? ''
	newTokenLabels[application.id] = ''
	success({message: t('user.settings.userReports.tokenCreated')})
}

async function toggleToken(application: IUserReportApplication, token: IUserReportToken) {
	const updated = await service.updateToken(application.id, token)
	tokensByApplication.value[application.id] = (tokensByApplication.value[application.id] ?? []).map(existing =>
		existing.id === updated.id ? updated : existing,
	)
	success({message: t('user.settings.userReports.tokenUpdated')})
}

async function deleteApplication(application: IUserReportApplication) {
	if (!window.confirm(t('user.settings.userReports.deleteConfirm', {name: application.name}))) {
		return
	}

	await service.deleteApplication(application.id)
	applications.value = applications.value.filter(app => app.id !== application.id)
	delete tokensByApplication.value[application.id]
	delete revealedAccessKeys.value[application.id]
	delete createdTokenSecrets.value[application.id]
	success({message: t('user.settings.userReports.deleteSuccess')})
}

onMounted(() => {
	load()
})
</script>

<template>
	<Card
		:title="$t('user.settings.userReports.title')"
		:loading="loading"
	>
		<p class="mb-4">
			{{ $t('user.settings.userReports.description') }}
		</p>

		<div class="user-reports-create">
			<FormField
				:label="$t('user.settings.userReports.fields.name')"
				layout="two-col"
			>
				<FormInput
					v-model="createForm.name"
					type="text"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.userReports.fields.defaultProject')"
				layout="two-col"
			>
				<ProjectSearch
					:model-value="getProject(createForm.projectId)"
					@update:modelValue="createForm.projectId = $event?.id ?? 0"
				/>
			</FormField>
			<XButton
				class="mt-2"
				:disabled="createForm.name.trim() === '' || createForm.projectId === 0"
				@click="createApplication"
			>
				{{ $t('user.settings.userReports.create') }}
			</XButton>
		</div>

		<p
			v-if="!hasApplications"
			class="mt-4"
		>
			{{ $t('user.settings.userReports.empty') }}
		</p>
	</Card>

	<Card
		v-for="application in applications"
		:key="application.id"
		:title="application.name"
		class="mt-4"
	>
		<div class="field-group">
			<FormField
				:label="$t('user.settings.userReports.fields.name')"
				layout="two-col"
			>
				<FormInput
					v-model="application.name"
					type="text"
				/>
			</FormField>

			<FormField
				:label="$t('user.settings.userReports.fields.defaultProject')"
				layout="two-col"
			>
				<ProjectSearch
					:model-value="getProject(application.projectId)"
					@update:modelValue="setProject(application, 'projectId', $event)"
				/>
			</FormField>

			<FormField
				:label="$t('user.settings.userReports.fields.maxUploadSize')"
				layout="two-col"
			>
				<FormInput
					:model-value="getUploadSizeMb(application)"
					type="number"
					min="0"
					:placeholder="$t('user.settings.userReports.maxUploadSizePlaceholder')"
					@update:modelValue="setUploadSizeMb(application, $event)"
				/>
			</FormField>

			<FormField
				:label="$t('user.settings.userReports.fields.defaultBucket')"
				layout="two-col"
			>
				<div class="select is-fullwidth">
					<select
						v-model.number="application.bucketId"
						:disabled="getBucketOptions(application, 'bucketId').length === 0"
					>
						<option :value="0">
							{{ $t('user.settings.userReports.noBucket') }}
						</option>
						<option
							v-for="bucket in getBucketOptions(application, 'bucketId')"
							:key="bucket.id"
							:value="bucket.id"
						>
							{{ bucket.title }}
						</option>
					</select>
				</div>
			</FormField>
		</div>

		<h4 class="title is-6 mt-5 mb-3">
			{{ $t('user.settings.userReports.severity.title') }}
		</h4>

		<div class="field-group">
			<FormField
				:label="$t('user.settings.userReports.severity.criticalProject')"
				layout="two-col"
			>
				<ProjectSearch
					:model-value="getProject(application.criticalProjectId)"
					@update:modelValue="setProject(application, 'criticalProjectId', $event)"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.userReports.severity.criticalBucket')"
				layout="two-col"
			>
				<div class="select is-fullwidth">
					<select
						v-model.number="application.criticalBucketId"
						:disabled="getBucketOptions(application, 'criticalBucketId').length === 0"
					>
						<option :value="0">
							{{ $t('user.settings.userReports.noBucket') }}
						</option>
						<option
							v-for="bucket in getBucketOptions(application, 'criticalBucketId')"
							:key="bucket.id"
							:value="bucket.id"
						>
							{{ bucket.title }}
						</option>
					</select>
				</div>
			</FormField>
			<FormField
				:label="$t('user.settings.userReports.severity.criticalPriority')"
				layout="two-col"
			>
				<div class="select is-fullwidth">
					<select v-model.number="application.criticalPriority">
						<option
							v-for="priority in priorityOptions"
							:key="`critical-${priority.value}`"
							:value="priority.value"
						>
							{{ priority.label }}
						</option>
					</select>
				</div>
			</FormField>

			<FormField
				:label="$t('user.settings.userReports.severity.highProject')"
				layout="two-col"
			>
				<ProjectSearch
					:model-value="getProject(application.highProjectId)"
					@update:modelValue="setProject(application, 'highProjectId', $event)"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.userReports.severity.highBucket')"
				layout="two-col"
			>
				<div class="select is-fullwidth">
					<select
						v-model.number="application.highBucketId"
						:disabled="getBucketOptions(application, 'highBucketId').length === 0"
					>
						<option :value="0">
							{{ $t('user.settings.userReports.noBucket') }}
						</option>
						<option
							v-for="bucket in getBucketOptions(application, 'highBucketId')"
							:key="bucket.id"
							:value="bucket.id"
						>
							{{ bucket.title }}
						</option>
					</select>
				</div>
			</FormField>
			<FormField
				:label="$t('user.settings.userReports.severity.highPriority')"
				layout="two-col"
			>
				<div class="select is-fullwidth">
					<select v-model.number="application.highPriority">
						<option
							v-for="priority in priorityOptions"
							:key="`high-${priority.value}`"
							:value="priority.value"
						>
							{{ priority.label }}
						</option>
					</select>
				</div>
			</FormField>

			<FormField
				:label="$t('user.settings.userReports.severity.lowProject')"
				layout="two-col"
			>
				<ProjectSearch
					:model-value="getProject(application.lowProjectId)"
					@update:modelValue="setProject(application, 'lowProjectId', $event)"
				/>
			</FormField>
			<FormField
				:label="$t('user.settings.userReports.severity.lowBucket')"
				layout="two-col"
			>
				<div class="select is-fullwidth">
					<select
						v-model.number="application.lowBucketId"
						:disabled="getBucketOptions(application, 'lowBucketId').length === 0"
					>
						<option :value="0">
							{{ $t('user.settings.userReports.noBucket') }}
						</option>
						<option
							v-for="bucket in getBucketOptions(application, 'lowBucketId')"
							:key="bucket.id"
							:value="bucket.id"
						>
							{{ bucket.title }}
						</option>
					</select>
				</div>
			</FormField>
			<FormField
				:label="$t('user.settings.userReports.severity.lowPriority')"
				layout="two-col"
			>
				<div class="select is-fullwidth">
					<select v-model.number="application.lowPriority">
						<option
							v-for="priority in priorityOptions"
							:key="`low-${priority.value}`"
							:value="priority.value"
						>
							{{ priority.label }}
						</option>
					</select>
				</div>
			</FormField>
		</div>

		<div class="user-reports-actions mt-4">
			<XButton @click="saveApplication(application)">
				{{ $t('misc.save') }}
			</XButton>
			<XButton
				class="is-outlined"
				@click="regenerateAccessKey(application)"
			>
				{{ $t('user.settings.userReports.regenerateAccessKey') }}
			</XButton>
			<XButton
				class="is-danger is-outlined"
				@click="deleteApplication(application)"
			>
				{{ $t('misc.delete') }}
			</XButton>
		</div>

		<Message
			v-if="revealedAccessKeys[application.id]"
			class="mt-4"
			variant="success"
		>
			<strong>{{ $t('user.settings.userReports.accessKey') }}:</strong> {{ revealedAccessKeys[application.id] }}
		</Message>

		<h4 class="title is-6 mt-5 mb-3">
			{{ $t('user.settings.userReports.tokens.title') }}
		</h4>

		<Message
			v-if="createdTokenSecrets[application.id]"
			class="mb-4"
			variant="success"
		>
			<strong>{{ $t('user.settings.userReports.tokens.latestToken') }}:</strong> {{ createdTokenSecrets[application.id] }}
		</Message>

		<div class="field is-grouped">
			<div class="control is-expanded">
				<FormInput
					v-model="newTokenLabels[application.id]"
					type="text"
					:placeholder="$t('user.settings.userReports.tokens.newLabel')"
				/>
			</div>
			<div class="control">
				<XButton @click="createToken(application)">
					{{ $t('user.settings.userReports.tokens.create') }}
				</XButton>
			</div>
		</div>

		<table
			v-if="(tokensByApplication[application.id] ?? []).length > 0"
			class="table is-fullwidth mt-3"
		>
			<thead>
				<tr>
					<th>{{ $t('user.settings.userReports.tokens.label') }}</th>
					<th>{{ $t('user.settings.userReports.tokens.created') }}</th>
					<th>{{ $t('user.settings.userReports.tokens.enabled') }}</th>
				</tr>
			</thead>
			<tbody>
				<tr
					v-for="token in tokensByApplication[application.id]"
					:key="token.id"
				>
					<td>{{ token.label }}</td>
					<td>{{ formatDisplayDate(token.created) }}</td>
					<td>
						<label class="checkbox">
							<input
								v-model="token.isEnabled"
								type="checkbox"
								@change="toggleToken(application, token)"
							>
							{{ token.isEnabled ? $t('user.settings.userReports.tokens.statusEnabled') : $t('user.settings.userReports.tokens.statusDisabled') }}
						</label>
					</td>
				</tr>
			</tbody>
		</table>
	</Card>
</template>

<style scoped>
.user-reports-actions {
	display: flex;
	gap: 0.75rem;
	flex-wrap: wrap;
}
</style>
