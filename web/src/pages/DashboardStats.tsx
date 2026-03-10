import React, { useEffect, useState } from 'react';
import { Card, Statistic, Table, Progress, Row, Col, Spin, Empty, Tag } from 'antd';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from 'recharts';
import {
  TeamOutlined,
  RiseOutlined,
  ApartmentOutlined,
  BarChartOutlined,
  ThunderboltOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
} from '@ant-design/icons';
import api from '../api';

// 类型定义
interface DashboardStats {
  today_total_requests: number;
  today_input_tokens: number;
  today_output_tokens: number;
  active_users: number;
  total_users: number;
  department_count: number;
  avg_requests_per_user: number;
}

interface TopUser {
  user_id: string;
  name: string;
  department: string;
  request_count: number;
  input_tokens: number;
  output_tokens: number;
}

interface HourlyStat {
  hour: string;
  requests: number;
}

interface DepartmentStat {
  department: string;
  user_count: number;
  request_count: number;
}

interface ModelStat {
  model_id: string;
  request_count: number;
  input_tokens: number;
  output_tokens: number;
}

// 饼图颜色配置
const PIE_COLORS = [
  '#1890ff',
  '#52c41a',
  '#faad14',
  '#f5222d',
  '#722ed1',
  '#13c2c2',
  '#eb2f96',
  '#fa541c',
];

