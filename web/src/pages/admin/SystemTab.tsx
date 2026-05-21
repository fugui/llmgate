import React, { useEffect, useState } from 'react';
import { Card, Button, Form, Input, Space, Row, Col, Switch, InputNumber, Statistic, message } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import api from '../../api';

const SystemTab: React.FC = () => {
  const [configForm] = Form.useForm();
  const [concurrencyForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({
    usersCount: 0,
    modelsCount: 0,
    policiesCount: 0,
    backendsCount: 0,
  });

  const [messageApi, contextHolder] = message.useMessage();

  const fetchSystemConfig = async () => {
    try {
      const res = await api.get('/api/v1/config/frontend');
      configForm.setFieldsValue(res.data.data);
    } catch {
      messageApi.error('获取系统配置失败');
    }
  };

  const fetchConcurrencyConfig = async () => {
    try {
      const res = await api.get('/api/v1/admin/config/concurrency');
      concurrencyForm.setFieldsValue(res.data.data);
    } catch {
      messageApi.error('获取并发配置失败');
    }
  };

  const fetchStats = async () => {
    try {
      const [usersRes, modelsRes, policiesRes] = await Promise.all([
        api.get('/api/v1/admin/users', { params: { page: 1, page_size: 1 } }),
        api.get('/api/v1/admin/models'),
        api.get('/api/v1/admin/policies'),
      ]);

      const modelsList = modelsRes.data.data || [];
      let totalBackends = 0;

      // Fetch backends for each model in parallel to sum up backends count
      await Promise.all(
        modelsList.map(async (model: { id: string }) => {
          try {
            const backendsRes = await api.get(`/api/v1/admin/models/${encodeURIComponent(model.id)}/backends`);
            totalBackends += (backendsRes.data.data || []).length;
          } catch {
            // Ignore single errors
          }
        })
      );

      setStats({
        usersCount: usersRes.data.total || 0,
        modelsCount: modelsList.length,
        policiesCount: (policiesRes.data.data || []).length,
        backendsCount: totalBackends,
      });
    } catch {
      messageApi.error('获取统计数据失败');
    }
  };

  const handleSystemConfigSubmit = async (values: any) => {
    try {
      await api.put('/api/v1/admin/config/frontend', values);
      messageApi.success('系统配置保存成功');
      fetchSystemConfig();
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      messageApi.error(error.response?.data?.error || '保存系统配置失败');
    }
  };

  const handleConcurrencyConfigSubmit = async (values: { user_limit: number }) => {
    try {
      await api.put('/api/v1/admin/config/concurrency', values);
      messageApi.success('并发配置保存成功，已动态生效');
      fetchConcurrencyConfig();
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      messageApi.error(error.response?.data?.error || '保存并发配置失败');
    }
  };

  const loadAllData = async () => {
    setLoading(true);
    await Promise.all([
      fetchSystemConfig(),
      fetchConcurrencyConfig(),
      fetchStats(),
    ]);
    setLoading(false);
  };

  useEffect(() => {
    loadAllData();
  }, []);

  return (
    <div>
      {contextHolder}
      <Row gutter={16}>
        <Col span={12}>
          <Card title="前端配置" style={{ marginBottom: 16 }}>
            <Form
              form={configForm}
              layout="horizontal"
              labelCol={{ span: 6 }}
              wrapperCol={{ span: 18 }}
              onFinish={handleSystemConfigSubmit}
            >
              <Form.Item
                name="feedback_url"
                label="反馈链接"
                rules={[{ type: 'url', message: '请输入有效的URL' }]}
              >
                <Input placeholder="如：https://example.com/feedback" />
              </Form.Item>
              <Form.Item
                name="dev_manual_url"
                label="开发手册链接"
                rules={[{ type: 'url', message: '请输入有效的URL' }]}
              >
                <Input placeholder="如：https://example.com/docs" />
              </Form.Item>
              <Form.Item
                name="sso_enabled"
                label="SSO 启用"
                valuePropName="checked"
              >
                <Switch disabled />
              </Form.Item>
              <Form.Item wrapperCol={{ offset: 6, span: 18 }}>
                <Space>
                  <Button type="primary" onClick={() => configForm.submit()} loading={loading}>
                    保存
                  </Button>
                  <Button
                    icon={<ReloadOutlined />}
                    onClick={fetchSystemConfig}
                  >
                    刷新
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
          <Card title="并发控制" style={{ marginBottom: 16 }}>
            <Form
              form={concurrencyForm}
              layout="horizontal"
              labelCol={{ span: 8 }}
              wrapperCol={{ span: 16 }}
              onFinish={handleConcurrencyConfigSubmit}
            >
              <Form.Item
                name="user_limit"
                label="用户并发限制"
                rules={[{ required: true, message: '请输入用户并发限制' }]}
                extra="每个用户最大并发请求数，0 表示不限制"
              >
                <InputNumber min={0} max={1000} style={{ width: '100%' }} placeholder="如：10" />
              </Form.Item>
              <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
                <Space>
                  <Button type="primary" onClick={() => concurrencyForm.submit()} loading={loading}>
                    保存
                  </Button>
                  <Button
                    icon={<ReloadOutlined />}
                    onClick={fetchConcurrencyConfig}
                  >
                    刷新
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="系统统计">
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Statistic
                  title="总用户数"
                  value={stats.usersCount}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="总模型数"
                  value={stats.modelsCount}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="总策略数"
                  value={stats.policiesCount}
                  valueStyle={{ color: '#722ed1' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="总后端数"
                  value={stats.backendsCount}
                  valueStyle={{ color: '#fa8c16' }}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default SystemTab;
