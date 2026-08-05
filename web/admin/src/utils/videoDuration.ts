export interface VideoMetadataElement {
  duration: number
  onloadedmetadata: ((event: Event) => unknown) | null
  onerror: ((event: Event) => unknown) | null
  preload: string
  src: string
}

export interface VideoDurationDependencies {
  createObjectURL: (file: Blob) => string
  revokeObjectURL: (url: string) => void
  createVideo: () => VideoMetadataElement
}

const browserDependencies: VideoDurationDependencies = {
  createObjectURL: (file) => URL.createObjectURL(file),
  revokeObjectURL: (url) => URL.revokeObjectURL(url),
  createVideo: () => document.createElement('video'),
}

export function readVideoDurationSeconds(
  file: Pick<File, 'type'>,
  dependencies: VideoDurationDependencies = browserDependencies,
): Promise<number> {
  return new Promise((resolve, reject) => {
    const url = dependencies.createObjectURL(file as Blob)
    const video = dependencies.createVideo()
    const finish = (callback: () => void) => {
      video.onloadedmetadata = null
      video.onerror = null
      dependencies.revokeObjectURL(url)
      callback()
    }

    video.preload = 'metadata'
    video.onloadedmetadata = () => {
      if (!Number.isFinite(video.duration) || video.duration <= 0) {
        finish(() => reject(new Error('无法读取视频时长')))
        return
      }
      finish(() => resolve(Math.ceil(video.duration)))
    }
    video.onerror = () => {
      finish(() => reject(new Error('无法读取视频时长')))
    }
    video.src = url
  })
}
