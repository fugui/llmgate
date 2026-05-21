import React, { useEffect, useState } from 'react';
import { Card, Table, Button, Tag, message, Row, Col, Statistic } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined, ReloadOutlined } from '@ant-design/icons';
import api from '../../api';
import type { Model, Backend, BackendHealth } from './types';


const HealthTab: React.FC = () => {
  const [models, setModels] = useState<Model[]>([]);
  const [backends, setBackends] = useState<Backend[]>([]);
  const [healthStatus, setHealthStatus] = useState<Record<string, BackendHealth>>({});
  const [loading, setLoading] = useState(false);
  const [messageApi, contextHolder] = message.useMessage();

  const fetchHealthStatus = async () => {
    try {
      const res = await api.get('/api/v1/admin/models/health');
      setHealthStatus(res.data.data || {});
    } catch {
      messageApi.error('获取健康状态失败');
    }
  };

  const fetchAllBackends = async (currentModels: Model[]) => {
    try {
      const allBackends: Backend[] = [];
      for (const model of currentModels) {
        try {
          const res = await api.get(`/api/v1/admin/models/${encodeURIComponent(model.id)}/backends`);
          const modelBackends = (res.data.data || []).map((b: Backend) => ({
            ...b,
            model_id: model.id,
            model_name: model.name,
          }));
          allBackends.push(...modelBackends);
        } catch {
          // Skip failed requests
        }
      }
      setBackends(allBackends);
    } catch {
      messageApi.error('获取后端列表失败');
    }
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      const modelsRes = await api.get('/api/v1/admin/models');
      const modelsList = modelsRes.data.data || [];
      setModels(modelsList);
      
      // Fetch health and backends
      await Promise.all([
        fetchHealthStatus(),
        fetchAllBackends(modelsList)
      ]);
    } catch {
      messageApi.error('获取数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  useEffect(() => {
    // Auto refresh health status every 30 seconds
    const interval = setInterval(() => {
      fetchHealthStatus();
    }, 30000);
    return () => clearInterval(interval);
  }, []);

  const healthColumns = [
    {
      title: '后端ID',
      dataIndex: 'id',
      key: 'id',
      ellipsis: true,
    },
    {
      title: '所属模型',
      dataIndex: 'model_id',
      key: 'model_id',
      render: (modelId: string) => {
        const model = models.find(m => m.id === modelId);
        return model ? model.name : modelId;
      },
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => name || '-',
    },
    {
      title: 'URL',
      dataIndex: 'base_url',
      key: 'base_url',
      ellipsis: true,
    },
    {
      title: '健康状态',
      key: 'health',
      render: (_: unknown, record: Backend) => {
        if (record.enabled === false) {
          return <Tag color="default">已禁用</Tag>;
        }

        const health = healthStatus[record.id];
        if (!health) {
          return <Tag color="orange">检查中</Tag>;
        }
        return health.healthy ? (
          <Tag color="green" icon={<CheckCircleOutlined />}>健康</Tag>
        ) : (
          <Tag color="red" icon={<CloseCircleOutlined />}>不健康</Tag>
        );
      },
    },
    {
      title: '延迟',
      key: 'latency',
      render: (_: unknown, record: Backend) => {
        const health = healthStatus[record.id];
        if (!health || health.latency_ms === 0) return '-';
        return `${health.latency_ms}ms`;
      },
    },
    {
      title: '并发',
      key: 'concurrency',
      render: (_: unknown, record: Backend) => {
        const health = healthStatus[record.id];
        if (!health) return '-';
        const active = health.active_concurrency ?? 0;
        const max = health.max_concurrency ?? 0;
        if (max > 0) {
          const ratio = active / max;
          const color = ratio >= 0.9 ? 'red' : ratio >= 0.7 ? 'orange' : 'green';
          return <Tag color={color}>{active}/{max}</Tag>;
        }
        return active > 0 ? <Tag color="blue">{active}</Tag> : '-';
      },
    },
    {
      title: '失败次数',
      key: 'fail_count',
      render: (_: unknown, record: Backend) => {
        const health = healthStatus[record.id];
        if (!health || health.fail_count === 0) return '-';
        return <Tag color="red">{health.fail_count}</Tag>;
      },
    },
    {
      title: '最后检查',
      key: 'last_check',
      render: (_: unknown, record: Backend) => {
        const health = healthStatus[record.id];
        if (!health || !health.last_check) return '-';
        const lastCheckVal = health.last_check as any;
        const dateStr = typeof lastCheckVal === 'string' ? lastCheckVal : (lastCheckVal.Time || '');
        if (!dateStr || dateStr.startsWith('0001-01-01')) return '-';
        const d = new Date(dateStr);
        return isNaN(d.getTime()) ? '-' : d.toLocaleString();
      },
    },
  ];

  const healthyCount = Object.values(healthStatus).filter(h => h.healthy).length;
  const unhealthyCount = Object.values(healthStatus).filter(h => !h.healthy).length;
  const totalBackends = Object.keys(healthStatus).length;

  return (
    <div>
      {contextHolder}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <Statistic
              title="总后端数"
              value={totalBackends}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="健康后端"
              value={healthyCount}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="不健康后端"
              value={unhealthyCount}
              valueStyle={{ color: unhealthyCount > 0 ? '#cf1322' : '#999' }}
              prefix={<CloseCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>
      <div style={{ marginBottom: 16 }}>
        <Button
          icon={<ReloadOutlined />}
          onClick={() => {
            fetchHealthStatus();
            fetchAllBackends(models);
          }}
        >
          刷新
        </Button>
      </div>
      <Table
        dataSource={backends}
        columns={healthColumns}
        rowKey="id"
        loading={loading}
      />
    </div>
  );
};

export default HealthTab;
