import { constants } from 'node:fs'
import { access, rm } from 'node:fs/promises'
import { spawn } from 'node:child_process'
import path from 'node:path'

const installDirectories = [
  'node_modules',
  'shared/node_modules',
  'admin/node_modules',
  'pc/node_modules',
  'h5/node_modules',
]

function runNpmInstall() {
  const registry = process.env.NPM_CONFIG_REGISTRY?.trim()
    || 'https://registry.npmmirror.com'
  return new Promise((resolve) => {
    const child = spawn(
      'npm',
      ['ci', '--include=dev', `--registry=${registry}`],
      { stdio: 'inherit' },
    )
    child.on('error', () => resolve(1))
    child.on('close', (code) => resolve(code ?? 1))
  })
}

async function hasTypeScriptExecutable() {
  try {
    await access(path.join('node_modules', '.bin', 'tsc'), constants.X_OK)
    return true
  } catch {
    return false
  }
}

async function resetIncompleteInstall() {
  await Promise.all(installDirectories.map((directory) => (
    rm(directory, { recursive: true, force: true })
  )))
}

function wait(milliseconds) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds))
}

async function installDependencies() {
  const maxAttempts = 3
  const retryDelay = Number.parseInt(
    process.env.IMAIPLAY_NPM_RETRY_DELAY_MS ?? '5000',
    10,
  )

  for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
    await resetIncompleteInstall()
    const exitCode = await runNpmInstall()
    if (exitCode === 0 && await hasTypeScriptExecutable()) {
      return
    }

    const reason = exitCode === 0
      ? 'TypeScript executable is missing after npm ci'
      : `npm ci exited with code ${exitCode}`
    console.error(`Dependency installation attempt ${attempt}/${maxAttempts} failed: ${reason}`)

    if (attempt < maxAttempts && retryDelay > 0) {
      await wait(retryDelay)
    }
  }

  throw new Error('TypeScript executable is missing after dependency installation retries')
}

installDependencies().catch((error) => {
  console.error(error instanceof Error ? error.message : error)
  process.exitCode = 1
})
