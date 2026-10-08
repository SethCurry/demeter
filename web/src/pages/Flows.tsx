import React, { useState } from 'react';
import { Typography, Table, Tag, Space, Button, Modal, Form, Input, Select, message, Alert, Spin } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { useAsync } from '../hooks/useAsync';
import { useNavigate } from 'react-router-dom';
import { api, nullInt, type Flow, type System, type PlantSite } from '../api';

const { Title, Text } = Typography;

const Flows: React.FC = () => {
  const navigate = useNavigate();
  const flw = useAsync<Flow[]>(() => api.listFlows(), []);
  const sys = useAsync<System[]>(() => api.listSystems(), []);
  const sites = useAsync<PlantSite[]>(() => api.listPlantSites(), []);

  const [modalOpen, setModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm<{ name: string; systemId: number; parentFlowId?: number }>();

  const flows = flw.data ?? [];
  const systems = sys.data ?? [];
  const plantSites = sites.data ?? [];
  const loading = flw.loading || sys.loading || sites.loading;
  const error = flw.error || sys.error || sites.error;

  const systemById = new Map(systems.map((s) => [s.ID, s]));
  const flowById = new Map(flows.map((f) => [f.ID, f]));
  const sitesByFlow = new Map<number, number>();
  plantSites.forEach((p) => sitesByFlow.set(p.FlowID, (sitesByFlow.get(p.FlowID) ?? 0) + 1));

  const submit = (values: { name: string; systemId: number; parentFlowId?: number }) => {
    setSubmitting(true);
    api
      .createFlow(values.name, values.systemId, values.parentFlowId)
      .then(() => {
        message.success('Flow created');
        setModalOpen(false);
        form.resetFields();
        flw.reload();
      })
      .catch((e) => message.error(e instanceof Error ? e.message : String(e)))
      .finally(() => setSubmitting(false));
  };

  const columns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 60 },
    {
      title: 'Name',
      key: 'name',
      render: (_: unknown, r: Flow) => (
        <a onClick={() => navigate(`/flows/${r.ID}`)}>{r.Name}</a>
      ),
    },
    {
      title: 'System',
      key: 'system',
      render: (_: unknown, r: Flow) => {
        const s = systemById.get(r.SystemID);
        return s ? <a onClick={() => navigate(`/systems/${s.ID}`)}>{s.Name}</a> : '—';
      },
    },
    {
      title: 'Parent Flow',
      key: 'parentFlow',
      render: (_: unknown, r: Flow) => {
        const pid = nullInt(r.ParentFlowID);
        if (pid == null) return <Tag>root</Tag>;
        const parent = flowById.get(pid);
        return parent ? (
          <a onClick={() => navigate(`/flows/${parent.ID}`)}>{parent.Name}</a>
        ) : (
          `#${pid}`
        );
      },
    },
    {
      title: 'Plant Sites',
      key: 'sites',
      render: (_: unknown, r: Flow) => <Tag color="green">{sitesByFlow.get(r.ID) ?? 0}</Tag>,
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>Flows</Title>
          <Text type="secondary">Loops share a single flow of water.</Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => { flw.reload(); sys.reload(); sites.reload(); }}>
            Refresh
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            New Flow
          </Button>
        </Space>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Table rowKey="ID" columns={columns} dataSource={flows} pagination={false} />
      </Spin>

      <Modal
        title="New Flow"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        okText="Create"
        confirmLoading={submitting}
      >
        <Form form={form} layout="vertical" onFinish={submit}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="Flow name" />
          </Form.Item>
          <Form.Item name="systemId" label="System" rules={[{ required: true }]}>
            <Select
              options={systems.map((s) => ({ value: s.ID, label: s.Name }))}
              placeholder="Select system"
            />
          </Form.Item>
          <Form.Item name="parentFlowId" label="Parent Flow">
            <Select
              allowClear
              options={flows.map((f) => ({ value: f.ID, label: f.Name }))}
              placeholder="Root flow"
            />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default Flows;