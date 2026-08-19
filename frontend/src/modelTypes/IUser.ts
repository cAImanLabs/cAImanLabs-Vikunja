import type {IAbstract} from './IAbstract'
import type {IUserSettings} from './IUserSettings'

export const AUTH_TYPES = {
	'UNKNOWN': 0,
	'USER': 1,
	'LINK_SHARE': 2,
} as const

export type AuthType = typeof AUTH_TYPES[keyof typeof AUTH_TYPES]

export interface IUser extends IAbstract {
	id: number
	email: string
	username: string
	name: string
	exp: number
	type: AuthType

	created: Date
	updated: Date
	settings: IUserSettings

	isLocalUser: boolean
	deletionScheduledAt: string | Date | null
	isAdmin?: boolean
	botOwnerId?: number
	// Hex color shown on tasks assigned to this user. Set by a team admin via
	// the team member endpoints; empty/undefined means no color.
	color?: string
}
