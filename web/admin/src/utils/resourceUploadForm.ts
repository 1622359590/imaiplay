export function createResourceUploadForm(file: File, durationSeconds?: number): FormData {
  const data = new FormData()
  data.append('file', file)
  if (durationSeconds && durationSeconds > 0) {
    data.append('duration_seconds', String(Math.floor(durationSeconds)))
  }
  return data
}
