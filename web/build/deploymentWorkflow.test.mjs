import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const workflowUrl = new URL('../../.github/workflows/deploy.yml', import.meta.url)
const composeOverrideUrl = new URL('../../docker-compose.deploy.yml', import.meta.url)
const packageUrl = new URL('../package.json', import.meta.url)
const lockfileUrl = new URL('../package-lock.json', import.meta.url)

test('deployment builds the release image on the GitHub runner', async () => {
  const workflow = await readFile(workflowUrl, 'utf8')

  assert.match(workflow, /uses: actions\/checkout@v4/)
  assert.match(workflow, /docker build --tag "\$release_image" \./)
  assert.doesNotMatch(workflow, /docker compose[^\n]*up[^\n]*--build/)
})

test('deployment streams the immutable image and starts it without rebuilding', async () => {
  const [workflow, composeOverride] = await Promise.all([
    readFile(workflowUrl, 'utf8'),
    readFile(composeOverrideUrl, 'utf8'),
  ])

  assert.match(workflow, /docker save "\$release_image" \| gzip -1 \| ssh/)
  assert.match(workflow, /docker load/)
  assert.match(workflow, /IMAI_PLAY_IMAGE=\$release_image_quoted/)
  assert.match(workflow, /docker-compose\.deploy\.yml/)
  assert.match(workflow, /up -d --no-build --remove-orphans/)
  assert.match(composeOverride, /image: \$\{IMAI_PLAY_IMAGE:\?IMAI_PLAY_IMAGE is required\}/)
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
