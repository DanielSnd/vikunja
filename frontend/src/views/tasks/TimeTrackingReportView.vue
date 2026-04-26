<template>
	<div class="time-tracking-report">
		<div class="report-header">
			<div>
				<h1>{{ $t('timeTrackingReport.title') }}</h1>
				<p>{{ $t('timeTrackingReport.description') }}</p>
			</div>
			<BaseButton
				:loading="isLoading"
				icon="sync"
				@click="loadReport"
			>
				{{ $t('misc.refresh') }}
			</BaseButton>
		</div>

		<div class="report-filters">
			<div class="field">
				<label
					class="label"
					for="time-tracking-group-by"
				>
					{{ $t('timeTrackingReport.filters.groupBy') }}
				</label>
				<div class="select is-fullwidth">
					<select
						id="time-tracking-group-by"
						v-model="groupBy"
					>
						<option value="project">
							{{ $t('timeTrackingReport.groupBy.project') }}
						</option>
						<option value="label">
							{{ $t('timeTrackingReport.groupBy.label') }}
						</option>
					</select>
				</div>
			</div>

			<div class="field">
				<label class="label">
					{{ $t('timeTrackingReport.filters.period') }}
				</label>
				<DatepickerWithRange v-model="dateRange">
					<template #trigger="{toggle, buttonText}">
						<BaseButton
							class="is-fullwidth report-range-button"
							icon="calendar"
							@click="toggle"
						>
							{{ buttonText }}
						</BaseButton>
					</template>
				</DatepickerWithRange>
			</div>

			<div class="field">
				<label class="label">
					{{ $t('timeTrackingReport.filters.projects') }}
				</label>
				<Multiselect
					v-model="selectedProjects"
					:search-results="foundProjects"
					:placeholder="$t('timeTrackingReport.filters.projectsPlaceholder')"
					label="title"
					:multiple="true"
					:show-empty="true"
					:close-after-select="false"
					@search="findProjects"
				/>
			</div>

			<div class="field">
				<label class="label">
					{{ $t('timeTrackingReport.filters.labels') }}
				</label>
				<Multiselect
					v-model="selectedLabels"
					:search-results="foundLabels"
					:placeholder="$t('timeTrackingReport.filters.labelsPlaceholder')"
					label="title"
					:multiple="true"
					:show-empty="true"
					:close-after-select="false"
					@search="findLabels"
				>
					<template #searchResult="{option}">
						<span
							:style="getLabelStyles(option)"
							class="tag search-result"
						>
							<span>{{ option.title }}</span>
						</span>
					</template>
				</Multiselect>
			</div>
		</div>

		<Loading
			v-if="isLoading"
			class="mt-6"
		/>

		<template v-else-if="report !== null">
			<div class="report-summary">
				<div class="summary-card">
					<span class="summary-label">{{ $t('timeTrackingReport.summary.total') }}</span>
					<strong>{{ formatDuration(report.totalSeconds) }}</strong>
				</div>
				<div class="summary-card">
					<span class="summary-label">{{ $t('timeTrackingReport.summary.groups') }}</span>
					<strong>{{ report.items.length }}</strong>
				</div>
				<div class="summary-card">
					<span class="summary-label">{{ averageLabel }}</span>
					<strong>{{ formatDuration(averageBucketSeconds) }}</strong>
				</div>
			</div>

			<Message
				v-if="report.items.length === 0"
				class="mt-4"
			>
				{{ $t('timeTrackingReport.empty') }}
			</Message>

			<template v-else>
				<div class="report-chart-card">
					<div class="report-chart-header">
						<h2>{{ trendTitle }}</h2>
						<div class="report-legend">
							<span
								v-for="(item, index) in report.items"
								:key="item.id"
								class="legend-item"
							>
								<span
									class="legend-swatch"
									:style="{backgroundColor: getSeriesColor(item, index)}"
								/>
								{{ item.title }}
							</span>
						</div>
					</div>

					<div class="report-chart">
						<div
							v-for="bucket in chartBuckets"
							:key="bucket.key"
							class="chart-day"
						>
							<div class="chart-total">
								{{ formatDuration(bucket.totalSeconds) }}
							</div>
							<div class="chart-bar">
								<div
									v-for="(item, itemIndex) in report.items"
									:key="`${item.id}-${bucket.key}`"
									class="chart-segment"
									:style="{
										height: `${getSegmentHeight(bucket.itemSeconds[itemIndex] ?? 0)}%`,
										backgroundColor: getSeriesColor(item, itemIndex),
									}"
								/>
							</div>
							<div class="chart-label">
								{{ bucket.label }}
							</div>
						</div>
					</div>
				</div>

				<div class="report-table-card">
					<table class="table is-fullwidth">
						<thead>
							<tr>
								<th>{{ groupByLabel }}</th>
								<th>{{ $t('timeTrackingReport.table.total') }}</th>
								<th>{{ averageLabel }}</th>
								<th>{{ $t('timeTrackingReport.table.share') }}</th>
							</tr>
						</thead>
						<tbody>
							<tr
								v-for="(item, index) in report.items"
								:key="item.id"
							>
								<td>
									<div class="item-cell">
										<span
											class="item-dot"
											:style="{backgroundColor: getSeriesColor(item, index)}"
										/>
										{{ item.title }}
									</div>
								</td>
								<td>{{ formatDuration(item.totalSeconds) }}</td>
								<td>{{ formatDuration(Math.round(item.totalSeconds / Math.max(1, chartBuckets.length))) }}</td>
								<td>{{ formatShare(item.totalSeconds) }}</td>
							</tr>
						</tbody>
					</table>
				</div>
			</template>
		</template>
	</div>
