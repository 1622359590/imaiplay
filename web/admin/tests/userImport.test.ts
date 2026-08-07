import assert from 'node:assert/strict'
import test from 'node:test'
import {
  createUserImportFormData,
  importResultSummary,
  userImportErrorsCSV,
  userImportTemplateCSV,
  validateUserImportFile,
} from '../src/utils/userImport.ts'

test('user import accepts CSV and XLSX files only', () => {
  assert.equal(validateUserImportFile({ name: 'users.csv', size: 100 }), undefined)
  assert.equal(validateUserImportFile({ name: 'users.XLSX', size: 100 }), undefined)
  assert.equal(validateUserImportFile({ name: 'users.xls', size: 100 }), '仅支持 CSV 或 XLSX 文件')
  assert.equal(validateUserImportFile({ name: 'users.txt', size: 100 }), '仅支持 CSV 或 XLSX 文件')
})

test('user import request uses the multipart file field expected by the API', async () => {
  const file = new File(['users'], 'users.csv', { type: 'text/csv' })
  const form = createUserImportFormData(file)
  const uploaded = form.get('file')
  assert.ok(uploaded instanceof File)
  assert.equal(uploaded.name, 'users.csv')
  assert.equal(await uploaded.text(), 'users')
})

test('user import template is an Excel-friendly UTF-8 CSV with an example row', () => {
  const csv = userImportTemplateCSV()
  assert.ok(csv.startsWith('\uFEFF姓名,邮箱,手机号（可选）,角色（可选）,初始密码\r\n'))
  assert.ok(csv.includes('示例学员,learner@example.com,,学员,password123'))
})

test('user import error CSV escapes cells and never contains a password column', () => {
  const csv = userImportErrorsCSV([
    {
      row: 3,
      name: '张,三',
      email: 'a@example.com',
      phone: '',
      role: 'learner',
      reason: '邮箱"重复\n请修改',
    },
  ])

  assert.equal(
    csv,
    '\uFEFF行号,姓名,邮箱,手机号,角色,失败原因\r\n3,"张,三",a@example.com,,learner,"邮箱""重复\n请修改"',
  )
  assert.equal(csv.includes('password'), false)
  assert.equal(csv.includes('初始密码'), false)
})

test('user import error CSV neutralizes spreadsheet formulas from uploaded values', () => {
  const csv = userImportErrorsCSV([{
    row: 2,
    name: '=2+2',
    email: 'safe@example.com',
    phone: '+13800138000',
    role: '@learner',
    reason: '格式错误',
  }])

  assert.ok(csv.includes("'=2+2"))
  assert.ok(csv.includes("'+13800138000"))
  assert.ok(csv.includes("'@learner"))
})

test('user import result distinguishes full, partial, and failed outcomes', () => {
  assert.deepEqual(importResultSummary({ total: 2, succeeded: 2, failed: 0, errors: [] }), {
    status: 'success',
    title: '成功导入 2 位成员',
  })
  assert.deepEqual(importResultSummary({ total: 3, succeeded: 2, failed: 1, errors: [] }), {
    status: 'warning',
    title: '成功 2 条，失败 1 条',
  })
  assert.deepEqual(importResultSummary({ total: 2, succeeded: 0, failed: 2, errors: [] }), {
    status: 'error',
    title: '导入失败 2 条',
  })
})
