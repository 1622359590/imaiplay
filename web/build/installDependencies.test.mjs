import assert from 'node:assert/strict'
import { chmod, mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { spawn } from 'node:child_process'
import test from 'node:test'

const installerPath = new URL('./installDependencies.mjs', import.meta.url)

function runInstaller(cwd, env) {
  return new Promise((resolve) => {
    const child = spawn(process.execPath, [installerPath.pathname], {
      cwd,
      env: { ...process.env, ...env },
      stdio: 'pipe',
    })
    let output = ''
    child.stdout.on('data', (chunk) => { output += chunk })
    child.stderr.on('data', (chunk) => { output += chunk })
    child.on('close', (code) => resolve({ code, output }))
  })
}

test('retries a successful npm exit when TypeScript was not installed', async () => {
  const root = await mkdtemp(path.join(tmpdir(), 'imaiplay-install-'))
  try {
    const bin = path.join(root, 'bin')
    await mkdir(bin)
    const fakeNpm = path.join(bin, 'npm')
    await writeFile(fakeNpm, `#!/bin/sh
set -eu
attempt_file="$PWD/attempts"
attempt=0
[ ! -f "$attempt_file" ] || attempt=$(cat "$attempt_file")
attempt=$((attempt + 1))
printf '%s' "$attempt" > "$attempt_file"
printf '%s' "$*" > "$PWD/npm-args"
if [ "$attempt" -eq 2 ]; then
  mkdir -p node_modules/.bin
  printf '#!/bin/sh\\nexit 0\\n' > node_modules/.bin/tsc
  chmod +x node_modules/.bin/tsc
fi
exit 0
`)
    await chmod(fakeNpm, 0o755)

    const result = await runInstaller(root, {
      PATH: `${bin}:${process.env.PATH}`,
      IMAIPLAY_NPM_RETRY_DELAY_MS: '0',
    })

    assert.equal(result.code, 0, result.output)
    assert.equal(await readFile(path.join(root, 'attempts'), 'utf8'), '2')
    assert.match(
      await readFile(path.join(root, 'npm-args'), 'utf8'),
      /--registry=https:\/\/registry\.npmmirror\.com/,
    )
  } finally {
    await rm(root, { recursive: true, force: true })
  }
})

test('fails after three incomplete dependency installations', async () => {
  const root = await mkdtemp(path.join(tmpdir(), 'imaiplay-install-'))
  try {
    const bin = path.join(root, 'bin')
    await mkdir(bin)
    const fakeNpm = path.join(bin, 'npm')
    await writeFile(fakeNpm, `#!/bin/sh
set -eu
attempt_file="$PWD/attempts"
attempt=0
[ ! -f "$attempt_file" ] || attempt=$(cat "$attempt_file")
attempt=$((attempt + 1))
printf '%s' "$attempt" > "$attempt_file"
exit 0
`)
    await chmod(fakeNpm, 0o755)

    const result = await runInstaller(root, {
      PATH: `${bin}:${process.env.PATH}`,
      IMAIPLAY_NPM_RETRY_DELAY_MS: '0',
    })

    assert.notEqual(result.code, 0)
    assert.equal(await readFile(path.join(root, 'attempts'), 'utf8'), '3')
    assert.match(result.output, /TypeScript executable is missing/)
  } finally {
    await rm(root, { recursive: true, force: true })
  }
})