</template>

<script setup lang="ts">
import {computed, onMounted, ref, shallowReactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import dayjs from 'dayjs'

import BaseButton from '@/components/base/BaseButton.vue'
import Loading from '@/components/misc/Loading.vue'
import Message from '@/components/misc/Message.vue'
import DatepickerWithRange from '@/components/date/DatepickerWithRange.vue'
import Multiselect from '@/components/input/Multiselect.vue'

import TaskTimeTrackingReportService from '@/services/taskTimeTrackingReport'
import type {ITaskTimeTrackingReport, ITaskTimeTrackingReportItem} from '@/modelTypes/ITaskTimeTrackingReport'
import type {IProject} from '@/modelTypes/IProject'
import type {ILabel} from '@/modelTypes/ILabel'
import {formatDuration} from '@/helpers/time/formatDuration'
import {formatDate} from '@/helpers/time/formatDate'
import {useProjectStore} from '@/stores/projects'
import {useLabelStore} from '@/stores/labels'
import {useLabelStyles} from '@/composables/useLabelStyles'

const {t} = useI18n({useScope: 'global'})

const projectStore = useProjectStore()
const labelStore = useLabelStore()
const reportService = shallowReactive(new TaskTimeTrackingReportService())
const {getLabelStyles} = useLabelStyles()

const report = ref<ITaskTimeTrackingReport | null>(null)
const groupBy = ref<'project' | 'label'>('project')
const selectedProjects = ref<IProject[]>([])
const selectedLabels = ref<ILabel[]>([])
const foundProjects = ref<IProject[]>([])
const foundLabels = ref<ILabel[]>([])
const dateRange = ref({
	dateFrom: new Date(Date.now() - (6 * 24 * 60 * 60 * 1000)).toISOString(),
	dateTo: new Date().toISOString(),
})

const palette = ['#205072', '#329d9c', '#56c596', '#7be495', '#f2c14e', '#ef476f', '#8d5fd3', '#5c7cfa']
type TrendBucketMode = 'daily' | 'weekly' | 'monthly'
type ChartBucket = {
	key: string
	label: string
	totalSeconds: number
	itemSeconds: number[]
}

const isLoading = computed(() => reportService.loading)
const bucketMode = computed<TrendBucketMode>(() => {
	const dayCount = report.value?.days.length ?? 0
	if (dayCount > 62) {
		return 'monthly'
	}
	if (dayCount > 14) {
		return 'weekly'
	}
	return 'daily'
})
const chartBuckets = computed<ChartBucket[]>(() => {
	if (report.value === null) {
		return []
	}

	if (bucketMode.value === 'daily') {
		return report.value.days.map((day, dayIndex) => ({
			key: day.date,
			label: formatDay(day.date),
			totalSeconds: day.totalSeconds,
			itemSeconds: report.value?.items.map(item => item.dailySeconds[dayIndex] ?? 0) ?? [],
		}))
	}

	const buckets = new Map<string, ChartBucket>()
	const reportEndDate = dayjs(report.value.dateTo)
	report.value.days.forEach((day, dayIndex) => {
		const date = dayjs(day.date)
		const bucketStart = bucketMode.value === 'weekly'
			? date.startOf('week')
			: date.startOf('month')
		const key = bucketMode.value === 'weekly'
			? bucketStart.format('YYYY-MM-DD')
			: bucketStart.format('YYYY-MM')

		if (!buckets.has(key)) {
			const naturalBucketEnd = bucketMode.value === 'weekly'
				? bucketStart.endOf('week')
				: bucketStart.endOf('month')
			const bucketEnd = naturalBucketEnd.isBefore(reportEndDate)
				? naturalBucketEnd
				: reportEndDate
			buckets.set(key, {
				key,
				label: bucketMode.value === 'weekly'
					? `${bucketStart.format('MMM D')} - ${bucketEnd.format(bucketStart.month() === bucketEnd.month() ? 'D' : 'MMM D')}`
					: bucketStart.format('MMM YYYY'),
				totalSeconds: 0,
				itemSeconds: new Array(report.value.items.length).fill(0),
			})
		}

		const bucket = buckets.get(key)!
		bucket.totalSeconds += day.totalSeconds
		report.value.items.forEach((item, itemIndex) => {
			bucket.itemSeconds[itemIndex] += item.dailySeconds[dayIndex] ?? 0
		})
	})

	return Array.from(buckets.values())
})
const averageBucketSeconds = computed(() => {
	if (report.value === null || chartBuckets.value.length === 0) {
		return 0
	}

	return Math.round(report.value.totalSeconds / chartBuckets.value.length)
})
const maxBucketSeconds = computed(() => Math.max(...chartBuckets.value.map(bucket => bucket.totalSeconds), 1))
const groupByLabel = computed(() => groupBy.value === 'project'
	? t('timeTrackingReport.groupBy.project')
	: t('timeTrackingReport.groupBy.label'),
)
const averageLabel = computed(() => t(`timeTrackingReport.summary.average.${bucketMode.value}`))
const trendTitle = computed(() => t(`timeTrackingReport.chart.title.${bucketMode.value}`))

function findProjects(query: string) {
	foundProjects.value = query === ''
		? projectStore.projectsArray.filter(project => project.id > 0 && !project.isArchived)
		: projectStore.searchProject(query)
}

function findLabels(query: string) {
	foundLabels.value = labelStore.filterLabelsByQuery(selectedLabels.value, query)
}

function getSeriesColor(item: ITaskTimeTrackingReportItem, index: number) {
	if (item.color !== '') {
		return item.color.startsWith('#') ? item.color : `#${item.color}`
	}

	return palette[index % palette.length]
}

function getSegmentHeight(seconds: number) {
	if (seconds <= 0) {
		return 0
	}

	return (seconds / maxBucketSeconds.value) * 100
}

function formatDay(date: string) {
	return formatDate(date, 'ddd D')
}

function formatShare(seconds: number) {
	if (report.value === null || report.value.totalSeconds === 0) {
		return '0%'
	}

	return `${Math.round((seconds / report.value.totalSeconds) * 100)}%`
}

async function loadReport() {
	report.value = await reportService.getReport({
		dateFrom: String(dateRange.value.dateFrom),
		dateTo: String(dateRange.value.dateTo),
		groupBy: groupBy.value,
		projectIds: selectedProjects.value.map(project => project.id),
		labelIds: selectedLabels.value.map(label => label.id),
	})
}

watch([groupBy, selectedProjects, selectedLabels, dateRange], loadReport, {deep: true})

onMounted(async () => {
	await Promise.all([
		projectStore.loadAllProjects(),
		labelStore.loadAllLabels(),
	])

	foundProjects.value = projectStore.projectsArray.filter(project => project.id > 0 && !project.isArchived)
	foundLabels.value = labelStore.labelsArray
	await loadReport()
})
</script>

<style lang="scss" scoped>
.time-tracking-report {
	display: flex;
	flex-direction: column;
	gap: 1.25rem;
}

.report-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 1rem;

	h1 {
		margin-block-end: .35rem;
	}

	p {
		color: var(--grey-700);
		margin: 0;
		max-inline-size: 52rem;
	}
}