// 格式化 Token 数量
const formatTokens = (count: number): string => {
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K`;
  return String(count);
};

const DashboardStatsPage: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [topUsers, setTopUsers] = useState<TopUser[]>([]);
  const [hourlyStats, setHourlyStats] = useState<HourlyStat[]>([]);
  const [departmentStats, setDepartmentStats] = useState<DepartmentStat[]>([]);
  const [modelStats, setModelStats] = useState<ModelStat[]>([]);

  useEffect(() => {
    fetchAllData();
  }, []);

  const fetchAllData = async () => {
    setLoading(true);
    try {
      const [statsRes, topUsersRes, hourlyRes, departmentsRes, modelsRes] =
        await Promise.all([
          api.get('/api/v1/dashboard/stats'),
          api.get('/api/v1/dashboard/top-users'),
          api.get('/api/v1/dashboard/hourly'),
          api.get('/api/v1/dashboard/departments'),
          api.get('/api/v1/dashboard/models'),
        ]);

      setStats(statsRes.data.data);
      setTopUsers(topUsersRes.data.data || []);
      setHourlyStats(hourlyRes.data.data || []);
      setDepartmentStats(departmentsRes.data.data || []);
      setModelStats(modelsRes.data.data || []);
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      setLoading(false);
    }
  };

  // TOP用户表格列
  const topUserColumns = [
    {
      title: '排名',
      key: 'rank',
      width: 55,
      render: (_: any, __: any, index: number) => (
        <span
          style={{
            fontWeight: index < 3 ? 'bold' : 'normal',
            color: index === 0 ? '#ffd700' : index === 1 ? '#c0c0c0' : index === 2 ? '#cd7f32' : '#999',
          }}
        >
          {index + 1}
        </span>
      ),
    },
    {
      title: '用户',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: '部门',
      dataIndex: 'department',
      key: 'department',
      ellipsis: true,
      render: (dept: string) => dept || '未设置',
    },
    {
      title: '请求数',
      dataIndex: 'request_count',
      key: 'request_count',
      width: 90,
      render: (count: number) => {
        const maxCount = topUsers[0]?.request_count || 1;
        const percent = (count / maxCount) * 100;
        return (
          <div>
            <div style={{ fontWeight: 'bold', marginBottom: 4 }}>{count}</div>
            <Progress
              percent={percent}
              showInfo={false}
              strokeColor="#1890ff"
              size="small"
            />
          </div>
        );
      },
    },
    {
      title: 'Tokens',
      key: 'tokens',
      width: 120,
      render: (_: any, record: TopUser) => (
        <div style={{ fontSize: 12 }}>
          <div>
            <ArrowUpOutlined style={{ color: '#52c41a', marginRight: 3 }} />
            <span style={{ color: '#52c41a' }}>{formatTokens(record.input_tokens || 0)}</span>
          </div>
          <div>
            <ArrowDownOutlined style={{ color: '#1890ff', marginRight: 3 }} />
            <span style={{ color: '#1890ff' }}>{formatTokens(record.output_tokens || 0)}</span>
          </div>
        </div>
      ),
    },
  ];

  // 部门统计表格列
  const departmentColumns = [
    {
      title: '部门',
      dataIndex: 'department',
      key: 'department',
      ellipsis: true,
    },
    {
      title: '用户数',
      dataIndex: 'user_count',
      key: 'user_count',
      width: 70,
    },
    {
      title: '请求数',
      dataIndex: 'request_count',
      key: 'request_count',
      width: 80,
    },
  ];

  // 模型统计表格列（用于展示模型 Token 汇总）
  const modelTokenColumns = [
    {
      title: '模型',
      dataIndex: 'model_id',
      key: 'model_id',
      ellipsis: true,
      render: (id: string) => <Tag color="blue">{id}</Tag>,
    },
    {
      title: '请求数',
      dataIndex: 'request_count',
      key: 'request_count',
      width: 75,
    },
    {
      title: '输入 Tokens',
      dataIndex: 'input_tokens',
      key: 'input_tokens',
      width: 100,
      render: (v: number) => (
        <span style={{ color: '#52c41a' }}>{formatTokens(v || 0)}</span>
      ),
    },
    {
      title: '输出 Tokens',
      dataIndex: 'output_tokens',
      key: 'output_tokens',
      width: 100,
      render: (v: number) => (
        <span style={{ color: '#1890ff' }}>{formatTokens(v || 0)}</span>
      ),
    },
  ];

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '60vh' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  const todayTotalTokens = (stats?.today_input_tokens || 0) + (stats?.today_output_tokens || 0);

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>
        <BarChartOutlined style={{ marginRight: 8 }} />
        数据看板
      </h2>

      {/* 系统概览卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日总请求"
              value={stats?.today_total_requests || 0}
              prefix={<RiseOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="活跃用户"
              value={stats?.active_users || 0}
              suffix={`/ ${stats?.total_users || 0}`}
              prefix={<TeamOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日总 Tokens"
              value={todayTotalTokens}
              formatter={(v) => formatTokens(Number(v))}
              prefix={<ThunderboltOutlined style={{ color: '#faad14' }} />}
              valueStyle={{ color: '#faad14' }}
            />
            <div style={{ marginTop: 8, fontSize: 12, color: '#999', display: 'flex', gap: 12 }}>
              <span>
                <ArrowUpOutlined style={{ color: '#52c41a' }} />
                {' 输入 '}{formatTokens(stats?.today_input_tokens || 0)}
              </span>
              <span>
                <ArrowDownOutlined style={{ color: '#1890ff' }} />
                {' 输出 '}{formatTokens(stats?.today_output_tokens || 0)}
              </span>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="部门数量"
              value={stats?.department_count || 0}
              prefix={<ApartmentOutlined style={{ color: '#722ed1' }} />}
              valueStyle={{ color: '#722ed1' }}
            />
            <div style={{ marginTop: 8, fontSize: 12, color: '#999' }}>
              人均请求 {(stats?.avg_requests_per_user || 0).toFixed(1)} 次
            </div>
          </Card>
        </Col>
      </Row>

      {/* 24小时趋势 + TOP10用户 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={14}>
          <Card title="最近24小时趋势">
            {hourlyStats.length > 0 && hourlyStats.some(s => s.requests > 0) ? (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={hourlyStats}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="hour" tick={{ fontSize: 12 }} />
                  <YAxis />
                  <Tooltip
                    formatter={(value: number) => [`${value} 请求`, '请求数']}
                    labelFormatter={(label: string) => `${label}`}
                  />
                  <Bar dataKey="requests" fill="#1890ff" name="请求数" />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <Empty description="暂无数据" style={{ padding: '60px 0' }} />
            )}
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card title="今日 TOP10 用户">
            {topUsers.length > 0 && topUsers.some(u => u.request_count > 0) ? (
              <Table
                dataSource={topUsers.filter(u => u.request_count > 0).slice(0, 10)}
                columns={topUserColumns}
                rowKey="user_id"
                pagination={false}
                size="small"
                scroll={{ y: 300 }}
              />
            ) : (
              <Empty description="暂无数据" style={{ padding: '60px 0' }} />
            )}
          </Card>
        </Col>
      </Row>

      {/* 模型 Token 统计 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24}>
          <Card title="今日模型 Token 消耗">
            {modelStats.length > 0 && modelStats.some(s => s.request_count > 0) ? (
              <Row gutter={16}>
                <Col xs={24} lg={12}>
                  <Table
                    dataSource={modelStats.filter(s => s.request_count > 0)}
                    columns={modelTokenColumns}
                    rowKey="model_id"
                    pagination={false}
                    size="small"
                    scroll={{ y: 250 }}
                  />
                </Col>
                <Col xs={24} lg={12}>
                  <ResponsiveContainer width="100%" height={260}>
                    <PieChart>
                      <Pie
                        data={modelStats.filter(s => (s.input_tokens || 0) + (s.output_tokens || 0) > 0).map(s => ({
                          ...s,
                          total_tokens: (s.input_tokens || 0) + (s.output_tokens || 0),
                        }))}
                        cx="50%"
                        cy="50%"
                        innerRadius={50}
                        outerRadius={90}
                        paddingAngle={2}
                        dataKey="total_tokens"
                        nameKey="model_id"
                        label={({ model_id, percent }) =>
                          `${model_id} ${(percent * 100).toFixed(0)}%`
                        }
                      >
                        {modelStats.map((_entry, index) => (
                          <Cell
                            key={`cell-${index}`}
                            fill={PIE_COLORS[index % PIE_COLORS.length]}
                          />
                        ))}
                      </Pie>
                      <Tooltip
                        formatter={(value: number, _name: string, props: any) => {
                          return [`${formatTokens(value)} Tokens`, props.payload.model_id];
                        }}
                      />
                      <Legend />
                    </PieChart>
                  </ResponsiveContainer>
                </Col>
              </Row>
            ) : (
              <Empty description="暂无数据" style={{ padding: '40px 0' }} />
            )}
          </Card>
        </Col>
      </Row>

      {/* 部门统计 + 模型请求分布 */}
      <Row gutter={16}>
        <Col xs={24} lg={12}>
          <Card title="部门使用统计">
            {departmentStats.length > 0 && departmentStats.some(s => s.request_count > 0) ? (
              <Table
                dataSource={departmentStats.filter(s => s.request_count > 0)}
                columns={departmentColumns}
                rowKey="department"
                pagination={false}
                size="small"
                scroll={{ y: 300 }}
              />
            ) : (
              <Empty description="暂无数据" style={{ padding: '60px 0' }} />
            )}
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="模型请求分布">
            {modelStats.length > 0 && modelStats.some(s => s.request_count > 0) ? (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={modelStats.filter(s => s.request_count > 0)}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={100}
                    paddingAngle={2}
                    dataKey="request_count"
                    nameKey="model_id"
                    label={({ model_id, percent }) =>
                      `${model_id} ${(percent * 100).toFixed(0)}%`
                    }
                  >
                    {modelStats
                      .filter(s => s.request_count > 0)
                      .map((_entry, index) => (
                        <Cell
                          key={`cell-${index}`}
                          fill={PIE_COLORS[index % PIE_COLORS.length]}
                        />
                      ))}
                  </Pie>
                  <Tooltip
                    formatter={(value: number, _name: string, props: any) => {
                      const total = modelStats.reduce((sum, s) => sum + s.request_count, 0);
                      const percent = total > 0 ? ((value / total) * 100).toFixed(1) : '0';
                      return [`${value} (${percent}%)`, props.payload.model_id];
                    }}
                  />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <Empty description="暂无数据" style={{ padding: '60px 0' }} />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default DashboardStatsPage;
