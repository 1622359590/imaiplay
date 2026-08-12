import assert from 'node:assert/strict'
import { chmod, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { spawn } from 'node:child_process'
import test from 'node:test'

const workflowUrl = new URL('../../.github/workflows/deploy.yml', import.meta.url)
const composeOverrideUrl = new URL('../../docker-compose.deploy.yml', import.meta.url)
const packageUrl = new URL('../package.json', import.meta.url)
const lockfileUrl = new URL('../package-lock.json', import.meta.url)
const deployScriptUrl = new URL('../../scripts/deploy-release.sh', import.meta.url)

function run(command, args, options = {}) {
  return new Promise((resolve) => {
    const child = spawn(command, args, options)
    let output = ''
    child.stdout?.on('data', (chunk) => { output += chunk })
    child.stderr?.on('data', (chunk) => { output += chunk })
    child.on('close', (code) => resolve({ code, output }))
  })
}

test('deployment builds the release image on the GitHub runner', async () => {
  const workflow = await readFile(workflowUrl, 'utf8')

  assert.match(workflow, /uses: actions\/checkout@v4/)
  assert.match(workflow, /docker build --tag "\$release_image" \./)
  assert.doesNotMatch(workflow, /docker compose[^\n]*up[^\n]*--build/)
})

test('deployment streams the immutable image and starts it without rebuilding', async () => {
  const [workflow, composeOverride, deployScript] = await Promise.all([
    readFile(workflowUrl, 'utf8'),
    readFile(composeOverrideUrl, 'utf8'),
    readFile(deployScriptUrl, 'utf8'),
  ])

  assert.match(workflow, /docker save "\$release_image" \| gzip -1 \| ssh/)
  assert.match(workflow, /docker load/)
  assert.match(workflow, /sh scripts\/deploy-release\.sh \$release_image_quoted/)
  assert.match(composeOverride, /image: \$\{IMAI_PLAY_IMAGE:\?IMAI_PLAY_IMAGE is required\}/)
  assert.match(deployScript, /docker-compose\.deploy\.yml/)
  assert.match(deployScript, /up -d --no-build --remove-orphans/)
})

test('release deployment repairs a missing Docker iptables chain once', async () => {
  const fixture = await mkdtemp(join(tmpdir(), 'imaiplay-deploy-'))
  const bin = join(fixture, 'bin')
  const calls = join(fixture, 'calls.log')
  await import('node:fs/promises').then(({ mkdir }) => mkdir(bin))
  await writeFile(join(bin, 'docker'), `#!/bin/sh
echo "docker $*" >> "$CALLS_FILE"
if [ "$1" = compose ] && [ ! -f "$DEPLOY_STATE" ]; then
  touch "$DEPLOY_STATE"
  echo 'iptables: No chain/target/match by that name.' >&2
  exit 1
fi
exit 0
`)
  await writeFile(join(bin, 'systemctl'), `#!/bin/sh
echo "systemctl $*" >> "$CALLS_FILE"
exit 0
`)
  await Promise.all([chmod(join(bin, 'docker'), 0o755), chmod(join(bin, 'systemctl'), 0o755)])

  try {
    const result = await run('sh', [deployScriptUrl.pathname, 'imaiplay-server:test'], {
      cwd: fixture,
      env: {
        ...process.env,
        PATH: `${bin}:/bin:/usr/bin`,
        CALLS_FILE: calls,
        DEPLOY_STATE: join(fixture, 'state'),
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    })
    const callLog = await readFile(calls, 'utf8')

    assert.equal(result.code, 0, result.output)
    assert.equal((callLog.match(/docker compose/g) ?? []).length, 2)
    assert.match(callLog, /systemctl restart docker/)
  } finally {
    await rm(fixture, { recursive: true, force: true })
  }
})

test('Linux release builds lock the native esbuild binary used by Vite', async () => {
  const [packageJson, lockfile] = await Promise.all([
    readFile(packageUrl, 'utf8').then(JSON.parse),
    readFile(lockfileUrl, 'utf8').then(JSON.parse),
  ])
  const esbuildVersion = lockfile.packages['admin/node_modules/esbuild']
    .optionalDependencies['@esbuild/linux-x64']

  assert.equal(packageJson.optionalDependencies['@esbuild/linux-x64'], esbuildVersion)
  assert.equal(
    lockfile.packages['node_modules/@esbuild/linux-x64'].version,
    esbuildVersion,
  )
})