.report-filters,
.report-chart-card,
.report-table-card {
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	padding: 1.25rem;
}

.report-filters {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
	gap: 1rem;
}

.report-range-button {
	justify-content: space-between;
}

.report-summary {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
	gap: 1rem;
}

.summary-card {
	background: linear-gradient(145deg, var(--grey-050), var(--white));
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	padding: 1rem 1.1rem;
	display: flex;
	flex-direction: column;
	gap: .35rem;

	strong {
		font-size: 1.35rem;
	}
}

.summary-label {
	color: var(--grey-700);
	font-size: .9rem;
	text-transform: uppercase;
	letter-spacing: .03em;
}

.report-chart-header {
	display: flex;
	flex-direction: column;
	gap: 1rem;
	margin-block-end: 1.25rem;

	h2 {
		margin: 0;
	}
}

.report-legend {
	display: flex;
	flex-wrap: wrap;
	gap: .75rem 1rem;
}

.legend-item,
.item-cell {
	display: inline-flex;
	align-items: center;
	gap: .55rem;
}

.legend-swatch,
.item-dot {
	inline-size: .75rem;
	block-size: .75rem;
	border-radius: 9999px;
	flex-shrink: 0;
}

.report-chart {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(72px, 1fr));
	gap: 1rem;
	align-items: end;
	min-block-size: 19rem;
}

