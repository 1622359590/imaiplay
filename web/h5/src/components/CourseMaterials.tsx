import { Button, Toast } from 'antd-mobile'
import { DownlandOutline, FileOutline } from 'antd-mobile-icons'
import { useState } from 'react'
import { downloadCourseMaterial } from '../api/course'
import type { CourseMaterial } from '../types/course'

const formatSize = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${Math.ceil(bytes / 1024)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

const extensionOf = (name: string) => {
  const extension = name.split('.').pop()
  return extension && extension !== name ? extension.toUpperCase() : 'FILE'
}

export function CourseMaterials({ materials }: { materials: CourseMaterial[] }) {
  const [downloading, setDownloading] = useState<string>()
  if (!materials.length) return null

  const download = async (material: CourseMaterial) => {
    setDownloading(material.id)
    try {
      const blob = await downloadCourseMaterial(material.id)
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = material.displayName
      anchor.click()
      URL.revokeObjectURL(url)
    } catch {
      Toast.show({ content: '资料下载失败，请稍后重试' })
    } finally {
      setDownloading(undefined)
    }
  }

  return (
    <section className="mobile-course-materials" aria-labelledby="mobile-course-materials-title">
      <div className="materials-heading">
        <div>
          <h2 id="mobile-course-materials-title">学习资料</h2>
          <p>课程配套文件，可随时下载查看</p>
        </div>
        <span>{materials.length} 份</span>
      </div>
      <div className="mobile-course-material-list">
        {materials.map((material) => (
          <div className="mobile-course-material-row" key={material.id}>
            <span className="mobile-course-material-icon"><FileOutline /></span>
            <div className="mobile-course-material-copy">
              <strong>{material.displayName}</strong>
              <small>{extensionOf(material.displayName)} · {formatSize(material.sizeBytes)}</small>
            </div>
            <Button
              fill="none"
              size="small"
              loading={downloading === material.id}
              disabled={Boolean(downloading) && downloading !== material.id}
              onClick={() => void download(material)}
            >
              <DownlandOutline /> 下载
            </Button>
          </div>
        ))}
      </div>
    </section>
  )
}
