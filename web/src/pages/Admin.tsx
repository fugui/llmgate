import React, { useEffect, useState } from 'react';
import { Layout, Card, Table, Button, Tag, message, Tabs } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import api from '../api';

const { Header, Content } = Layout;

const Admin: React.FC = () => {
  const [users, setUsers] = useState([]);
  const [models, setModels] = useState([]);
  const [policies, setPolicies] = useState([]);
  const [activeTab, setActiveTab] = useState('users');
  const navigate = useNavigate();

  const [messageApi, contextHolder] = message.useMessage();

  const fetchData = async () => {
    try {
      const [usersRes, modelsRes, policiesRes] = await Promise.all([
        api.get('/api/v1/admin/users'),
        api.get('/api/v1/admin/models'),
        api.get('/api/v1/admin/policies'),
      ]);
      setUsers(usersRes.data.data || []);
      setModels(modelsRes.data.data || []);
      setPolicies(policiesRes.data.data || []);
    } catch (err) {
      messageApi.error('获取数据失败');
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const userColumns = [
    { title: '邮箱', dataIndex: 'email' },
    { title: '姓名', dataIndex: 'name' },
    { title: '角色', dataIndex: 'role' },
    { title: '部门', dataIndex: 'department' },
    { 
      title: '配额策略', 
      dataIndex: 'quota_policy',
      render: (v: string) => <Tag>{v}</Tag>
    },
    {
      title: '操作',
      render: () => <Button size="small">编辑</Button>,
    },
  ];

  const modelColumns = [
    { title: 'ID', dataIndex: 'model_id' },
    { title: '名称', dataIndex: 'name' },
    { title: '后端地址', dataIndex: 'backend_url' },
    { 
      title: '状态', 
      dataIndex: 'enabled',
      render: (v: boolean) => v ? <Tag color="green">启用</Tag> : <Tag>禁用</Tag>
    },
  ];

  const policyColumns = [
    { title: '名称', dataIndex: 'name' },
    { title: '速率限制', dataIndex: 'rate_limit', render: (v: number) => `${v}/min` },
    { title: 'Token 日限额', dataIndex: 'token_quota_daily' },
    { title: 'Token 月限额', dataIndex: 'token_quota_monthly' },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Button 
            icon={<ArrowLeftOutlined />} 
            onClick={() => navigate('/dashboard')}
            style={{ marginRight: 16 }}
          >
            返回
          </Button>
          <h2 style={{ color: '#fff', margin: 0 }}>管理后台</h2>
        </div>
      </Header>
      <Content style={{ padding: 24 }}>
        {contextHolder}
        <Card>
          <Tabs activeKey={activeTab} onChange={setActiveTab}>
            <Tabs.TabPane tab="用户管理" key="users">
              <Table dataSource={users} columns={userColumns} rowKey="id" />
            </Tabs.TabPane>
            <Tabs.TabPane tab="模型管理" key="models">
              <Table dataSource={models} columns={modelColumns} rowKey="id" />
            </Tabs.TabPane>
            <Tabs.TabPane tab="配额策略" key="policies">
              <Table dataSource={policies} columns={policyColumns} rowKey="id" />
            </Tabs.TabPane>
          </Tabs>
        </Card>
      </Content>
    </Layout>
  );
};

export default Admin;