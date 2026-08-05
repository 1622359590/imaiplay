import { DownloadOutlined, FileOutlined } from '@ant-design/icons';
import { Button, Typography, message } from 'antd';
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
  const [downloading, setDownloading] = useState<string>();
  if (!materials.length) return null;

  const download = async (material: CourseMaterial) => {
    setDownloading(material.id);
    try {
      const blob = await downloadCourseMaterial(material.id);
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = material.displayName;
      anchor.click();
      URL.revokeObjectURL(url);
    } catch {
      message.error('资料下载失败，请稍后重试');
    } finally {
      setDownloading(undefined);
    }
  };

  return (
    <section className="course-materials" aria-labelledby="course-materials-title">
      <Typography.Title level={2} id="course-materials-title">学习资料</Typography.Title>
      <div className="course-material-list">
        {materials.map((material) => (
          <div className="course-material-row" key={material.id}>
            <span className="course-material-icon"><FileOutlined /></span>
            <div className="course-material-copy">
              <strong title={material.displayName}>{material.displayName}</strong>
              <span>{fileExtension(material.displayName)} · {formatSize(material.sizeBytes)}</span>
            </div>
            <Button
              type="text"
              icon={<DownloadOutlined />}
              loading={downloading === material.id}
              disabled={Boolean(downloading) && downloading !== material.id}
              onClick={() => void download(material)}
            >
              下载
            </Button>
          </div>
        ))}
      </div>
    </section>
  );
}
