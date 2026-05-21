import React, { useEffect, useState } from 'react';
import { Table, Button, Tag, message, Modal, Form, Input, Space, Popconfirm, Tooltip, InputNumber, Select } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, MinusCircleOutlined } from '@ant-design/icons';
import api from '../../api';
import type { Policy, PolicyFormValues, Model, TimeRange } from './types';


const PolicyTab: React.FC = () => {
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(false);

  // Policy modal states
  const [policyModalVisible, setPolicyModalVisible] = useState(false);
  const [policyModalTitle, setPolicyModalTitle] = useState('创建策略');
  const [editingPolicy, setEditingPolicy] = useState<Policy | null>(null);
  const [policyForm] = Form.useForm();

  const [messageApi, contextHolder] = message.useMessage();

  const fetchModels = async () => {
    try {
      const res = await api.get('/api/v1/admin/models');
      setModels(res.data.data || []);
    } catch {
      messageApi.error('获取模型列表失败');
    }
  };

  const fetchPolicies = async () => {
    setLoading(true);
    try {
      const res = await api.get('/api/v1/admin/policies');
      setPolicies(res.data.data || []);
    } catch {
      messageApi.error('获取策略列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPolicies();
    fetchModels();
  }, []);

  const handleCreatePolicy = () => {
    setEditingPolicy(null);
    setPolicyModalTitle('创建策略');
    policyForm.resetFields();
    policyForm.setFieldsValue({
      enabled: true,
      rate_limit: 60,
      rate_limit_window: 60,
      request_quota_daily: 1000,
      available_time_ranges: [],
      default_model: '',
    });
    setPolicyModalVisible(true);
  };

  const handleEditPolicy = (policy: Policy) => {
    setEditingPolicy(policy);
    setPolicyModalTitle('编辑策略');
    policyForm.setFieldsValue({
      name: policy.name,
      description: policy.description,
      rate_limit: policy.rate_limit,
      rate_limit_window: policy.rate_limit_window,
      request_quota_daily: policy.request_quota_daily,
      available_time_ranges: policy.available_time_ranges || [],
      models: policy.models || [],
      default_model: policy.default_model,
    });
    setPolicyModalVisible(true);
  };

  const handleDeletePolicy = async (policy: Policy) => {
    try {
      await api.delete(`/api/v1/admin/policies/${policy.name}`);
      messageApi.success('策略删除成功');
      fetchPolicies();
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      messageApi.error(error.response?.data?.error || '删除失败');
    }
  };

  const handlePolicySubmit = async (values: PolicyFormValues) => {
    try {
      if (editingPolicy) {
        await api.put(`/api/v1/admin/policies/${editingPolicy.name}`, {
          description: values.description,
          rate_limit: values.rate_limit,
          rate_limit_window: values.rate_limit_window,
          request_quota_daily: values.request_quota_daily,
          available_time_ranges: values.available_time_ranges || [],
          models: values.models,
          default_model: values.default_model,
        });
        messageApi.success('策略更新成功');
      } else {
        await api.post('/api/v1/admin/policies', {
          name: values.name,
          description: values.description,
          rate_limit: values.rate_limit,
          rate_limit_window: values.rate_limit_window,
          request_quota_daily: values.request_quota_daily,
          available_time_ranges: values.available_time_ranges || [],
          models: values.models,
          default_model: values.default_model,
        });
        messageApi.success('策略创建成功');
      }
      setPolicyModalVisible(false);
      fetchPolicies();
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      messageApi.error(error.response?.data?.error || '操作失败');
    }
  };

  const policyColumns = [
    { title: '名称', dataIndex: 'name' },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    {
      title: '速率限制',
      dataIndex: 'rate_limit',
      render: (v: number, record: Policy) => `${v}/${record.rate_limit_window || 60}s`,
    },
    {
      title: '每日限额',
      dataIndex: 'request_quota_daily',
      render: (v: number) => v === 0 ? <Tag>无限制</Tag> : v,
    },
    {
      title: '关联模型',
      dataIndex: 'models',
      render: (modelsList: string[]) =>
        modelsList?.length > 0
          ? <Space size="small">{modelsList.map(m => <Tag key={m}>{m}</Tag>)}</Space>
          : '-',
    },
    {
      title: '可用时段',
      dataIndex: 'available_time_ranges',
      render: (ranges: TimeRange[]) =>
        ranges && ranges.length > 0
          ? <Space size="small" wrap>{ranges.map((r, i) => <Tag key={i} color="cyan">{r.start}-{r.end}</Tag>)}</Space>
          : <Tag>全天</Tag>,
    },
    {
      title: '默认模型',
      dataIndex: 'default_model',
      render: (model: string) => model ? <Tag color="blue">{model}</Tag> : '-',
    },
    {
      title: '操作',
      render: (_: unknown, record: Policy) => (
        <Space>
          <Tooltip title="编辑">
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEditPolicy(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确认删除"
            description={`删除策略 "${record.name}"，确定要继续吗？`}
            onConfirm={() => handleDeletePolicy(record)}
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Tooltip title="删除">
              <Button type="link" danger size="small" icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {contextHolder}
      <div style={{ marginBottom: 16 }}>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={handleCreatePolicy}
        >
          创建策略
        </Button>
      </div>
      <Table
        dataSource={policies}
        columns={policyColumns}
        rowKey="name"
        loading={loading}
      />

      {/* Policy Create/Edit Modal */}
      <Modal
        title={policyModalTitle}
        open={policyModalVisible}
        onCancel={() => setPolicyModalVisible(false)}
        onOk={() => policyForm.submit()}
        okText={editingPolicy ? '保存' : '创建'}
        width={600}
        destroyOnClose
      >
        <Form
          form={policyForm}
          onFinish={handlePolicySubmit}
          layout="horizontal"
          labelCol={{ span: 6 }}
          wrapperCol={{ span: 18 }}
          style={{ marginTop: 24 }}
        >
          <Form.Item
            name="name"
            label="策略名称"
            rules={[{ required: !editingPolicy, message: '请输入策略名称' }]}
            extra="唯一标识，如：default, premium"
            hidden={!!editingPolicy}
          >
            <Input disabled={!!editingPolicy} placeholder="如：premium" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea rows={2} placeholder="策略描述信息（可选）" />
          </Form.Item>

          <Form.Item
            name="rate_limit"
            label="速率限制"
            rules={[{ required: true, message: '请输入速率限制' }]}
            extra="单位时间内允许的请求次数"
          >
            <InputNumber min={1} max={10000} style={{ width: '100%' }} placeholder="如：60" />
          </Form.Item>

          <Form.Item
            name="rate_limit_window"
            label="时间窗口"
            rules={[{ required: true, message: '请输入时间窗口' }]}
            extra="速率限制的时间窗口（秒）"
          >
            <InputNumber min={1} max={3600} style={{ width: '100%' }} placeholder="如：60" />
          </Form.Item>

          <Form.Item
            name="request_quota_daily"
            label="每日限额"
            rules={[{ required: true, message: '请输入每日限额' }]}
            extra="每天允许的请求次数（0表示无限制）"
          >
            <InputNumber min={0} max={1000000} style={{ width: '100%' }} placeholder="如：1000" />
          </Form.Item>

          <Form.Item label="可用时段" extra="不添加任何时段表示全天可用，支持跨午夜（如 22:00-06:00）">
            <Form.List name="available_time_ranges">
              {(fields, { add, remove }) => (
                <>
                  {fields.map(({ key, name, ...restField }) => (
                    <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                      <Form.Item
                        {...restField}
                        name={[name, 'start']}
                        rules={[{ required: true, message: '请输入开始时间' }, { pattern: /^([01]\d|2[0-4]):([0-5]\d)$/, message: 'HH:MM 格式' }]}
                        style={{ marginBottom: 0 }}
                      >
                        <Input placeholder="00:00" style={{ width: 90 }} />
                      </Form.Item>
                      <span>-</span>
                      <Form.Item
                        {...restField}
                        name={[name, 'end']}
                        rules={[{ required: true, message: '请输入结束时间' }, { pattern: /^([01]\d|2[0-4]):([0-5]\d)$/, message: 'HH:MM 格式' }]}
                        style={{ marginBottom: 0 }}
                      >
                        <Input placeholder="24:00" style={{ width: 90 }} />
                      </Form.Item>
                      <MinusCircleOutlined onClick={() => remove(name)} style={{ color: '#ff4d4f' }} />
                    </Space>
                  ))}
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    添加时段
                  </Button>
                </>
              )}
            </Form.List>
          </Form.Item>

          <Form.Item
            name="models"
            label="关联模型"
            extra="选择该策略允许的模型（多选）"
          >
            <Select
              mode="multiple"
              placeholder="请选择关联模型"
              style={{ width: '100%' }}
              options={[
                { label: '全部模型 (*)', value: '*' },
                ...models.map(model => ({ label: model.name, value: model.id }))
              ]}
            />
          </Form.Item>
          <Form.Item
            name="default_model"
            label="默认模型"
            extra="该策略所对应的默认回退模型（可选）"
          >
            <Select
              allowClear
              placeholder="请选择默认模型"
              style={{ width: '100%' }}
              options={models.map(model => ({ label: model.name, value: model.id }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default PolicyTab;
