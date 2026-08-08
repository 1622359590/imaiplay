import { ApartmentOutlined, ReadOutlined } from '@ant-design/icons';
import { Avatar, Card, message, Spin, Typography } from 'antd';
import { Navigate, useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { userFacingErrorMessage } from '@imaiplay/shared/api/errors';
import { useAuth } from '../context/AuthContext';
import { usePortal } from '../context/PortalContext';
import { performLoginNavigation } from '../utils/portalRouting';

export function OrganizationSelectPage() {
  const navigate = useNavigate();
  const { pendingSelection, selectOrganization } = useAuth();
  const { mode, tenantCode } = usePortal();
  const [submittingCode, setSubmittingCode] = useState<string>();

  if (!pendingSelection) return <Navigate to="/login" replace />;

  const chooseOrganization = async (organization: typeof pendingSelection.organizations[number]) => {
    if (submittingCode) return;
    setSubmittingCode(organization.code);
    try {
      const redirect = await selectOrganization(organization, mode, tenantCode);
      performLoginNavigation(redirect, navigate);
    } catch (error) {
      message.error(userFacingErrorMessage(error, '企业选择失败，请重新登录'));
      setSubmittingCode(undefined);
    }
  };

  return (
    <div className="login-page organization-select-page">
      <div className="login-container reveal">
        <div className="login-brand">
          <span className="login-brand-mark"><ReadOutlined /></span>
          <span>iMaiPlay</span>
        </div>
        <Card className="login-card organization-select-card glass-card" bordered={false}>
          <Typography.Title level={2} className="gradient-text">选择你的企业</Typography.Title>
          <Typography.Paragraph type="secondary">
            这个账号可进入多个企业学习中心，请选择本次要进入的企业。
          </Typography.Paragraph>
          <div className="organization-list">
            {pendingSelection.organizations.map((organization) => (
              <button
                className="organization-option"
                disabled={Boolean(submittingCode)}
                key={organization.code}
                onClick={() => void chooseOrganization(organization)}
                type="button"
              >
                <Avatar
                  className="organization-logo"
                  icon={<ApartmentOutlined />}
                  shape="square"
                  size={46}
                  src={organization.logo_url}
                />
                <span className="organization-copy">
                  <strong>{organization.name}</strong>
                  <small>{organization.role === 'learner' ? '学员' : organization.role}</small>
                </span>
                {submittingCode === organization.code
                  ? <Spin size="small" />
                  : <span className="organization-enter">进入</span>}
              </button>
            ))}
          </div>
        </Card>
      </div>
    </div>
  );
}
