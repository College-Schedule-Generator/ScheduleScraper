import path from 'path'
import fs from 'fs'
import puppeteer from 'puppeteer'

const browser = await puppeteer.launch({
	headless: false,
})

const page = await browser.newPage()
const url = 'https://selfservice.elcamino.edu/Student/Courses'
await page.goto(url)
await page.waitForTimeout(500)

const currentTerm = '2022/SU'
// SP: Spring
// SU: Summer
// FA: Fall
await page.select('#term-id', currentTerm)

const [response] = await Promise.all([
	page.waitForNavigation(),
	page.click('#submit-search-form', {
		delay: 23,
	}),
])

let getCurrentPage = () => {
	return page.evaluate(() => {
		const el = document.querySelector('#course-results-next-page')
		if (!el) return

		let ariaLabel = el.getAttribute('aria-label')
		if (!ariaLabel) return

		return parseInt(ariaLabel.slice(ariaLabel.lastIndexOf(' ') + 1), 10) - 1
	})
}

let waitForStuff = async () => {
	await Promise.all([
		page.waitForSelector('#course-search-result'),
		page.waitForSelector('#course-results-last-page'),
		page.waitForSelector('#course-results-total-pages'),
	])
	await page.waitForFunction(() => {
		const el = document.querySelector('.esg-spinner-container') as HTMLElement
		return el?.style?.display === 'none'
	})

	return page.waitForTimeout(3000)
}

await waitForStuff()
let currentPage = await getCurrentPage()
if (currentPage != 1) {
	await page.click('#course-results-first-page')
}

while (true) {
	await waitForStuff()
	let currentPage = await getCurrentPage()
	let totalPages = await page.evaluate(() => {
		const el = document.querySelector('#course-results-total-pages')
		if (!el) return

		return parseInt((el as HTMLElement).innerText, 10)
	})

	console.log(`Writing page ${currentPage} of ${totalPages}`)
	const content = await page.content()
	const cacheDir = path.join('../cache')
	const cacheFile = path.join(
		cacheDir,
		currentTerm.replace('/', '-'),
		`page-${currentPage}.html`
	)
	await fs.promises.mkdir(path.dirname(cacheFile), { recursive: true })
	await fs.promises.writeFile(cacheFile, content)

	if (currentPage === totalPages) {
		break
	}

	await page.click('#course-results-next-page')
}

await await browser.close()
