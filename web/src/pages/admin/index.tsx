import React, { useEffect } from 'react';
import { Card, Tabs } from 'antd';
import { useParams, useNavigate } from 'react-router-dom';
import UserTab from './UserTab';
import ModelTab from './ModelTab';
import PolicyTab from './PolicyTab';
import HealthTab from './HealthTab';
import SystemTab from './SystemTab';
import AdminLogs from './AdminLogs';
import AdminTopUsers from './AdminTopUsers';


const AdminShell: React.FC = () => {
  const { tab } = useParams<{ tab?: string }>();
  const navigate = useNavigate();

  const activeTabKey = tab || 'users';

  useEffect(() => {
    const validTabs = ['users', 'models', 'policies', 'health', 'system', 'logs', 'top7d'];
    const currentTab = tab || 'users';
    if (!validTabs.includes(currentTab)) {
      navigate('/admin/users', { replace: true });
    }
  }, [tab, navigate]);

  const handleTabChange = (key: string) => {
    navigate(`/admin/${key}`);
  };

  return (
    <Card title="管理后台">
      <Tabs activeKey={activeTabKey} onChange={handleTabChange}>
        <Tabs.TabPane tab="用户管理" key="users">
          <UserTab />
        </Tabs.TabPane>
        <Tabs.TabPane tab="模型管理" key="models">
          <ModelTab />
        </Tabs.TabPane>
        <Tabs.TabPane tab="配额策略" key="policies">
          <PolicyTab />
        </Tabs.TabPane>
        <Tabs.TabPane tab="健康监控" key="health">
          <HealthTab />
        </Tabs.TabPane>
        <Tabs.TabPane tab="系统配置" key="system">
          <SystemTab />
        </Tabs.TabPane>
        <Tabs.TabPane tab="全员访问日志" key="logs">
          <AdminLogs />
        </Tabs.TabPane>
        <Tabs.TabPane tab="最近7天TOP用户" key="top7d">
          <AdminTopUsers />
        </Tabs.TabPane>
      </Tabs>
    </Card>
  );
};

export default AdminShell;