.chart-day {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: .6rem;
	min-inline-size: 0;
}

.chart-total,
.chart-label {
	font-size: .85rem;
	text-align: center;
}

.chart-total {
	color: var(--grey-700);
}

.chart-bar {
	inline-size: 100%;
	max-inline-size: 3rem;
	block-size: 14rem;
	display: flex;
	flex-direction: column-reverse;
	justify-content: flex-start;
	gap: 2px;
	padding: .35rem;
	border-radius: calc(#{$radius} * 1.5);
	background:
		linear-gradient(180deg, rgba(32, 80, 114, .06), rgba(32, 80, 114, .14)),
		repeating-linear-gradient(
			to top,
			transparent,
			transparent 23%,
			rgba(32, 80, 114, .08) 23%,
			rgba(32, 80, 114, .08) 25%
		);
}

.chart-segment {
	inline-size: 100%;
	border-radius: 9999px;
	min-block-size: 0;
}

.tag {
	margin: .25rem !important;
}

.tag.search-result {
	margin: 0 !important;
}

@media screen and (max-width: $tablet) {
	.report-header {
		flex-direction: column;
	}

	.report-chart {
		grid-template-columns: repeat(auto-fit, minmax(58px, 1fr));
		gap: .75rem;
	}

	.report-chart-card,
	.report-table-card,
	.report-filters {
		padding: 1rem;
	}
}
</style>
