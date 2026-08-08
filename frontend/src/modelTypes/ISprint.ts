import type {IAbstract} from './IAbstract'
import type {IProject} from '@/modelTypes/IProject'
import type {IUser} from '@/modelTypes/IUser'

export const SPRINT_STATUSES = {
	PLANNING: 'planning',
	ACTIVE: 'active',
	COMPLETED: 'completed',
} as const
export type SprintStatus = typeof SPRINT_STATUSES[keyof typeof SPRINT_STATUSES]

export interface ISprint extends IAbstract {
	id: number
	title: string
	goal: string
	startDate: Date | null
	endDate: Date | null
	status: SprintStatus
	projectId: IProject['id']

	createdBy: IUser | null

	created: Date
	updated: Date
}

// The editable subset of a sprint, used by the create/edit form. Unlike
// Partial<ISprint>, fields here are never undefined so form controls (e.g.
// Datepicker's Date | null modelValue) don't have to special-case it.
export interface ISprintFormData {
	title: string
	goal: string
	startDate: Date | null
	endDate: Date | null
	status: SprintStatus
}

// One sprint's total vs completed story points, for plotting velocity across a project's sprints.
export interface ISprintVelocityPoint {
	sprintId: number
	title: string
	totalPoints: number
	completedPoints: number
}

// The remaining story points as of one day of a sprint's burndown.
export interface ISprintBurndownPoint {
	date: Date
	remaining: number
}

export interface ISprintBurndown {
	sprintId: number
	totalPoints: number
	startDate: Date
	endDate: Date
	series: ISprintBurndownPoint[]
}
