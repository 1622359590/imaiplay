import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import path from 'node:path'
import { pathToFileURL } from 'node:url'
import { gzipSync } from 'node:zlib'

const APPLICATIONS = ['admin', 'pc', 'h5']

function listJavaScriptFiles(directory, prefix = '') {
  return readdirSync(directory, { withFileTypes: true })
    .flatMap((entry) => {
      const relativePath = path.join(prefix, entry.name)
      const absolutePath = path.join(directory, entry.name)

      if (entry.isDirectory()) return listJavaScriptFiles(absolutePath, relativePath)
      return entry.isFile() && entry.name.endsWith('.js') ? [relativePath] : []
    })
    .sort()
}

export function inspectBundle(root, limitBytes = 500_000) {
  const applications = APPLICATIONS.map((app) => {
    const assetsDirectory = path.join(root, app, 'dist', 'assets')
    if (!existsSync(assetsDirectory) || !statSync(assetsDirectory).isDirectory()) {
      throw new Error(
        `Missing bundle assets directory for ${app}: ${assetsDirectory}. Run "npm run build:all" first.`,
      )
    }

    const assets = listJavaScriptFiles(assetsDirectory).map((file) => {
      const contents = readFileSync(path.join(assetsDirectory, file))
      return {
        app,
        file: file.split(path.sep).join('/'),
        rawBytes: contents.byteLength,
        gzipBytes: gzipSync(contents).byteLength,
      }
    })
    const maximum = assets.reduce(
      (largest, asset) => (largest === null || asset.rawBytes > largest.rawBytes ? asset : largest),
      null,
    )

    return {
      app,
      assets,
      maximum,
      totalRawBytes: assets.reduce((total, asset) => total + asset.rawBytes, 0),
      totalGzipBytes: assets.reduce((total, asset) => total + asset.gzipBytes, 0),
      chunkCount: assets.length,
    }
  })

  const oversized = applications.flatMap((application) =>
    application.assets
      .filter((asset) => asset.rawBytes > limitBytes)
      .map((asset) => ({
        app: asset.app,
        file: asset.file,
        size: asset.rawBytes,
        limit: limitBytes,
      })),
  )

  return { applications, oversized }
}

function formatBytes(bytes) {
  return `${bytes.toLocaleString('en-US')} B`
}

function printReport(report, limitBytes) {
  console.log(`JavaScript bundle budget: ${formatBytes(limitBytes)} raw per chunk`)

  for (const application of report.applications) {
    console.log(`\n${application.app}`)
    for (const asset of application.assets) {
      console.log(`  ${asset.file}: raw ${formatBytes(asset.rawBytes)}, gzip ${formatBytes(asset.gzipBytes)}`)
    }

    const maximum = application.maximum
      ? `${application.maximum.file} (${formatBytes(application.maximum.rawBytes)} raw, ${formatBytes(application.maximum.gzipBytes)} gzip)`
      : 'none'
    console.log(
      `  Summary: ${application.chunkCount} chunks, max ${maximum}, total ${formatBytes(application.totalRawBytes)} raw / ${formatBytes(application.totalGzipBytes)} gzip`,
    )
  }
}

function runCli() {
  const limitBytes = 500_000

  try {
    const report = inspectBundle(process.cwd(), limitBytes)
    printReport(report, limitBytes)

    if (report.oversized.length > 0) {
      console.error('\nBundle budget exceeded:')
      for (const asset of report.oversized) {
        console.error(
          `  ${asset.app}/${asset.file}: ${formatBytes(asset.size)} exceeds ${formatBytes(asset.limit)}`,
        )
      }
      process.exitCode = 1
    }
  } catch (error) {
    console.error(error instanceof Error ? error.message : error)
    process.exitCode = 1
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href) {
  runCli()
}
