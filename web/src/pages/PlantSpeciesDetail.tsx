import React from 'react';
import {
  Card,
  Typography,
  Table,
  Space,
  Button,
  Alert,
  Spin,
  Descriptions,
  Statistic,
  Row,
  Col,
} from 'antd';
import {
  ReloadOutlined,
  RollbackOutlined,
  TagsOutlined,
  ClusterOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { useAsync } from '../hooks/useAsync';
import {
  api,
  nullInt,
  nullTime,
  type PlantSpecies,
  type PlantGenus,
  type Plant,
  type PlantSite,
  type Flow,
} from '../api';

const { Title, Text } = Typography;

const PlantSpeciesDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const speciesId = Number(id);

  const validId = Number.isFinite(speciesId) && speciesId > 0;

  const spec = useAsync<PlantSpecies>(
    () => (validId ? api.getPlantSpecies(speciesId) : Promise.reject(new Error('invalid id'))),
    [speciesId]
  );
  const genera = useAsync<PlantGenus[]>(() => api.listPlantGenera(), []);
  const plants = useAsync<Plant[]>(
    () => (validId ? api.listPlants({ speciesId }) : Promise.resolve([])),
    [speciesId]
  );
  const sites = useAsync<PlantSite[]>(() => api.listPlantSites(), []);
  const flw = useAsync<Flow[]>(() => api.listFlows(), []);

  const species = spec.data;
  const allGenera = genera.data ?? [];
  const speciesPlants = plants.data ?? [];
  const plantSites = sites.data ?? [];
  const flows = flw.data ?? [];

  const loading = spec.loading || genera.loading || plants.loading || sites.loading || flw.loading;
  const error = spec.error || genera.error || plants.error || sites.error || flw.error;

  const genusById = new Map(allGenera.map((g) => [g.ID, g]));
  const siteById = new Map(plantSites.map((s) => [s.ID, s]));
  const flowById = new Map(flows.map((f) => [f.ID, f]));

  const genusId = species ? nullInt(species.PlantGenusID) : null;
  const genus = genusId != null ? genusById.get(genusId) : undefined;

  const reloadAll = () => {
    spec.reload();
    genera.reload();
    plants.reload();
    sites.reload();
    flw.reload();
  };

  if (!validId) {
    return (
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Alert type="error" message="Invalid species id" showIcon />
        <Button icon={<RollbackOutlined />} onClick={() => navigate('/plant-species')}>
          Back to species
        </Button>
      </Space>
    );
  }

  if (spec.error) {
    return (
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Alert type="error" message="Failed to load species" description={spec.error} showIcon />
        <Button icon={<RollbackOutlined />} onClick={() => navigate('/plant-species')}>
          Back to species
        </Button>
      </Space>
    );
  }

  const fmtTime = (t: string | null) => (t ? new Date(t).toLocaleDateString() : '—');

  const plantColumns = [
    {
      title: 'Plant',
      key: 'plant',
      render: (_: unknown, r: Plant) => (
        <a onClick={() => navigate(`/plants/${r.ID}`)}>#{r.ID}</a>
      ),
    },
    {
      title: 'Site',
      key: 'site',
      render: (_: unknown, r: Plant) => {
        const site = siteById.get(r.PlantSiteID);
        return site ? `#${site.ID} (${site.X}, ${site.Y}, ${site.Z})` : `#${r.PlantSiteID}`;
      },
    },
    {
      title: 'Flow',
      key: 'flow',
      render: (_: unknown, r: Plant) => {
        const site = siteById.get(r.PlantSiteID);
        const flow = site ? flowById.get(site.FlowID) : undefined;
        return flow ? (
          <a onClick={() => navigate(`/flows/${flow.ID}`)}>{flow.Name}</a>
        ) : (
          '—'
        );
      },
    },
    {
      title: 'Planted',
      key: 'planted',
      render: (_: unknown, r: Plant) => fmtTime(nullTime(r.PlantedOn)),
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Space align="center">
            <Button
              icon={<RollbackOutlined />}
              onClick={() => navigate('/plant-species')}
              type="text"
            />
            <Title level={3} style={{ marginBottom: 4 }}>
              <TagsOutlined />{' '}
              {species ? (genus ? `${genus.Name} ${species.Name}` : species.Name) : `Species #${speciesId}`}
            </Title>
          </Space>
          <Text type="secondary">Plants grown as this species.</Text>
        </div>
        <Button icon={<ReloadOutlined />} onClick={reloadAll}>
          Refresh
        </Button>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Card size="small">
            <Descriptions title="Species" column={{ xs: 1, sm: 2, md: 3 }} size="small">
              <Descriptions.Item label="ID">{species?.ID ?? '—'}</Descriptions.Item>
              <Descriptions.Item label="Name">{species?.Name ?? '—'}</Descriptions.Item>
              <Descriptions.Item label="Genus">
                {genus ? (
                  <Button
                    type="link"
                    style={{ padding: 0 }}
                    onClick={() => navigate(`/plant-genera/${genus.ID}`)}
                  >
                    <ClusterOutlined /> {genus.Name}
                  </Button>
                ) : (
                  <Text type="secondary">unclassified</Text>
                )}
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Row gutter={[16, 16]}>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic title="Plants of Species" value={speciesPlants.length} />
              </Card>
            </Col>
          </Row>

          <Card size="small" title={`Plants (${speciesPlants.length})`}>
            <Table
              rowKey="ID"
              columns={plantColumns}
              dataSource={speciesPlants}
              pagination={false}
              size="small"
              locale={{ emptyText: 'No plants of this species' }}
            />
          </Card>
        </Space>
      </Spin>
    </Space>
  );
};

export default PlantSpeciesDetail;