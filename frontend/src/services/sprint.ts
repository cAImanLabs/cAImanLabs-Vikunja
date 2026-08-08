import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'
import {objectToCamelCase, objectToSnakeCase} from '@/helpers/case'

import type {ISprint, ISprintFormData, ISprintVelocityPoint, ISprintBurndown} from '@/modelTypes/ISprint'

// Go's zero time.Time serializes as 0001-01-01T00:00:00Z when a date was never set.
function parseOptionalDate(value: unknown): Date | null {
	if (typeof value !== 'string' || value === '') {
		return null
	}
	const date = new Date(value)
	return date.getFullYear() <= 1 ? null : date
}

export function parseSprint(raw: Record<string, unknown>): ISprint {
	const s = objectToCamelCase(raw)
	return {
		id: s.id,
		title: s.title,
		goal: s.goal ?? '',
		startDate: parseOptionalDate(s.startDate),
		endDate: parseOptionalDate(s.endDate),
		status: s.status,
		projectId: s.projectId,
		createdBy: s.createdBy ?? null,
		created: new Date(s.created),
		updated: new Date(s.updated),
		maxPermission: s.maxPermission ?? null,
	}
}

export function useSprintService() {
	const http = AuthenticatedHTTPFactory()

	async function getAll(projectId: number): Promise<ISprint[]> {
		const {data} = await http.get(apiV2Url(`projects/${projectId}/sprints`), {
			params: {per_page: 250},
		})
		return (data.items ?? []).map(parseSprint)
	}

	async function create(projectId: number, sprint: ISprintFormData): Promise<ISprint> {
		const {data} = await http.post(apiV2Url(`projects/${projectId}/sprints`), objectToSnakeCase(sprint))
		return parseSprint(data)
	}

	async function update(projectId: number, sprint: ISprintFormData & {id: number}): Promise<ISprint> {
		const {data} = await http.put(apiV2Url(`projects/${projectId}/sprints/${sprint.id}`), objectToSnakeCase(sprint))
		return parseSprint(data)
	}

	async function remove(projectId: number, id: number): Promise<void> {
		await http.delete(apiV2Url(`projects/${projectId}/sprints/${id}`))
	}

	async function getVelocity(projectId: number): Promise<ISprintVelocityPoint[]> {
		const {data} = await http.get(apiV2Url(`projects/${projectId}/sprints/velocity`))
		return (data ?? []).map((raw: Record<string, unknown>) => objectToCamelCase(raw) as ISprintVelocityPoint)
	}

	async function getBurndown(projectId: number, sprintId: number): Promise<ISprintBurndown> {
		const {data} = await http.get(apiV2Url(`projects/${projectId}/sprints/${sprintId}/burndown`))
		const b = objectToCamelCase(data)
		return {
			sprintId: b.sprintId,
			totalPoints: b.totalPoints,
			startDate: new Date(b.startDate),
			endDate: new Date(b.endDate),
			series: (b.series ?? []).map((p: Record<string, unknown>) => ({
				date: new Date(p.date as string),
				remaining: p.remaining,
			})),
		}
	}

	return {getAll, create, update, remove, getVelocity, getBurndown}
}
