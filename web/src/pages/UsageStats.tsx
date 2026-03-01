import React, { useEffect, useState } from 'react';
import { Card, Descriptions, Table, Tag, Statistic, Row, Col, Progress } from 'antd';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import api from '../api';

const UsageStats: React.FC = () => {
  const [quota, setQuota] = useState<any>({});
  const [usageRecords, setUsageRecords] = useState<any[]>([]);
  const [weeklyData, setWeeklyData] = useState<any[]>([]);
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
      const records = usageRes.data.data || [];
      setUsageRecords(records);
      
      // 转换数据为图表格式
      const chartData = records.map((record: any) => {
        const date = new Date(record.date);
        const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
        return {
          date: weekdays[date.getDay()],
          requests: record.requests,
          tokens: record.tokens,
        };
      }).reverse(); // 按时间正序
      setWeeklyData(chartData);
    } catch (err) {
      console.error('Failed to fetch usage data:', err);
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    {
      title: '日期',
      dataIndex: 'date',
      key: 'date',
    },
    {
      title: '模型',
      dataIndex: 'model_id',
      key: 'model_id',
      render: (_: string, record: any) => record.model_id || '-',
    },
    {
      title: '请求数',
      dataIndex: 'requests',
      key: 'requests',
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
      title: '总 Token',
      dataIndex: 'tokens',
      key: 'tokens',
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
