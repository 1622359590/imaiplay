import assert from 'node:assert/strict'
import test from 'node:test'
import { collectPaginatedItems } from '../src/utils/pagination.ts'

test('collects every page without exceeding the backend page-size limit', async () => {
  const requests: Array<{ page: number; pageSize: number }> = []

  const items = await collectPaginatedItems(async (page, pageSize) => {
    requests.push({ page, pageSize })
    const allItems = Array.from({ length: 205 }, (_, index) => `learner-${index + 1}`)
    const offset = (page - 1) * pageSize
    return {
      items: allItems.slice(offset, offset + pageSize),
      total: allItems.length,
    }
  })

  assert.deepEqual(requests, [
    { page: 1, pageSize: 100 },
    { page: 2, pageSize: 100 },
    { page: 3, pageSize: 100 },
  ])
  assert.equal(items.length, 205)
  assert.equal(items.at(-1), 'learner-205')
})
