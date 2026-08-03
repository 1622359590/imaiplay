import { Result } from 'antd';
import { portalErrorContent } from '../utils/portalRouting';

interface PortalErrorPageProps {
  error?: unknown;
}

export function PortalErrorPage({ error }: PortalErrorPageProps) {
  const status = typeof error === 'object' && error !== null && 'response' in error
    ? (error as { response?: { status?: number } }).response?.status
    : undefined;
  const content = portalErrorContent(error);

  return (
    <Result
      status={status === 403 ? '403' : status === 404 ? '404' : 'error'}
      title={content.title}
      subTitle={content.description}
    />
  );
}
