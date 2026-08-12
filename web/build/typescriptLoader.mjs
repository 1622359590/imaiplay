import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import ts from 'typescript'

const TYPESCRIPT_EXTENSIONS = ['.ts', '.tsx']
const WORKSPACE_ROOT = fileURLToPath(new URL('../', import.meta.url))

function referencesNodeModules(value = '') {
  return value.replace(/\\/g, '/').includes('/node_modules/')
}

function isWorkspaceFile(url) {
  try {
    const parsed = new URL(url)
    if (parsed.protocol !== 'file:' || referencesNodeModules(parsed.pathname)) return false

    const relativePath = path.relative(WORKSPACE_ROOT, fileURLToPath(parsed))
    return (
      relativePath !== '..' &&
      !relativePath.startsWith(`..${path.sep}`) &&
      !path.isAbsolute(relativePath)
    )
  } catch {
    return false
  }
}

export async function resolve(specifier, context, nextResolve) {
  if (
    referencesNodeModules(specifier) ||
    referencesNodeModules(context.parentURL) ||
    !context.parentURL ||
    !isWorkspaceFile(context.parentURL) ||
    (!specifier.startsWith('./') && !specifier.startsWith('../'))
  ) {
    return nextResolve(specifier, context)
  }

  const candidateURL = new URL(specifier, context.parentURL).href
  if (!isWorkspaceFile(candidateURL)) return nextResolve(specifier, context)

  try {
    return await nextResolve(specifier, context)
  } catch (error) {
    if (
      !(error instanceof Error) ||
      !('code' in error) ||
      error.code !== 'ERR_MODULE_NOT_FOUND'
    ) {
      throw error
    }

    for (const extension of TYPESCRIPT_EXTENSIONS) {
      try {
        return await nextResolve(`${specifier}${extension}`, context)
      } catch (extensionError) {
        if (
          !(extensionError instanceof Error) ||
          !('code' in extensionError) ||
          extensionError.code !== 'ERR_MODULE_NOT_FOUND'
        ) {
          throw extensionError
        }
      }
    }

    throw error
  }
}

export async function load(url, context, nextLoad) {
  const extension = TYPESCRIPT_EXTENSIONS.find((candidate) =>
    new URL(url).pathname.endsWith(candidate),
  )
  if (!extension || !isWorkspaceFile(url)) return nextLoad(url, context)

  const source = await readFile(new URL(url), 'utf8')
  const transpiled = ts.transpileModule(source, {
    compilerOptions: {
      jsx: ts.JsxEmit.ReactJSX,
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
    },
    fileName: new URL(url).pathname,
  })

  return {
    format: 'module',
    shortCircuit: true,
    source: transpiled.outputText,
  }
}
