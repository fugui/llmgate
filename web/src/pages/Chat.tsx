import React, { useState, useRef, useEffect } from 'react';
import { Input, Button, Select, Card, message, Space, Tag } from 'antd';
import { SendOutlined, ClearOutlined } from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import api from '../api';

const { TextArea } = Input;
const { Option } = Select;

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  model?: string;
}

interface Model {
  id: string;
  name: string;
}

const Chat: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [selectedModel, setSelectedModel] = useState<string>('');
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 获取可用模型列表
  useEffect(() => {
    const fetchModels = async () => {
      try {
        const res = await api.get('/api/v1/user/quota');
        const modelList = res.data.data?.models || [];
        setModels(modelList.map((m: string) => ({ id: m, name: m })));
        if (modelList.length > 0) {
          setSelectedModel(modelList[0]);
        }
      } catch (err) {
        message.error('获取模型列表失败');
      }
    };
    fetchModels();
  }, []);

  // 滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim() || !selectedModel) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input.trim(),
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setLoading(true);

    try {
      // 获取用户的 API Key
      const keysRes = await api.get('/api/v1/user/keys');
      const keys = keysRes.data.data || [];
      const activeKey = keys.find((k: any) => k.enabled);

      if (!activeKey) {
        message.error('没有可用的 API Key，请先创建一个');
        setLoading(false);
        return;
      }

      // 调用聊天接口
      const response = await fetch('/v1/chat/completions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${activeKey.key_prefix}`,
        },
        body: JSON.stringify({
          model: selectedModel,
          messages: [
            ...messages.map(m => ({ role: m.role, content: m.content })),
            { role: 'user', content: userMessage.content }
          ],
          stream: true,
        }),
      });

      if (!response.ok) {
        throw new Error('请求失败');
      }

      // 处理流式响应
      const reader = response.body?.getReader();
      const decoder = new TextDecoder();
      let assistantContent = '';

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: '',
        timestamp: new Date(),
        model: selectedModel,
      };

      setMessages(prev => [...prev, assistantMessage]);

      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n');

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') continue;

            try {
              const parsed = JSON.parse(data);
              const content = parsed.choices?.[0]?.delta?.content || '';
              assistantContent += content;

              setMessages(prev =>
                prev.map(m =>
                  m.id === assistantMessage.id
                    ? { ...m, content: assistantContent }
                    : m
                )
              );
            } catch (e) {
              // 忽略解析错误
            }
          }
        }
      }
    } catch (err: any) {
      message.error(err.message || '发送消息失败');
    } finally {
      setLoading(false);
    }
  };

  const handleClear = () => {
    setMessages([]);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', minHeight: 500 }}>
      {/* 模型选择 */}
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <span>选择模型：</span>
          <Select
            value={selectedModel}
            onChange={setSelectedModel}
            style={{ width: 200 }}
            disabled={models.length === 0}
          >
            {models.map(model => (
              <Option key={model.id} value={model.id}>{model.name}</Option>
            ))}
          </Select>
        </Space>

        <Button
          icon={<ClearOutlined />}
          onClick={handleClear}
          disabled={messages.length === 0}
        >
          清空对话
        </Button>
      </div>

      {/* 消息列表 */}
      <div style={{
        flex: 1,
        overflow: 'auto',
        border: '1px solid #f0f0f0',
        borderRadius: 8,
        padding: 16,
        marginBottom: 16,
        background: '#fafafa',
      }}>
        {messages.length === 0 ? (
          <div style={{
            height: '100%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#999',
            fontSize: 16,
          }}>
            开始与 AI 助手对话
          </div>
        ) : (
          messages.map(msg => (
            <div
              key={msg.id}
              style={{
                marginBottom: 16,
                display: 'flex',
                justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
              }}
            >
              <Card
                size="small"
                style={{
                  maxWidth: '80%',
                  background: msg.role === 'user' ? '#1890ff' : '#fff',
                  color: msg.role === 'user' ? '#fff' : 'inherit',
                }}
                bodyStyle={{ padding: 12 }}
              >
                <div style={{ marginBottom: 4 }}>
                  <Tag color={msg.role === 'user' ? 'blue' : 'green'}>
                    {msg.role === 'user' ? '我' : 'AI'}
                  </Tag>
                  {msg.model && (
                    <Tag style={{ marginLeft: 8, fontSize: 12 }}>
                      {msg.model}
                    </Tag>
                  )}
                </div>
                <div style={{ fontSize: 14, lineHeight: 1.6 }}>
                  {msg.role === 'assistant' ? (
                    <ReactMarkdown>{msg.content}</ReactMarkdown>
                  ) : (
                    msg.content
                  )}
                </div>
              </Card>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* 输入框 */}
      <div style={{ display: 'flex', gap: 12 }}>
        <TextArea
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="输入消息，按 Enter 发送，Shift+Enter 换行..."
          autoSize={{ minRows: 2, maxRows: 6 }}
          style={{ flex: 1 }}
          disabled={loading}
        />
        <Button
          type="primary"
          icon={<SendOutlined />}
          onClick={handleSend}
          loading={loading}
          disabled={!input.trim() || !selectedModel}
          style={{ height: 'auto' }}
        >
          发送
        </Button>
      </div>
    </div>
  );
};

export default Chat;
