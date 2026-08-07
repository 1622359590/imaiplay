export interface PaginatedItems<T> {
  items: T[]
  total: number
}

const BACKEND_MAX_PAGE_SIZE = 100

export async function collectPaginatedItems<T>(
  fetchPage: (page: number, pageSize: number) => Promise<PaginatedItems<T>>,
): Promise<T[]> {
  const collected: T[] = []
  let page = 1

  while (true) {
    const result = await fetchPage(page, BACKEND_MAX_PAGE_SIZE)
    collected.push(...result.items)

    if (collected.length >= result.total || result.items.length === 0) {
      return collected
    }
    page += 1
  }
}
