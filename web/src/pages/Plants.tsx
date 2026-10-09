import React, { useState } from 'react';
import { Typography, Table, Tag, Space, Button, Modal, Form, Select, InputNumber, DatePicker, message, Alert, Spin } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { useAsync } from '../hooks/useAsync';
import { useNavigate } from 'react-router-dom';
import { api, nullTime, nullInt, type PlantSite, type Flow, type Plant, type PlantSpecies, type PlantGenus } from '../api';

const { Title, Text } = Typography;

const Plants: React.FC = () => {
  const navigate = useNavigate();
  const sites = useAsync<PlantSite[]>(() => api.listPlantSites(), []);
  const flw = useAsync<Flow[]>(() => api.listFlows(), []);
  const plants = useAsync<Plant[]>(() => api.listPlants(), []);
  const spec = useAsync<PlantSpecies[]>(() => api.listPlantSpecies(), []);
  const gen = useAsync<PlantGenus[]>(() => api.listPlantGenera(), []);

  const [plantModalOpen, setPlantModalOpen] = useState(false);
  const [siteModalOpen, setSiteModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [plantForm] = Form.useForm<{ plantSiteId: number; plantedOn?: dayjs.Dayjs; speciesId?: number }>();
  const [siteForm] = Form.useForm<{ flowId: number; x: number; y: number; z: number }>();

  const plantSites = sites.data ?? [];
  const flows = flw.data ?? [];
  const allPlants = plants.data ?? [];
  const allSpecies = spec.data ?? [];
  const allGenera = gen.data ?? [];
  const loading = sites.loading || flw.loading || plants.loading || spec.loading || gen.loading;
  const error = sites.error || flw.error || plants.error || spec.error || gen.error;

  const flowById = new Map(flows.map((f) => [f.ID, f]));
  const genusById = new Map(allGenera.map((g) => [g.ID, g]));
  const speciesById = new Map(allSpecies.map((s) => [s.ID, s]));
  const plantBySite = new Map<number, Plant>();
  allPlants.forEach((p) => plantBySite.set(p.PlantSiteID, p));

  // Renders a species as "Genus species" when the genus is known.
  const speciesLabel = (s: PlantSpecies) => {
    const gid = nullInt(s.PlantGenusID);
    const genus = gid != null ? genusById.get(gid) : undefined;
    return genus ? `${genus.Name} ${s.Name}` : s.Name;
  };

  const createPlant = (values: { plantSiteId: number; plantedOn?: dayjs.Dayjs; speciesId?: number }) => {
    setSubmitting(true);
    api
      .createPlant(
        values.plantSiteId,
        values.plantedOn ? values.plantedOn.format('YYYY-MM-DDTHH:mm:ssZ') : undefined,
        values.speciesId,
      )
      .then(() => {
        message.success('Plant added');
        setPlantModalOpen(false);
        plantForm.resetFields();
        plants.reload();
      })
      .catch((e) => message.error(e instanceof Error ? e.message : String(e)))
      .finally(() => setSubmitting(false));
  };

  const createSite = (values: { flowId: number; x: number; y: number; z: number }) => {
    setSubmitting(true);
    api
      .createPlantSite(values.flowId, values.x, values.y, values.z)
      .then(() => {
        message.success('Plant site created');
        setSiteModalOpen(false);
        siteForm.resetFields();
        sites.reload();
      })
      .catch((e) => message.error(e instanceof Error ? e.message : String(e)))
      .finally(() => setSubmitting(false));
  };

  const columns = [
    { title: 'Site ID', dataIndex: 'ID', key: 'id', width: 70 },
    {
      title: 'Flow',
      key: 'flow',
      render: (_: unknown, r: PlantSite) => {
        const f = flowById.get(r.FlowID);
        return f ? <a onClick={() => navigate(`/flows/${f.ID}`)}>{f.Name}</a> : '—';
      },
    },
    { title: 'X', dataIndex: 'X', key: 'x', width: 60 },
    { title: 'Y', dataIndex: 'Y', key: 'y', width: 60 },
    { title: 'Z', dataIndex: 'Z', key: 'z', width: 60 },
    {
      title: 'Plant',
      key: 'plant',
      render: (_: unknown, r: PlantSite) => {
        const plant = plantBySite.get(r.ID);
        if (!plant) return <Tag>empty</Tag>;
        return (
          <Button type="link" style={{ padding: 0 }} onClick={() => navigate(`/plants/${plant.ID}`)}>
            <Tag color="green" style={{ cursor: 'pointer' }}>#{plant.ID}</Tag>
          </Button>
        );
      },
    },
    {
      title: 'Species',
      key: 'species',
      render: (_: unknown, r: PlantSite) => {
        const plant = plantBySite.get(r.ID);
        const sid = plant ? nullInt(plant.SpeciesID) : null;
        const species = sid != null ? speciesById.get(sid) : undefined;
        if (!species) return '—';
        return (
          <a onClick={() => navigate(`/plant-species/${species.ID}`)}>{speciesLabel(species)}</a>
        );
      },
    },
    {
      title: 'Planted',
      key: 'plantedOn',
      render: (_: unknown, r: PlantSite) => {
        const plant = plantBySite.get(r.ID);
        const planted = plant ? nullTime(plant.PlantedOn) : null;
        return planted ? new Date(planted).toLocaleDateString() : '—';
      },
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>Plants</Title>
          <Text type="secondary">Plant sites are individual locations within a flow.</Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => { sites.reload(); flw.reload(); plants.reload(); spec.reload(); gen.reload(); }}>
            Refresh
          </Button>
          <Button icon={<PlusOutlined />} onClick={() => setSiteModalOpen(true)}>
            New Plant Site
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setPlantModalOpen(true)}>
            Add Plant
          </Button>
        </Space>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Table rowKey="ID" columns={columns} dataSource={plantSites} pagination={false} />
      </Spin>

      <Modal
        title="Add Plant"
        open={plantModalOpen}
        onCancel={() => setPlantModalOpen(false)}
        onOk={() => plantForm.submit()}
        okText="Create"
        confirmLoading={submitting}
      >
        <Form form={plantForm} layout="vertical" onFinish={createPlant}>
          <Form.Item name="plantSiteId" label="Plant Site" rules={[{ required: true }]}>
            <Select
              options={plantSites.map((s) => ({
                value: s.ID,
                label: `#${s.ID} — ${flowById.get(s.FlowID)?.Name ?? 'unknown flow'} (${s.X},${s.Y},${s.Z})`,
              }))}
              placeholder="Select plant site"
            />
          </Form.Item>
          <Form.Item name="plantedOn" label="Planted On">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="speciesId" label="Species">
            <Select
              allowClear
              options={allSpecies.map((s) => ({ value: s.ID, label: speciesLabel(s) }))}
              placeholder="Optional species"
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="New Plant Site"
        open={siteModalOpen}
        onCancel={() => setSiteModalOpen(false)}
        onOk={() => siteForm.submit()}
        okText="Create"
        confirmLoading={submitting}
      >
        <Form form={siteForm} layout="vertical" onFinish={createSite} initialValues={{ x: 0, y: 0, z: 0 }}>
          <Form.Item name="flowId" label="Flow" rules={[{ required: true }]}>
            <Select
              options={flows.map((f) => ({ value: f.ID, label: f.Name }))}
              placeholder="Select flow"
            />
          </Form.Item>
          <Space>
            <Form.Item name="x" label="X" rules={[{ required: true }]}>
              <InputNumber />
            </Form.Item>
            <Form.Item name="y" label="Y" rules={[{ required: true }]}>
              <InputNumber />
            </Form.Item>
            <Form.Item name="z" label="Z" rules={[{ required: true }]}>
              <InputNumber />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </Space>
  );
};

export default Plants;