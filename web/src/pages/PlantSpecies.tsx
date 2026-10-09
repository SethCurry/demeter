import React, { useState } from 'react';
import { Typography, Table, Tag, Space, Button, Modal, Form, Input, Select, message, Alert, Spin } from 'antd';
import { PlusOutlined, ReloadOutlined, TagsOutlined } from '@ant-design/icons';
import { useAsync } from '../hooks/useAsync';
import { useNavigate } from 'react-router-dom';
import { api, nullInt, type PlantSpecies as PlantSpeciesModel, type PlantGenus, type Plant } from '../api';

const { Title, Text } = Typography;

const PlantSpecies: React.FC = () => {
  const navigate = useNavigate();
  const species = useAsync<PlantSpeciesModel[]>(() => api.listPlantSpecies(), []);
  const genera = useAsync<PlantGenus[]>(() => api.listPlantGenera(), []);
  const plants = useAsync<Plant[]>(() => api.listPlants(), []);

  const [modalOpen, setModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm<{ name: string; genusId?: number }>();

  const allSpecies = species.data ?? [];
  const allGenera = genera.data ?? [];
  const allPlants = plants.data ?? [];
  const loading = species.loading || genera.loading || plants.loading;
  const error = species.error || genera.error || plants.error;

  const genusById = new Map(allGenera.map((g) => [g.ID, g]));
  const plantsBySpecies = new Map<number, number>();
  allPlants.forEach((p) => {
    if (p.SpeciesID.Valid) {
      const sid = p.SpeciesID.Int64;
      plantsBySpecies.set(sid, (plantsBySpecies.get(sid) ?? 0) + 1);
    }
  });

  // Renders a species as "Genus species" when the genus is known, or the
  // bare species name otherwise (mirrors botanical binomial nomenclature).
  const speciesLabel = (s: PlantSpeciesModel) => {
    const gid = nullInt(s.PlantGenusID);
    const genus = gid != null ? genusById.get(gid) : undefined;
    return genus ? `${genus.Name} ${s.Name}` : s.Name;
  };

  const submit = (values: { name: string; genusId?: number }) => {
    setSubmitting(true);
    api
      .createPlantSpecies(values.name, values.genusId)
      .then(() => {
        message.success('Species created');
        setModalOpen(false);
        form.resetFields();
        species.reload();
      })
      .catch((e) => message.error(e instanceof Error ? e.message : String(e)))
      .finally(() => setSubmitting(false));
  };

  const columns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 70 },
    {
      title: 'Name',
      key: 'name',
      render: (_: unknown, r: PlantSpeciesModel) => (
        <a onClick={() => navigate(`/plant-species/${r.ID}`)}>{speciesLabel(r)}</a>
      ),
    },
    {
      title: 'Genus',
      key: 'genus',
      render: (_: unknown, r: PlantSpeciesModel) => {
        const gid = nullInt(r.PlantGenusID);
        if (gid == null) return <Tag>unclassified</Tag>;
        const genus = genusById.get(gid);
        return genus ? (
          <a onClick={() => navigate(`/plant-genera/${genus.ID}`)}>{genus.Name}</a>
        ) : (
          `#${gid}`
        );
      },
    },
    {
      title: 'Plants',
      key: 'plants',
      render: (_: unknown, r: PlantSpeciesModel) => (
        <Tag color="green">{plantsBySpecies.get(r.ID) ?? 0}</Tag>
      ),
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>Plant Species</Title>
          <Text type="secondary">Species identify the kind of plant at a site.</Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => { species.reload(); genera.reload(); plants.reload(); }}>
            Refresh
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            New Species
          </Button>
        </Space>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Table rowKey="ID" columns={columns} dataSource={allSpecies} pagination={false} />
      </Spin>

      <Modal
        title="New Species"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        okText="Create"
        confirmLoading={submitting}
      >
        <Form form={form} layout="vertical" onFinish={submit}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input prefix={<TagsOutlined />} placeholder="Species name" />
          </Form.Item>
          <Form.Item name="genusId" label="Genus">
            <Select
              allowClear
              placeholder="Optional genus"
              options={allGenera.map((g) => ({ value: g.ID, label: g.Name }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default PlantSpecies;