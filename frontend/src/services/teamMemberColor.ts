import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'

import type {ITeam} from '@/modelTypes/ITeam'
import type {IUser} from '@/modelTypes/IUser'

export function useTeamMemberColorService() {
	const http = AuthenticatedHTTPFactory()

	async function setColor(teamId: ITeam['id'], username: string, color: string): Promise<IUser> {
		const {data} = await http.put(apiV2Url(`teams/${teamId}/members/${username}/color`), {color})
		return data
	}

	return {setColor}
}
