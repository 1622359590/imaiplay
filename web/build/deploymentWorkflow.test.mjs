import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const workflowUrl = new URL('../../.github/workflows/deploy.yml', import.meta.url)
const composeOverrideUrl = new URL('../../docker-compose.deploy.yml', import.meta.url)

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
