import React, { useState } from 'react';
import { Typography, Table, Tag, Space, Button, Modal, Form, Input, Select, message, Alert, Spin } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAsync } from '../hooks/useAsync';
import { api, type Enclosure, type System, type Flow } from '../api';

const { Title, Text } = Typography;

const Systems: React.FC = () => {
  const enc = useAsync<Enclosure[]>(() => api.listEnclosures(), []);
  const sys = useAsync<System[]>(() => api.listSystems(), []);
  const flw = useAsync<Flow[]>(() => api.listFlows(), []);

  const navigate = useNavigate();
  const [modalOpen, setModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm<{ name: string; enclosureId: number }>();

  const enclosures = enc.data ?? [];
  const systems = sys.data ?? [];
  const flows = flw.data ?? [];
  const loading = enc.loading || sys.loading || flw.loading;
  const error = enc.error || sys.error || flw.error;

  const enclosureById = new Map(enclosures.map((e) => [e.ID, e]));
  const flowsBySystem = new Map<number, number>();
  flows.forEach((f) => flowsBySystem.set(f.SystemID, (flowsBySystem.get(f.SystemID) ?? 0) + 1));

  const submit = (values: { name: string; enclosureId: number }) => {
    setSubmitting(true);
    api
      .createSystem(values.name, values.enclosureId)
      .then(() => {
        message.success('System created');
        setModalOpen(false);
        form.resetFields();
        sys.reload();
        flw.reload();
      })
      .catch((e) => message.error(e instanceof Error ? e.message : String(e)))
      .finally(() => setSubmitting(false));
  };

  const columns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 60 },
    {
      title: 'Name',
      dataIndex: 'Name',
      key: 'name',
      render: (name: string, r: System) => (
        <a onClick={() => navigate(`/systems/${r.ID}`)}>{name}</a>
      ),
    },
    {
      title: 'Enclosure',
      key: 'enclosure',
      render: (_: unknown, r: System) => {
        const encl = enclosureById.get(r.EnclosureID);
        return encl ? (
          <a onClick={() => navigate(`/enclosures/${encl.ID}`)}>{encl.Name}</a>
        ) : (
          '—'
        );
      },
    },
    {
      title: 'Flows',
      key: 'flows',
      render: (_: unknown, r: System) => <Tag color="green">{flowsBySystem.get(r.ID) ?? 0}</Tag>,
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>Systems</Title>
          <Text type="secondary">Systems are defined by sharing a reservoir.</Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => { enc.reload(); sys.reload(); flw.reload(); }}>
            Refresh
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            New System
          </Button>
        </Space>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Table rowKey="ID" columns={columns} dataSource={systems} pagination={false} />
      </Spin>

      <Modal
        title="New System"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        okText="Create"
        confirmLoading={submitting}
      >
        <Form form={form} layout="vertical" onFinish={submit}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="System name" />
          </Form.Item>
          <Form.Item name="enclosureId" label="Enclosure" rules={[{ required: true }]}>
            <Select
              options={enclosures.map((e) => ({ value: e.ID, label: e.Name }))}
              placeholder="Select enclosure"
            />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default Systems;