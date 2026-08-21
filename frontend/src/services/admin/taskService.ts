import AbstractService from '@/services/abstractService'
import AdminTaskModel from '@/models/adminTask'
import type {IAdminTask} from '@/modelTypes/IAdminTask'
import type {IUser} from '@/modelTypes/IUser'
import type {IUserTaskCompletionStat} from '@/modelTypes/IUserTaskCompletionStat'
import {objectToCamelCase} from '@/helpers/case'

export default class AdminTaskService extends AbstractService<IAdminTask> {
	constructor() {
		super({
			getAll: '/admin/tasks',
		})
	}

	modelFactory(data: Partial<IAdminTask>) {
		return new AdminTaskModel(data)
	}

	async reassignCompletedBy(taskId: IAdminTask['id'], completedById: IUser['id']) {
		const {data} = await this.http.patch(`/admin/tasks/${taskId}/completed-by`, {completed_by_id: completedById})
		return this.modelUpdateFactory(data)
	}

	async getCompletionStats(): Promise<IUserTaskCompletionStat[]> {
		const {data} = await this.http.get('/admin/tasks/completion-stats')
		return (data as Record<string, unknown>[]).map(row => objectToCamelCase(row) as IUserTaskCompletionStat)
	}
}
