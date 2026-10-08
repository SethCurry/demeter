import React from 'react';
import { Row, Col, Card, Statistic, Typography, List, Tag, Space, Alert, Spin } from 'antd';
import {
  HomeOutlined,
  ExperimentOutlined,
  BulbOutlined,
  ShareAltOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useAsync } from '../hooks/useAsync';
import { useNavigate } from 'react-router-dom';
import { api, type Enclosure, type System, type Flow, type PlantSite } from '../api';

const { Title, Text } = Typography;

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const enc = useAsync<Enclosure[]>(() => api.listEnclosures(), []);
  const sys = useAsync<System[]>(() => api.listSystems(), []);
  const flw = useAsync<Flow[]>(() => api.listFlows(), []);
  const sites = useAsync<PlantSite[]>(() => api.listPlantSites(), []);

  const loading = enc.loading || sys.loading || flw.loading || sites.loading;
  const error = enc.error || sys.error || flw.error || sites.error;

  const enclosures = enc.data ?? [];
  const systems = sys.data ?? [];
  const flows = flw.data ?? [];
  const plantSites = sites.data ?? [];

  const enclosureById = new Map(enclosures.map((e) => [e.ID, e]));
  const flowsBySystem = new Map<number, number>();
  flows.forEach((f) => flowsBySystem.set(f.SystemID, (flowsBySystem.get(f.SystemID) ?? 0) + 1));

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div>
        <Title level={3} style={{ marginBottom: 4 }}>Dashboard</Title>
        <Text type="secondary">Live overview of all hydroponics systems.</Text>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Row gutter={[16, 16]}>
          <Col xs={12} sm={12} md={6}>
            <Card>
              <Statistic title="Enclosures" value={enclosures.length} prefix={<HomeOutlined />} />
            </Card>
          </Col>
          <Col xs={12} sm={12} md={6}>
            <Card>
              <Statistic title="Systems" value={systems.length} prefix={<ExperimentOutlined />} />
            </Card>
          </Col>
          <Col xs={12} sm={12} md={6}>
            <Card>
              <Statistic title="Flows" value={flows.length} prefix={<ShareAltOutlined />} />
            </Card>
          </Col>
          <Col xs={12} sm={12} md={6}>
            <Card>
              <Statistic title="Plant Sites" value={plantSites.length} prefix={<BulbOutlined />} />
            </Card>
          </Col>
        </Row>
      </Spin>

      <Card title="Systems Status" extra={<ThunderboltOutlined />}>
        <List
          dataSource={systems}
          locale={{ emptyText: loading ? 'Loading…' : 'No systems' }}
          renderItem={(s) => {
            const enclosure = enclosureById.get(s.EnclosureID);
            const flowCount = flowsBySystem.get(s.ID) ?? 0;
            return (
              <List.Item>
                <Space size="large" wrap style={{ width: '100%' }}>
                  <div style={{ minWidth: 180 }}>
                    <a onClick={() => navigate(`/systems/${s.ID}`)}><strong>{s.Name}</strong></a>
                    <div><Text type="secondary">{enclosure?.Name ?? '—'}</Text></div>
                  </div>
                  <Tag color="green">{flowCount} flows</Tag>
                  <Tag color="gold">Enclosure #{s.EnclosureID}</Tag>
                </Space>
              </List.Item>
            );
          }}
        />
      </Card>
    </Space>
  );
};

export default Dashboard;