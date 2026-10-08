import React, { useState } from 'react';
import {
  Typography,
  Table,
  Tag,
  Space,
  Button,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  message,
  Switch,
  Card,
} from 'antd';
import { PlusOutlined, PlayCircleOutlined, ReloadOutlined } from '@ant-design/icons';
import { useAsync } from '../hooks/useAsync';
import { useNavigate } from 'react-router-dom';
import { api, type System } from '../api';

const { Title, Text } = Typography;

interface Task {
  id: number;
  name: string;
  systemId: number;
  type: 'top_up' | 'drain';
  amountL: number;
  enabled: boolean;
  lastRun: string | null;
}

const initialTasks: Task[] = [
  { id: 1, name: 'Lettuce top-up', systemId: 1, type: 'top_up', amountL: 5, enabled: true, lastRun: '2025-10-03 09:00' },
  { id: 2, name: 'Tomato weekly drain', systemId: 2, type: 'drain', amountL: 40, enabled: true, lastRun: '2025-09-29 18:00' },
];

const Maintenance: React.FC = () => {
  const navigate = useNavigate();
  const sys = useAsync<System[]>(() => api.listSystems(), []);
  const systems = sys.data ?? [];

  const [tasks, setTasks] = useState<Task[]>(initialTasks);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const addTask = (values: Omit<Task, 'id' | 'lastRun'>) => {
    setTasks((prev) => [...prev, { ...values, id: prev.length + 1, lastRun: null }]);
    setModalOpen(false);
    form.resetFields();
    message.success('Maintenance task created');
  };

  const toggle = (id: number, enabled: boolean) => {
    setTasks((prev) => prev.map((t) => (t.id === id ? { ...t, enabled } : t)));
  };

  const runNow = (task: Task) => {
    setTasks((prev) =>
      prev.map((t) => (t.id === task.id ? { ...t, lastRun: new Date().toISOString().slice(0, 16).replace('T', ' ') } : t))
    );
    message.success(`Ran "${task.name}"`);
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    {
      title: 'System',
      key: 'system',
      render: (_: unknown, r: Task) => {
        const s = systems.find((s) => s.ID === r.systemId);
        return s ? <a onClick={() => navigate(`/systems/${s.ID}`)}>{s.Name}</a> : '—';
      },
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: Task['type']) => (
        <Tag color={type === 'drain' ? 'volcano' : 'cyan'}>{type === 'drain' ? 'Drain' : 'Top up'}</Tag>
      ),
    },
    { title: 'Amount', key: 'amountL', render: (_: unknown, r: Task) => `${r.amountL} L` },
    {
      title: 'Enabled',
      key: 'enabled',
      render: (_: unknown, r: Task) => <Switch checked={r.enabled} onChange={(v) => toggle(r.id, v)} />,
    },
    { title: 'Last run', key: 'lastRun', render: (_: unknown, r: Task) => r.lastRun ?? '—' },
    {
      title: '',
      key: 'run',
      render: (_: unknown, r: Task) => (
        <Button size="small" icon={<PlayCircleOutlined />} onClick={() => runNow(r)}>Run now</Button>
      ),
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>Maintenance</Title>
          <Text type="secondary">Automated top-ups and routine drains.</Text>
        </div>
        <Button icon={<ReloadOutlined />} onClick={() => sys.reload()}>Refresh</Button>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          New Task
        </Button>
      </div>

      <Card>
        <Text type="secondary">
          Automated maintenance tops up water and performs routine drains to keep reservoirs balanced.
        </Text>
      </Card>

      <Table rowKey="id" columns={columns} dataSource={tasks} pagination={false} />

      <Modal
        title="New Maintenance Task"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        okText="Create"
      >
        <Form form={form} layout="vertical" onFinish={addTask} initialValues={{ type: 'top_up', amountL: 5, enabled: true }}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="Task name" />
          </Form.Item>
          <Form.Item name="systemId" label="System" rules={[{ required: true }]}>
            <Select
              options={systems.map((s) => ({ value: s.ID, label: s.Name }))}
              placeholder="Select system"
            />
          </Form.Item>
          <Form.Item name="type" label="Type">
            <Select
              options={[
                { value: 'top_up', label: 'Top up' },
                { value: 'drain', label: 'Drain' },
              ]}
            />
          </Form.Item>
          <Form.Item name="amountL" label="Amount (L)">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="enabled" label="Enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default Maintenance;