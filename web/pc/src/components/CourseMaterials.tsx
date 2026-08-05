import { DownloadOutlined, FileOutlined } from '@ant-design/icons';
import { Button, Empty } from 'antd';
import { useState } from 'react';
import {
  downloadCourseMaterial,
  type CourseMaterial,
} from '../api/course';

const fileExtension = (name: string) => {
  const extension = name.split('.').pop();
  return extension && extension !== name ? extension.toUpperCase() : 'FILE';
};

const formatSize = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.ceil(bytes / 1024)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
};

export function CourseMaterials({ materials }: { materials: CourseMaterial[] }) {
  const [downloading, setDownloading] = useState<Set<string>>(() => new Set());
  const [errors, setErrors] = useState<Record<string, string>>({});

  const download = async (material: CourseMaterial) => {
    setDownloading((current) => new Set(current).add(material.id));
    setErrors((current) => {
      const next = { ...current };
      delete next[material.id];
      return next;
    });
    let objectURL: string | undefined;
    let anchor: HTMLAnchorElement | undefined;
    try {
      const blob = await downloadCourseMaterial(material.id);
      objectURL = URL.createObjectURL(blob);
      anchor = document.createElement('a');
      anchor.href = objectURL;
      anchor.download = material.displayName;
      document.body.append(anchor);
      anchor.click();
    } catch {
      setErrors((current) => ({ ...current, [material.id]: '下载失败，请重试' }));
    } finally {
      const urlToRevoke = objectURL;
      anchor?.remove();
      if (urlToRevoke) window.setTimeout(() => URL.revokeObjectURL(urlToRevoke), 0);
      setDownloading((current) => {
        const next = new Set(current);
        next.delete(material.id);
        return next;
      });
    }
  };

  return (
    <section className="course-materials" aria-label="课程附件">
      {materials.length ? <div className="course-material-list">
        {materials.map((material) => (
          <div className="course-material-row" key={material.id}>
            <span className="course-material-icon"><FileOutlined /></span>
            <div className="course-material-copy">
              <strong title={material.displayName}>{material.displayName}</strong>
              <span>{fileExtension(material.displayName)} · {formatSize(material.sizeBytes)}</span>
              {errors[material.id] && <span className="course-material-error" role="alert">{errors[material.id]}</span>}
            </div>
            <Button
              type="text"
              icon={<DownloadOutlined />}
              loading={downloading.has(material.id)}
              onClick={() => void download(material)}
            >
              {errors[material.id] ? '重试下载' : '下载'}
            </Button>
          </div>
        ))}
      </div> : (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无课程附件" />
      )}
    </section>
  );
}
