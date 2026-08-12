import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { once } from 'node:events'
import { readFile } from 'node:fs/promises'
import http from 'node:http'
import test from 'node:test'

const verificationScript = new URL('../../scripts/verify-deployment.sh', import.meta.url)

function runVerification(url, retries = '0') {
  return new Promise((resolve) => {
    const child = spawn('sh', [verificationScript.pathname, url], {
      env: {
        ...process.env,
        DEPLOY_HEALTH_RETRIES: retries,
        DEPLOY_HEALTH_RETRY_DELAY_SECONDS: '0',
      },
      stdio: 'pipe',
    })
    let output = ''
    child.stdout.on('data', (chunk) => { output += chunk })
    child.stderr.on('data', (chunk) => { output += chunk })
    child.on('close', (code) => resolve({ code, output }))
  })
}

test('deployment verification supports the older curl shipped by the server', async () => {
  const script = await readFile(verificationScript, 'utf8')
  assert.doesNotMatch(script, /--retry-all-errors/)
})

test('deployment verification rejects an application with a disconnected database', async () => {
  const server = http.createServer((request, response) => {
    if (request.url === '/health') {
      response.writeHead(200).end('{"status":"ok"}')
      return
    }
    response.writeHead(503).end('{"database":"disconnected"}')
  })
  server.listen(0, '127.0.0.1')
  await once(server, 'listening')

  try {
    const address = server.address()
    const result = await runVerification(`http://127.0.0.1:${address.port}/health/db`)
    assert.notEqual(result.code, 0)
    assert.match(result.output, /503/)
  } finally {
    server.close()
    await once(server, 'close')
  }
})

test('deployment verification accepts a connected database', async () => {
  const server = http.createServer((_request, response) => {
    response.writeHead(200).end('{"database":"connected"}')
  })
  server.listen(0, '127.0.0.1')
  await once(server, 'listening')

  try {
    const address = server.address()
    const result = await runVerification(`http://127.0.0.1:${address.port}/health/db`)
    assert.equal(result.code, 0, result.output)
  } finally {
    server.close()
    await once(server, 'close')
  }
})
