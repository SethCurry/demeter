import React, { useState } from 'react';
import {
  Row,
  Col,
  Card,
  Statistic,
  Typography,
  Table,
  Tag,
  Space,
  Button,
  Alert,
  Spin,
  Descriptions,
} from 'antd';
import {
  ReloadOutlined,
  RollbackOutlined,
  AimOutlined,
  ExperimentOutlined,
  DeploymentUnitOutlined,
  HomeOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { useAsync } from '../hooks/useAsync';
import PlantSites3D from '../components/PlantSites3D';
import {
  api,
  nullTime,
  nullInt,
  type Plant,
  type PlantSite,
  type Flow,
  type System,
  type Enclosure,
  type PlantNote,
  type PlantSpecies,
  type PlantGenus,
} from '../api';

const { Title, Text } = Typography;

const PlantDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const plantId = Number(id);

  const validId = Number.isFinite(plantId) && plantId > 0;

  const pl = useAsync<Plant>(
    () => (validId ? api.getPlant(plantId) : Promise.reject(new Error('invalid id'))),
    [plantId]
  );
  const sites = useAsync<PlantSite[]>(() => api.listPlantSites(), []);
  const flw = useAsync<Flow[]>(() => api.listFlows(), []);
  const sys = useAsync<System[]>(() => api.listSystems(), []);
  const enc = useAsync<Enclosure[]>(() => api.listEnclosures(), []);
  const plants = useAsync<Plant[]>(() => api.listPlants(), []);
  const notes = useAsync<PlantNote[]>(() => api.listPlantNotes(), []);
  const spec = useAsync<PlantSpecies[]>(() => api.listPlantSpecies(), []);
  const gen = useAsync<PlantGenus[]>(() => api.listPlantGenera(), []);

  const plant = pl.data;
  const plantSites = sites.data ?? [];
  const flows = flw.data ?? [];
  const systems = sys.data ?? [];
  const enclosures = enc.data ?? [];
  const allPlants = plants.data ?? [];
  const allNotes = notes.data ?? [];
  const allSpecies = spec.data ?? [];
  const allGenera = gen.data ?? [];

  const loading =
    pl.loading || sites.loading || flw.loading || sys.loading || enc.loading || plants.loading || notes.loading || spec.loading || gen.loading;
  const error =
    pl.error || sites.error || flw.error || sys.error || enc.error || plants.error || notes.error || spec.error || gen.error;

  const flowById = new Map(flows.map((f) => [f.ID, f]));
  const systemById = new Map(systems.map((s) => [s.ID, s]));
  const enclosureById = new Map(enclosures.map((e) => [e.ID, e]));
  const speciesById = new Map(allSpecies.map((s) => [s.ID, s]));
  const genusById = new Map(allGenera.map((g) => [g.ID, g]));

  const site = plant ? plantSites.find((s) => s.ID === plant.PlantSiteID) : undefined;
  const flow = site ? flowById.get(site.FlowID) : undefined;
  const system = flow ? systemById.get(flow.SystemID) : undefined;
  const enclosure = system ? enclosureById.get(system.EnclosureID) : undefined;

  const speciesId = plant ? nullInt(plant.SpeciesID) : null;
  const species = speciesId != null ? speciesById.get(speciesId) : undefined;
  const speciesGenusId = species ? nullInt(species.PlantGenusID) : null;
  const speciesGenus = speciesGenusId != null ? genusById.get(speciesGenusId) : undefined;

  // All flows in the same system, so the 3D diagram shows the whole system
  // context with this plant's site highlighted.
  const systemFlows = flow ? flows.filter((f) => f.SystemID === flow.SystemID) : [];
  const systemFlowIds = new Set(systemFlows.map((f) => f.ID));
  const systemSites = plantSites.filter((s) => systemFlowIds.has(s.FlowID));
  const systemSiteIds = new Set(systemSites.map((s) => s.ID));
  const systemPlants = allPlants.filter((p) => systemSiteIds.has(p.PlantSiteID));

  // Plant notes are keyed by `PlanID` in the backend; the flat list exposed
  // by the API cannot be attributed precisely to a single plant, so we show
  // all of them here (mirrors the system/flow note behavior).
  const plantNotes = allNotes;

  const reloadAll = () => {
    pl.reload();
    sites.reload();
    flw.reload();
    sys.reload();
    enc.reload();
    plants.reload();
    notes.reload();
    spec.reload();
    gen.reload();
  };

  if (!validId) {
    return (
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Alert type="error" message="Invalid plant id" showIcon />
        <Button icon={<RollbackOutlined />} onClick={() => navigate('/plants')}>
          Back to plants
        </Button>
      </Space>
    );
  }

  if (pl.error) {
    return (
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Alert type="error" message="Failed to load plant" description={pl.error} showIcon />
        <Button icon={<RollbackOutlined />} onClick={() => navigate('/plants')}>
          Back to plants
        </Button>
      </Space>
    );
  }

  const fmtTime = (t: string | null) => (t ? new Date(t).toLocaleString() : '—');
  const planted = plant ? nullTime(plant.PlantedOn) : null;

  const noteColumns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 60 },
    {
      title: 'Timestamp',
      key: 'ts',
      render: (_: unknown, r: PlantNote) => fmtTime(nullTime(r.Timestamp)),
    },
    {
      title: 'Content',
      key: 'content',
      render: (_: unknown, r: PlantNote) => {
        const c = r.Content && r.Content.Valid ? r.Content.String : null;
        return c ?? '—';
      },
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Space align="center">
            <Button
              icon={<RollbackOutlined />}
              onClick={() => navigate('/plants')}
              type="text"
            />
            <Title level={3} style={{ marginBottom: 4 }}>
              <AimOutlined /> {plant ? `Plant #${plant.ID}` : `Plant #${plantId}`}
            </Title>
          </Space>
          <Text type="secondary">Details and location of this plant within its system.</Text>
        </div>
        <Button icon={<ReloadOutlined />} onClick={reloadAll}>
          Refresh
        </Button>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Card size="small">
            <Descriptions title="Plant" column={{ xs: 1, sm: 2, md: 3 }} size="small">
              <Descriptions.Item label="ID">{plant?.ID ?? '—'}</Descriptions.Item>
              <Descriptions.Item label="Planted On">
                {planted ? new Date(planted).toLocaleDateString() : '—'}
              </Descriptions.Item>
              <Descriptions.Item label="Species">
                {species ? (
                  <Button
                    type="link"
                    style={{ padding: 0 }}
                    onClick={() => navigate(`/plant-species/${species.ID}`)}
                  >
                    {speciesGenus ? `${speciesGenus.Name} ${species.Name}` : species.Name}
                  </Button>
                ) : (
                  '—'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Plant Site">
                {site ? (
                  <Button
                    type="link"
                    style={{ padding: 0 }}
                    onClick={() => navigate('/plants')}
                  >
                    #{site.ID} ({site.X}, {site.Y}, {site.Z})
                  </Button>
                ) : (
                  '—'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Flow">
                {flow ? (
                  <Button
                    type="link"
                    style={{ padding: 0 }}
                    onClick={() => navigate(`/flows/${flow.ID}`)}
                  >
                    <DeploymentUnitOutlined /> {flow.Name}
                  </Button>
                ) : (
                  '—'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="System">
                {system ? (
                  <Button
                    type="link"
                    style={{ padding: 0 }}
                    onClick={() => navigate(`/systems/${system.ID}`)}
                  >
                    <ExperimentOutlined /> {system.Name}
                  </Button>
                ) : (
                  '—'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Enclosure">
                {enclosure ? (
                  <Button
                    type="link"
                    style={{ padding: 0 }}
                    onClick={() => navigate(`/enclosures/${enclosure.ID}`)}
                  >
                    <HomeOutlined /> {enclosure.Name}
                  </Button>
                ) : (
                  '—'
                )}
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Row gutter={[16, 16]}>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="Site Position"
                  value={site ? `(${site.X}, ${site.Y}, ${site.Z})` : '—'}
                  prefix={<AimOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="Sites in System"
                  value={systemSites.length}
                  prefix={<AimOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="Plants in System"
                  value={systemPlants.length}
                  prefix={<AimOutlined />}
                />
              </Card>
            </Col>
          </Row>

          <Card size="small" title="Plant Location in System (3D)">
            <PlantSites3D
              flows={systemFlows}
              plantSites={systemSites}
              plants={systemPlants}
              highlightSiteId={site?.ID}
            />
          </Card>

          <Card size="small" title={`Plant Notes (${plantNotes.length})`}>
            <Table
              rowKey="ID"
              columns={noteColumns}
              dataSource={plantNotes}
              pagination={false}
              size="small"
              locale={{ emptyText: 'No notes' }}
            />
          </Card>
        </Space>
      </Spin>
    </Space>
  );
};

export default PlantDetail;