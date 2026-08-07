export interface UserImportError {
  row: number
  name: string
  email: string
  phone: string
  role: string
  reason: string
}

export interface UserImportResult {
  total: number
  succeeded: number
  failed: number
  errors: UserImportError[]
}

export interface UserImportSummary {
  status: 'success' | 'warning' | 'error'
  title: string
}

interface ImportFile {
  name: string
  size: number
}

const UTF8_BOM = '\uFEFF'
const IMPORT_HEADER = ['姓名', '邮箱', '手机号（可选）', '角色（可选）', '初始密码']
const ERROR_HEADER = ['行号', '姓名', '邮箱', '手机号', '角色', '失败原因']
const MAX_IMPORT_FILE_BYTES = 10 * 1024 * 1024

export function validateUserImportFile(file: ImportFile): string | undefined {
  if (!/\.(csv|xlsx)$/i.test(file.name)) return '仅支持 CSV 或 XLSX 文件'
  if (file.size > MAX_IMPORT_FILE_BYTES) return '导入文件不能超过 10MB'
  return undefined
}

export function createUserImportFormData(file: Blob): FormData {
  const form = new FormData()
  form.set('file', file)
  return form
}

export function userImportTemplateCSV(): string {
  return UTF8_BOM + [
    IMPORT_HEADER,
    ['示例学员', 'learner@example.com', '', '学员', 'password123'],
  ].map(csvRow).join('\r\n')
}

export function userImportErrorsCSV(errors: UserImportError[]): string {
  const rows: Array<Array<string | number>> = [
    ERROR_HEADER,
    ...errors.map((error) => [
      error.row,
      error.name,
      error.email,
      error.phone,
      error.role,
      error.reason,
    ]),
  ]
  return UTF8_BOM + rows.map(csvRow).join('\r\n')
}

export function importResultSummary(result: UserImportResult): UserImportSummary {
  if (result.failed === 0) {
    return { status: 'success', title: `成功导入 ${result.succeeded} 位成员` }
  }
  if (result.succeeded === 0) {
    return { status: 'error', title: `导入失败 ${result.failed} 条` }
  }
  return { status: 'warning', title: `成功 ${result.succeeded} 条，失败 ${result.failed} 条` }
}

export function downloadUserImportCSV(contents: string, filename: string): void {
  const url = URL.createObjectURL(new Blob([contents], { type: 'text/csv;charset=utf-8' }))
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  URL.revokeObjectURL(url)
}

function csvRow(values: Array<string | number>): string {
  return values.map((value) => csvCell(String(value))).join(',')
}

function csvCell(value: string): string {
  const safe = /^[=+\-@\t\r\n]/.test(value) ? `'${value}` : value
  if (!/[",\r\n]/.test(safe)) return safe
  return `"${safe.replace(/"/g, '""')}"`
}
