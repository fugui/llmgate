import React, { useEffect, useState } from 'react';
import { Card, Descriptions, Table, Tag, Statistic, Row, Col, Progress } from 'antd';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import api from '../api';

const UsageStats: React.FC = () => {
  const [quota, setQuota] = useState<any>({});
  const [usageRecords, setUsageRecords] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [quotaRes, usageRes] = await Promise.all([
        api.get('/api/v1/user/quota'),
        api.get('/api/v1/user/usage'),
      ]);

      setQuota(quotaRes.data.data || {});
      setUsageRecords(usageRes.data.data || []);
    } catch (err) {
      console.error('Failed to fetch usage data:', err);
    } finally {
      setLoading(false);
    }
  };

  // 模拟最近7天的使用数据
  const weeklyData = [
    { date: '周一', requests: 45, tokens: 12000 },
    { date: '周二', requests: 52, tokens: 15000 },
    { date: '周三', requests: 38, tokens: 9800 },
    { date: '周四', requests: 65, tokens: 21000 },
    { date: '周五', requests: 48, tokens: 13500 },
    { date: '周六', requests: 25, tokens: 6500 },
    { date: '周日', requests: 30, tokens: 8200 },
  ];

  const columns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (text: string) => new Date(text).toLocaleString(),
    },
    {
      title: '模型',
      dataIndex: 'model_id',
      key: 'model_id',
    },
    {
      title: '输入 Token',
      dataIndex: 'input_tokens',
      key: 'input_tokens',
    },
    {
      title: '输出 Token',
      dataIndex: 'output_tokens',
      key: 'output_tokens',
    },
    {
      title: '耗时',
      dataIndex: 'latency_ms',
      key: 'latency_ms',
      render: (ms: number) => `${ms}ms`,
    },
    {
      title: '状态',
      dataIndex: 'status_code',
      key: 'status_code',
      render: (code: number) => (
        <Tag color={code === 200 ? 'green' : 'red'}>
          {code === 200 ? '成功' : '失败'}
        </Tag>
      ),
    },
  ];

  const tokenUsagePercent = quota.token_limit
    ? Math.round((quota.token_used / quota.token_limit) * 100)
    : 0;

  const requestUsagePercent = quota.rate_limit
    ? Math.round(((quota.requests_used || 0) / quota.rate_limit) * 100)
    : 0;

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>使用统计</h2>

      {/* 配额概览 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <Statistic
              title="今日 Token 使用量"
              value={quota.token_used || 0}
              suffix={`/ ${quota.token_limit || 0}`}
            />
            <Progress
              percent={tokenUsagePercent}
              status={tokenUsagePercent > 90 ? 'exception' : 'active'}
              style={{ marginTop: 8 }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="今日请求次数"
              value={quota.requests_used || 0}
              suffix={`/ ${quota.rate_limit || 0}`}
            />
            <Progress
              percent={requestUsagePercent}
              status={requestUsagePercent > 90 ? 'exception' : 'active'}
              style={{ marginTop: 8 }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="可用模型数"
              value={quota.models?.length || 0}
              suffix="个"
            />
            <div style={{ marginTop: 8 }}>
              {quota.models?.map((model: string) => (
                <Tag key={model} style={{ margin: '0 4px 4px 0' }}>
                  {model}
                </Tag>
              ))}
            </div>
          </Card>
        </Col>
      </Row>

      {/* 配额详情 */}
      <Card title="配额详情" style={{ marginBottom: 24 }}>
        <Descriptions bordered column={2}>
          <Descriptions.Item label="速率限制">
            {quota.rate_limit} 请求/分钟
          </Descriptions.Item>
          <Descriptions.Item label="Token 日限额">
            {quota.token_quota_daily?.toLocaleString() || '无限制'}
          </Descriptions.Item>
          <Descriptions.Item label="Token 月限额">
            {quota.token_quota_monthly?.toLocaleString() || '无限制'}
          </Descriptions.Item>
          <Descriptions.Item label="重置时间">
            {quota.reset_time || '每日 00:00'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 使用趋势图 */}
      <Card title="最近7天使用趋势" style={{ marginBottom: 24 }}>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={weeklyData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis yAxisId="left" />
            <YAxis yAxisId="right" orientation="right" />
            <Tooltip />
            <Bar yAxisId="left" dataKey="requests" name="请求数" fill="#1890ff" />
            <Bar yAxisId="right" dataKey="tokens" name="Token 数" fill="#52c41a" />
          </BarChart>
        </ResponsiveContainer>
      </Card>

      {/* 最近使用记录 */}
      <Card title="最近使用记录">
        <Table
          dataSource={usageRecords.slice(0, 10)}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={false}
        />
      </Card>
    </div>
  );
};

export default UsageStats;
