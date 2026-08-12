import { readFile } from 'node:fs/promises'

import ts from 'typescript'

const TYPESCRIPT_EXTENSIONS = ['.ts', '.tsx']

export async function resolve(specifier, context, nextResolve) {
  try {
    return await nextResolve(specifier, context)
  } catch (error) {
    if (
      !(error instanceof Error) ||
      !('code' in error) ||
      error.code !== 'ERR_MODULE_NOT_FOUND' ||
      (!specifier.startsWith('./') && !specifier.startsWith('../'))
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
  const extension = TYPESCRIPT_EXTENSIONS.find((candidate) => url.endsWith(candidate))
  if (!extension) return nextLoad(url, context)

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
