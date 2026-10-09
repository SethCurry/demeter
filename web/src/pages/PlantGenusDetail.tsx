import React, { useState } from 'react';
import {
  Card,
  Typography,
  Table,
  Tag,
  Space,
  Button,
  Alert,
  Spin,
  Descriptions,
  Modal,
  Form,
  Input,
  message,
  Statistic,
  Row,
  Col,
} from 'antd';
import {
  ReloadOutlined,
  RollbackOutlined,
  ClusterOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { useAsync } from '../hooks/useAsync';
import { api, type PlantGenus, type PlantSpecies, type Plant } from '../api';

const { Title, Text } = Typography;

const PlantGenusDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const genusId = Number(id);

  const validId = Number.isFinite(genusId) && genusId > 0;

  const gen = useAsync<PlantGenus>(
    () => (validId ? api.getPlantGenus(genusId) : Promise.reject(new Error('invalid id'))),
    [genusId]
  );
  const species = useAsync<PlantSpecies[]>(
    () => (validId ? api.listPlantSpecies(genusId) : Promise.resolve([])),
    [genusId]
  );
  const plants = useAsync<Plant[]>(() => api.listPlants(), []);

  const [addModalOpen, setAddModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm<{ name: string }>();

  const genus = gen.data;
  const genusSpecies = species.data ?? [];
  const allPlants = plants.data ?? [];

  const loading = gen.loading || species.loading || plants.loading;
  const error = gen.error || species.error || plants.error;

  const speciesById = new Map<number, PlantSpecies>(genusSpecies.map((s) => [s.ID, s]));
  const plantsBySpecies = new Map<number, number>();
  allPlants.forEach((p) => {
    if (p.SpeciesID.Valid) {
      const sid = p.SpeciesID.Int64;
      if (speciesById.has(sid)) {
        plantsBySpecies.set(sid, (plantsBySpecies.get(sid) ?? 0) + 1);
      }
    }
  });

  const reloadAll = () => {
    gen.reload();
    species.reload();
    plants.reload();
  };

  const addSpecies = (values: { name: string }) => {
    setSubmitting(true);
    api
      .createPlantSpecies(values.name, genusId)
      .then(() => {
        message.success('Species added');
        setAddModalOpen(false);
        form.resetFields();
        species.reload();
      })
      .catch((e) => message.error(e instanceof Error ? e.message : String(e)))
      .finally(() => setSubmitting(false));
  };

  if (!validId) {
    return (
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Alert type="error" message="Invalid genus id" showIcon />
        <Button icon={<RollbackOutlined />} onClick={() => navigate('/plant-genera')}>
          Back to genera
        </Button>
      </Space>
    );
  }

  if (gen.error) {
    return (
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Alert type="error" message="Failed to load genus" description={gen.error} showIcon />
        <Button icon={<RollbackOutlined />} onClick={() => navigate('/plant-genera')}>
          Back to genera
        </Button>
      </Space>
    );
  }

  const speciesColumns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 60 },
    {
      title: 'Name',
      key: 'name',
      render: (_: unknown, r: PlantSpecies) => (
        <a onClick={() => navigate(`/plant-species/${r.ID}`)}>{r.Name}</a>
      ),
    },
    {
      title: 'Plants',
      key: 'plants',
      render: (_: unknown, r: PlantSpecies) => (
        <Tag color="green">{plantsBySpecies.get(r.ID) ?? 0}</Tag>
      ),
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Space align="center">
            <Button
              icon={<RollbackOutlined />}
              onClick={() => navigate('/plant-genera')}
              type="text"
            />
            <Title level={3} style={{ marginBottom: 4 }}>
              <ClusterOutlined /> {genus ? genus.Name : `Genus #${genusId}`}
            </Title>
          </Space>
          <Text type="secondary">Species classified under this genus.</Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={reloadAll}>
            Refresh
          </Button>
          <Button icon={<PlusOutlined />} onClick={() => setAddModalOpen(true)}>
            Add Species
          </Button>
        </Space>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Card size="small">
            <Descriptions title="Genus" column={{ xs: 1, sm: 2, md: 3 }} size="small">
              <Descriptions.Item label="ID">{genus?.ID ?? '—'}</Descriptions.Item>
              <Descriptions.Item label="Name">{genus?.Name ?? '—'}</Descriptions.Item>
            </Descriptions>
          </Card>

          <Row gutter={[16, 16]}>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="Species in Genus"
                  value={genusSpecies.length}
                  prefix={<ClusterOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="Plants of this Genus"
                  value={genusSpecies.reduce((n, s) => n + (plantsBySpecies.get(s.ID) ?? 0), 0)}
                />
              </Card>
            </Col>
          </Row>

          <Card size="small" title={`Species in ${genus?.Name ?? 'Genus'} (${genusSpecies.length})`}>
            <Table
              rowKey="ID"
              columns={speciesColumns}
              dataSource={genusSpecies}
              pagination={false}
              size="small"
              locale={{ emptyText: 'No species in this genus' }}
            />
          </Card>
        </Space>
      </Spin>

      <Modal
        title={`Add Species to ${genus?.Name ?? 'Genus'}`}
        open={addModalOpen}
        onCancel={() => setAddModalOpen(false)}
        onOk={() => form.submit()}
        okText="Create"
        confirmLoading={submitting}
      >
        <Form form={form} layout="vertical" onFinish={addSpecies}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="Species name" />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default PlantGenusDetail;