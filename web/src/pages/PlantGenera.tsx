import React, { useState } from 'react';
import { Typography, Table, Tag, Space, Button, Modal, Form, Input, message, Alert, Spin } from 'antd';
import { PlusOutlined, ReloadOutlined, ClusterOutlined } from '@ant-design/icons';
import { useAsync } from '../hooks/useAsync';
import { useNavigate } from 'react-router-dom';
import { api, type PlantGenus, type PlantSpecies } from '../api';

const { Title, Text } = Typography;

const PlantGenera: React.FC = () => {
  const navigate = useNavigate();
  const genera = useAsync<PlantGenus[]>(() => api.listPlantGenera(), []);
  const species = useAsync<PlantSpecies[]>(() => api.listPlantSpecies(), []);

  const [modalOpen, setModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm<{ name: string }>();

  const allGenera = genera.data ?? [];
  const allSpecies = species.data ?? [];
  const loading = genera.loading || species.loading;
  const error = genera.error || species.error;

  const speciesByGenus = new Map<number, number>();
  allSpecies.forEach((s) => {
    if (s.PlantGenusID.Valid) {
      const gid = s.PlantGenusID.Int64;
      speciesByGenus.set(gid, (speciesByGenus.get(gid) ?? 0) + 1);
    }
  });

  const submit = (values: { name: string }) => {
    setSubmitting(true);
    api
      .createPlantGenus(values.name)
      .then(() => {
        message.success('Genus created');
        setModalOpen(false);
        form.resetFields();
        genera.reload();
      })
      .catch((e) => message.error(e instanceof Error ? e.message : String(e)))
      .finally(() => setSubmitting(false));
  };

  const columns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 70 },
    {
      title: 'Name',
      key: 'name',
      render: (_: unknown, r: PlantGenus) => (
        <a onClick={() => navigate(`/plant-genera/${r.ID}`)}>{r.Name}</a>
      ),
    },
    {
      title: 'Species',
      key: 'species',
      render: (_: unknown, r: PlantGenus) => (
        <Tag color="green">{speciesByGenus.get(r.ID) ?? 0}</Tag>
      ),
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>Plant Genera</Title>
          <Text type="secondary">Genera group related plant species.</Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => { genera.reload(); species.reload(); }}>
            Refresh
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            New Genus
          </Button>
        </Space>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Table rowKey="ID" columns={columns} dataSource={allGenera} pagination={false} />
      </Spin>

      <Modal
        title="New Genus"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        okText="Create"
        confirmLoading={submitting}
      >
        <Form form={form} layout="vertical" onFinish={submit}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input prefix={<ClusterOutlined />} placeholder="Genus name" />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default PlantGenera;