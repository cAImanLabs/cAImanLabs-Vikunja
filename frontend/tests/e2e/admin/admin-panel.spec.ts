import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {login} from '../../support/authenticateUser'

// The admin panel is a free feature in this fork - only is_admin is enforced, no license required.
test.describe('Admin panel', () => {
	test('an admin user can open /admin and see the overview', async ({page, apiContext}) => {
		const [admin] = await UserFactory.create(1, {is_admin: true}, false)
		await login(page, apiContext, admin)

		await page.goto('/admin')

		await expect(page.locator('.side-nav-shell > nav')).toBeVisible()
		await expect(page.locator('.card-header-title', {hasText: 'Overview'})).toBeVisible()
		await expect(page.locator('.admin-overview__card').first()).toBeVisible()
	})

	test('a non-admin user visiting /admin lands on the not-found page', async ({authenticatedPage: page}) => {
		await page.goto('/admin')
		await expect(page).not.toHaveURL(/\/admin$/)
	})

	test('an admin can navigate to users and projects tabs', async ({page, apiContext}) => {
		const [admin] = await UserFactory.create(1, {is_admin: true}, false)
		await login(page, apiContext, admin)
		await page.goto('/admin')

		const nav = page.locator('.side-nav-shell > nav')
		await nav.getByRole('link', {name: /users/i}).click()
		await expect(page).toHaveURL(/\/admin\/users/)

		await nav.getByRole('link', {name: /projects/i}).click()
		await expect(page).toHaveURL(/\/admin\/projects/)
	})
})
