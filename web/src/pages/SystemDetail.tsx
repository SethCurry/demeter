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
  ExperimentOutlined,
  DeploymentUnitOutlined,
  AimOutlined,
  HomeOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { useAsync } from '../hooks/useAsync';
import AddNoteModal from '../components/AddNoteModal';
import PlantSites3D from '../components/PlantSites3D';
import {
  api,
  nullInt,
  nullTime,
  nullString,
  type System,
  type Enclosure,
  type Flow,
  type PlantSite,
  type Plant,
  type SimpleNote,
} from '../api';

const { Title, Text } = Typography;

const SystemDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const systemId = Number(id);

  const [noteModalOpen, setNoteModalOpen] = useState(false);

  const validId = Number.isFinite(systemId) && systemId > 0;

  const sys = useAsync<System>(
    () => (validId ? api.getSystem(systemId) : Promise.reject(new Error('invalid id'))),
    [systemId]
  );
  const enc = useAsync<Enclosure[]>(() => api.listEnclosures(), []);
  const flw = useAsync<Flow[]>(() => api.listFlows(validId ? { systemId } : undefined), [systemId]);
  const sites = useAsync<PlantSite[]>(() => api.listPlantSites(), []);
  const plants = useAsync<Plant[]>(() => api.listPlants(), []);
  const notes = useAsync<SimpleNote[]>(() => api.listSystemNotes(), []);

  const system = sys.data;
  const enclosures = enc.data ?? [];
  const flows = (flw.data ?? []).filter((f) => f.SystemID === systemId);
  const plantSites = sites.data ?? [];
  const allPlants = plants.data ?? [];
  const allNotes = notes.data ?? [];

  const loading = sys.loading || enc.loading || flw.loading || sites.loading || plants.loading || notes.loading;
  const error = sys.error || enc.error || flw.error || sites.error || plants.error || notes.error;

  const enclosureById = new Map(enclosures.map((e) => [e.ID, e]));
  const flowById = new Map(flows.map((f) => [f.ID, f]));

  const flowIds = new Set(flows.map((f) => f.ID));
  const systemSites = plantSites.filter((s) => flowIds.has(s.FlowID));
  const siteIds = new Set(systemSites.map((s) => s.ID));
  const systemPlants = allPlants.filter((p) => siteIds.has(p.PlantSiteID));

  // Notes belong to this system when their content mentions it; the backend
  // currently exposes a flat list, so we cannot attribute them precisely.
  const systemNotes = allNotes;

  const reloadAll = () => {
    sys.reload();
    enc.reload();
    flw.reload();
    sites.reload();
    plants.reload();
    notes.reload();
  };

  if (!validId) {
    return (
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Alert type="error" message="Invalid system id" showIcon />
        <Button icon={<RollbackOutlined />} onClick={() => navigate('/systems')}>
          Back to systems
        </Button>
      </Space>
    );
  }

  if (sys.error) {
    return (
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Alert type="error" message="Failed to load system" description={sys.error} showIcon />
        <Button icon={<RollbackOutlined />} onClick={() => navigate('/systems')}>
          Back to systems
        </Button>
      </Space>
    );
  }

  const enclosure = system ? enclosureById.get(system.EnclosureID) : undefined;
  const fmtTime = (t: string | null) => (t ? new Date(t).toLocaleString() : '—');

  const flowColumns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 60 },
    {
      title: 'Name',
      dataIndex: 'Name',
      key: 'name',
      render: (name: string, r: Flow) => (
        <a onClick={() => navigate(`/flows/${r.ID}`)}>{name}</a>
      ),
    },
    {
      title: 'Parent Flow',
      key: 'parent',
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
      render: (_: unknown, r: Flow) => (
        <Tag color="green">
          {plantSites.filter((s) => s.FlowID === r.ID).length}
        </Tag>
      ),
    },
  ];

  const siteColumns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 60 },
    {
      title: 'Flow',
      key: 'flow',
      render: (_: unknown, r: PlantSite) => {
        const f = flowById.get(r.FlowID);
        return f ? <a onClick={() => navigate(`/flows/${f.ID}`)}>{f.Name}</a> : '—';
      },
    },
    {
      title: 'Position',
      key: 'pos',
      render: (_: unknown, r: PlantSite) => `(${r.X}, ${r.Y}, ${r.Z})`,
    },
  ];

  const plantColumns = [
    { title: 'Plant ID', dataIndex: 'ID', key: 'id', width: 80 },
    {
      title: 'Site',
      key: 'site',
      render: (_: unknown, r: Plant) => {
        const site = systemSites.find((s) => s.ID === r.PlantSiteID);
        return site ? `#${site.ID} (${site.X},${site.Y},${site.Z})` : '—';
      },
    },
    {
      title: 'Flow',
      key: 'flow',
      render: (_: unknown, r: Plant) => {
        const site = systemSites.find((s) => s.ID === r.PlantSiteID);
        if (!site) return '—';
        const f = flowById.get(site.FlowID);
        return f ? <a onClick={() => navigate(`/flows/${f.ID}`)}>{f.Name}</a> : '—';
      },
    },
    {
      title: 'Planted',
      key: 'planted',
      render: (_: unknown, r: Plant) => {
        const planted = nullTime(r.PlantedOn);
        return planted ? new Date(planted).toLocaleDateString() : '—';
      },
    },
  ];

  const noteColumns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 60 },
    {
      title: 'Timestamp',
      key: 'ts',
      render: (_: unknown, r: SimpleNote) => fmtTime(nullTime(r.Timestamp)),
    },
    {
      title: 'Content',
      key: 'content',
      render: (_: unknown, r: SimpleNote) => nullString(r.Content) ?? '—',
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Space align="center">
            <Button
              icon={<RollbackOutlined />}
              onClick={() => navigate('/systems')}
              type="text"
            />
            <Title level={3} style={{ marginBottom: 4 }}>
              <ExperimentOutlined /> {system ? system.Name : `System #${systemId}`}
            </Title>
          </Space>
          <Text type="secondary">Details, flows, and contents of this system.</Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={reloadAll}>
            Refresh
          </Button>
          <Button icon={<PlusOutlined />} onClick={() => setNoteModalOpen(true)}>
            Add Note
          </Button>
        </Space>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Card size="small">
            <Descriptions title="System" column={{ xs: 1, sm: 2, md: 3 }} size="small">
              <Descriptions.Item label="ID">{system?.ID ?? '—'}</Descriptions.Item>
              <Descriptions.Item label="Name">{system?.Name ?? '—'}</Descriptions.Item>
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
              <Descriptions.Item label="Flows">{flows.length}</Descriptions.Item>
              <Descriptions.Item label="Plant Sites">{systemSites.length}</Descriptions.Item>
              <Descriptions.Item label="Plants">{systemPlants.length}</Descriptions.Item>
            </Descriptions>
          </Card>

          <Row gutter={[16, 16]}>
            <Col xs={24} sm={12} md={8}>
              <Card size="small">
                <Statistic title="Flows" value={flows.length} prefix={<DeploymentUnitOutlined />} />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={8}>
              <Card size="small">
                <Statistic title="Plant Sites" value={systemSites.length} prefix={<AimOutlined />} />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={8}>
              <Card size="small">
                <Statistic title="Plants" value={systemPlants.length} prefix={<AimOutlined />} />
              </Card>
            </Col>
          </Row>

          <Card size="small" title="Plant Site Layout (3D)">
            <PlantSites3D flows={flows} plantSites={systemSites} plants={systemPlants} />
          </Card>

          <Card size="small" title={`Flows (${flows.length})`}>
            <Table
              rowKey="ID"
              columns={flowColumns}
              dataSource={flows}
              pagination={false}
              size="small"
              locale={{ emptyText: 'No flows in this system' }}
            />
          </Card>

          <Card size="small" title={`Plant Sites (${systemSites.length})`}>
            <Table
              rowKey="ID"
              columns={siteColumns}
              dataSource={systemSites}
              pagination={false}
              size="small"
              locale={{ emptyText: 'No plant sites in this system' }}
            />
          </Card>

          <Card size="small" title={`Plants (${systemPlants.length})`}>
            <Table
              rowKey="ID"
              columns={plantColumns}
              dataSource={systemPlants}
              pagination={false}
              size="small"
              locale={{ emptyText: 'No plants in this system' }}
            />
          </Card>

          <Card size="small" title={`Notes (${systemNotes.length})`}>
            <Table
              rowKey="ID"
              columns={noteColumns}
              dataSource={systemNotes}
              pagination={false}
              size="small"
              locale={{ emptyText: 'No notes' }}
            />
          </Card>
        </Space>
      </Spin>

      <AddNoteModal
        open={noteModalOpen}
        title="Add System Note"
        onClose={() => setNoteModalOpen(false)}
        onSubmit={(content, timestamp) =>
          api.createSystemNote(content, timestamp).then(() => notes.reload())
        }
      />
    </Space>
  );
};

export default SystemDetail;