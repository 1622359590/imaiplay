import assert from 'node:assert/strict'
import test from 'node:test'
import {
  responseMessage,
  responseStatus,
  userFacingErrorMessage,
} from '../src/api/errors.ts'

test('reads an Axios-compatible response', () => {
  const error = { response: { status: 403, data: { message: '租户已暂停' } } }
  assert.equal(responseStatus(error), 403)
  assert.equal(responseMessage(error), '租户已暂停')
})

test('does not invent fields for unknown errors', () => {
  assert.equal(responseStatus(new Error('network')), undefined)
  assert.equal(responseMessage(new Error('network')), undefined)
})

test('preserves readable Chinese business messages', () => {
  const error = { response: { status: 400, data: { message: '验证码已失效，请重新获取' } } }
  assert.equal(userFacingErrorMessage(error), '验证码已失效，请重新获取')
})

test('translates known backend business messages', () => {
  const cases = [
    ['invalid email or password', '邮箱或密码错误'],
    ['permission denied', '您没有权限执行此操作'],
    ['resource is in use', '该资源正在使用中，暂时无法删除'],
  ] as const

  for (const [message, expected] of cases) {
    const error = { response: { status: 400, data: { message } } }
    assert.equal(userFacingErrorMessage(error), expected)
  }
})

test('translates HTTP statuses when no readable business message exists', () => {
  const cases = [
    [401, '登录信息已失效，请重新登录'],
    [403, '您没有权限执行此操作'],
    [404, '请求的内容不存在'],
    [409, '数据已存在或当前状态发生冲突'],
    [413, '上传文件过大，请选择较小的文件'],
    [429, '操作过于频繁，请稍后再试'],
    [500, '服务器暂时异常，请稍后重试'],
  ] as const

  for (const [status, expected] of cases) {
    assert.equal(userFacingErrorMessage({ response: { status, data: {} } }), expected)
  }
})

test('translates network failures and timeouts', () => {
  assert.equal(userFacingErrorMessage({ code: 'ERR_NETWORK', message: 'Network Error' }), '网络连接失败，请检查网络后重试')
  assert.equal(userFacingErrorMessage({ code: 'ECONNABORTED', message: 'timeout of 10000ms exceeded' }), '请求超时，请稍后重试')
  assert.equal(userFacingErrorMessage(new Error('Network Error')), '网络连接失败，请检查网络后重试')
})

test('uses the supplied Chinese fallback for unknown technical errors', () => {
  assert.equal(userFacingErrorMessage(new Error('Request failed with status code 418'), '登录失败，请稍后重试'), '登录失败，请稍后重试')
})
