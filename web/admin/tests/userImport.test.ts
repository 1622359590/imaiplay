import assert from 'node:assert/strict'
import test from 'node:test'
import {
  createUserImportFormData,
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
