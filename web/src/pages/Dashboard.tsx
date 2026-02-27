import React, { useEffect, useState } from 'react';
import { Layout, Card, Button, Table, Tag, message, Modal, Form, Input, Descriptions } from 'antd';
import { PlusOutlined, CopyOutlined, DeleteOutlined, LogoutOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

const { Header, Content } = Layout;

interface APIKey {
  id: string;
  name: string;
  key_prefix: string;
  created_at: string;
  last_used_at?: string;
  enabled: boolean;
}

const Dashboard: React.FC = () => {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [quota, setQuota] = useState<any>({});
  const [modalVisible, setModalVisible] = useState(false);
  const [newKey, setNewKey] = useState('');
  const [form] = Form.useForm();
  const navigate = useNavigate();
  
  const user = JSON.parse(localStorage.getItem('user') || '{}');
  const token = localStorage.getItem('token');

  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;

  const fetchData = async () => {
    try {
      const [keysRes, quotaRes] = await Promise.all([
        axios.get('http://localhost:8080/api/v1/user/keys'),
        axios.get('http://localhost:8080/api/v1/user/quota'),
      ]);
      setKeys(keysRes.data);
      setQuota(quotaRes.data);
    } catch (err) {
      message.error('获取数据失败');
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleCreate = async (values: { name: string }) => {
    try {
      const res = await axios.post('http://localhost:8080/api/v1/user/keys', values);
      setNewKey(res.data.key);
      setModalVisible(false);
      form.resetFields();
      fetchData();
    } catch (err) {
      message.error('创建失败');
    }
  };

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除后该 API Key 将无法使用',
      onOk: async () => {
        try {
          await axios.delete(`http://localhost:8080/api/v1/user/keys/${id}`);
          message.success('删除成功');
          fetchData();
        } catch (err) {
          message.error('删除失败');
        }
      },
    });
  };

  const logout = () => {
    localStorage.clear();
    navigate('/login');
  };

  const columns = [
    { title: '名称', dataIndex: 'name' },
    { title: 'Key', render: (r: APIKey) => `${r.key_prefix}****` },
    { title: '创建时间', dataIndex: 'created_at' },
    { 
      title: '状态', 
      dataIndex: 'enabled',
      render: (v: boolean) => v ? <Tag color="green">启用</Tag> : <Tag>禁用</Tag>
    },
    {
      title: '操作',
      render: (_: any, record: APIKey) => (
        <Button 
          icon={<DeleteOutlined />} 
          danger 
          size="small"
          onClick={() => handleDelete(record.id)}
        >
          删除
        </Button>
      ),
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ color: '#fff', margin: 0 }}>LLMGATE 控制台</h2>
        <div>
          <span style={{ color: '#fff', marginRight: 16 }}>{user.name}</span>
          {user.role === 'admin' && (
            <Button type="link" onClick={() => navigate('/admin')}>管理后台</Button>
          )}
          <Button icon={<LogoutOutlined />} onClick={logout}>退出</Button>
        </div>
      </Header>
      <Content style={{ padding: 24 }}>
        <Card title="配额使用情况" style={{ marginBottom: 24 }}>
          <Descriptions>
            <Descriptions.Item label="速率限制">{quota.rate_limit} 请求/分钟</Descriptions.Item>
            <Descriptions.Item label="Token 配额">{quota.token_used} / {quota.token_limit}</Descriptions.Item>
            <Descriptions.Item label="可用模型">{quota.models?.join(', ')}</Descriptions.Item>
          </Descriptions>
        </Card>

        <Card 
          title="API Keys" 
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
              创建 Key
            </Button>
          }
        >
          <Table dataSource={keys} columns={columns} rowKey="id" />
        </Card>

        <Modal
          title="创建 API Key"
          open={modalVisible}
          onCancel={() => setModalVisible(false)}
          footer={null}
        >
          {newKey ? (
            <div>
              <p>请保存您的 API Key，它只会显示一次：</p>
              <pre style={{ background: '#f5f5f5', padding: 16 }}>{newKey}</pre>
              <Button 
                icon={<CopyOutlined />} 
                onClick={() => {
                  navigator.clipboard.writeText(newKey);
                  message.success('已复制');
                }}
              >
                复制
              </Button>
              <Button onClick={() => { setNewKey(''); setModalVisible(false); }} style={{ marginLeft: 8 }}>
                关闭
              </Button>
            </div>
          ) : (
            <Form form={form} onFinish={handleCreate}>
              <Form.Item name="name" rules={[{ required: true }]} label="名称">
                <Input placeholder="如：开发测试" />
              </Form.Item>
              <Button type="primary" htmlType="submit">创建</Button>
            </Form>
          )}
        </Modal>
      </Content>
    </Layout>
  );
};

export default Dashboard;