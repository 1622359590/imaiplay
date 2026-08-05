export async function loadResourceChart<T>(
  loader: () => Promise<T>,
): Promise<T | null> {
  try {
    return await loader()
  } catch {
    return null
  }
}
