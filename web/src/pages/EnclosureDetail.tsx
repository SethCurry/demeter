import React, { useState } from "react";
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
} from "antd";
import {
  ReloadOutlined,
  RollbackOutlined,
  ExperimentOutlined,
  DeploymentUnitOutlined,
  AimOutlined,
  HomeOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import { useParams, useNavigate } from "react-router-dom";
import { useAsync } from "../hooks/useAsync";
import AddNoteModal from "../components/AddNoteModal";
import {
  api,
  nullInt,
  nullTime,
  nullString,
  type Enclosure,
  type System,
  type Flow,
  type PlantSite,
  type PlantWithFlow,
  type EnclosureAirTemperature,
  type EnclosureAirHumidity,
  type SimpleNote,
} from "../api";
import PlantListCard from "../components/PlantListCard";
import SystemListCard from "../components/SystemListCard";
import TemperatureChart from "../components/charts/TemperatureChart";
import HumidityChart from "../components/charts/HumidityChart";

const { Title, Text } = Typography;

const READING_LIMIT = 20;

const EnclosureDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const enclosureId = Number(id);

  const [noteModalOpen, setNoteModalOpen] = useState(false);

  const validId = Number.isFinite(enclosureId) && enclosureId > 0;

  const enc = useAsync<Enclosure>(
    () =>
      validId
        ? api.getEnclosure(enclosureId)
        : Promise.reject(new Error("invalid id")),
    [enclosureId],
  );
  const sys = useAsync<System[]>(
    () => api.listSystems(validId ? enclosureId : undefined),
    [enclosureId],
  );
  const flw = useAsync<Flow[]>(() => api.listFlows(), []);
  const sites = useAsync<PlantSite[]>(() => api.listPlantSites(), []);
  const plants = useAsync<PlantWithFlow[]>(
    () => api.enclosurePlants(enclosureId),
    [],
  );
  const temps = useAsync<EnclosureAirTemperature[]>(
    () =>
      validId
        ? api.listEnclosureAirTemperatures(enclosureId, READING_LIMIT)
        : Promise.resolve([]),
    [enclosureId],
  );
  const hums = useAsync<EnclosureAirHumidity[]>(
    () =>
      validId
        ? api.listEnclosureAirHumidity(enclosureId, READING_LIMIT)
        : Promise.resolve([]),
    [enclosureId],
  );
  const notes = useAsync<SimpleNote[]>(() => api.listEnclosureNotes(), []);

  const enclosure = enc.data;
  const systems = (sys.data ?? []).filter((s) => s.EnclosureID === enclosureId);
  const flows = flw.data ?? [];
  const plantSites = sites.data ?? [];
  const allPlants = plants.data ?? [];
  const temperatures = temps.data ?? [];
  const humidities = hums.data ?? [];
  const allNotes = notes.data ?? [];
  const enclosureNotes = allNotes;

  const loading =
    enc.loading ||
    sys.loading ||
    flw.loading ||
    sites.loading ||
    plants.loading ||
    temps.loading ||
    hums.loading ||
    notes.loading;
  const error =
    enc.error ||
    sys.error ||
    flw.error ||
    sites.error ||
    plants.error ||
    temps.error ||
    hums.error ||
    notes.error;

  const systemIds = new Set(systems.map((s) => s.ID));
  const enclosureFlows = flows.filter((f) => systemIds.has(f.SystemID));
  const flowIds = new Set(enclosureFlows.map((f) => f.ID));
  const enclosureSites = plantSites.filter((s) => flowIds.has(s.FlowID));
  const siteIds = new Set(enclosureSites.map((s) => s.ID));
  const enclosurePlants = allPlants.filter((p) => siteIds.has(p.PlantSiteID));

  const systemById = new Map(systems.map((s) => [s.ID, s]));
  const flowById = new Map(enclosureFlows.map((f) => [f.ID, f]));

  const reloadAll = () => {
    enc.reload();
    sys.reload();
    flw.reload();
    sites.reload();
    plants.reload();
    temps.reload();
    hums.reload();
    notes.reload();
  };

  if (!validId) {
    return (
      <Space direction="vertical" size="large" style={{ width: "100%" }}>
        <Alert type="error" message="Invalid enclosure id" showIcon />
        <Button
          icon={<RollbackOutlined />}
          onClick={() => navigate("/enclosures")}
        >
          Back to enclosures
        </Button>
      </Space>
    );
  }

  if (enc.error) {
    return (
      <Space direction="vertical" size="large" style={{ width: "100%" }}>
        <Alert
          type="error"
          message="Failed to load enclosure"
          description={enc.error}
          showIcon
        />
        <Button
          icon={<RollbackOutlined />}
          onClick={() => navigate("/enclosures")}
        >
          Back to enclosures
        </Button>
      </Space>
    );
  }

  const latestTemp = temperatures[0];
  const latestHumidity = humidities[0];
  const fmtTime = (t: string | null) =>
    t ? new Date(t).toLocaleString() : "—";

  const flowColumns = [
    { title: "ID", dataIndex: "ID", key: "id", width: 60 },
    {
      title: "Name",
      dataIndex: "Name",
      key: "name",
      render: (name: string, r: Flow) => (
        <a onClick={() => navigate(`/flows/${r.ID}`)}>{name}</a>
      ),
    },
    {
      title: "System",
      key: "system",
      render: (_: unknown, r: Flow) => {
        const s = systemById.get(r.SystemID);
        return s ? (
          <a onClick={() => navigate(`/systems/${s.ID}`)}>{s.Name}</a>
        ) : (
          "—"
        );
      },
    },
    {
      title: "Parent Flow",
      key: "parent",
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
  ];

  const noteColumns = [
    { title: "ID", dataIndex: "ID", key: "id", width: 60 },
    {
      title: "Timestamp",
      key: "ts",
      render: (_: unknown, r: SimpleNote) => fmtTime(nullTime(r.Timestamp)),
    },
    {
      title: "Content",
      key: "content",
      render: (_: unknown, r: SimpleNote) => nullString(r.Content) ?? "—",
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: "100%" }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
        }}
      >
        <div>
          <Space align="center">
            <Button
              icon={<RollbackOutlined />}
              onClick={() => navigate("/enclosures")}
              type="text"
            />
            <Title level={3} style={{ marginBottom: 4 }}>
              <HomeOutlined />{" "}
              {enclosure ? enclosure.Name : `Enclosure #${enclosureId}`}
            </Title>
          </Space>
          <Text type="secondary">
            Details, metrics, and contents of this enclosure.
          </Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={reloadAll}>
            Refresh
          </Button>
          <Button
            icon={<PlusOutlined />}
            onClick={() => setNoteModalOpen(true)}
          >
            Add Note
          </Button>
        </Space>
      </div>

      {error && (
        <Alert
          type="error"
          message="Failed to load data"
          description={error}
          showIcon
        />
      )}

      <Spin spinning={loading}>
        <Space direction="vertical" size="large" style={{ width: "100%" }}>
          <Card size="small">
            <Descriptions
              title="Enclosure"
              column={{ xs: 1, sm: 2, md: 3 }}
              size="small"
            >
              <Descriptions.Item label="ID">
                {enclosure?.ID ?? "—"}
              </Descriptions.Item>
              <Descriptions.Item label="Name">
                {enclosure?.Name ?? "—"}
              </Descriptions.Item>
              <Descriptions.Item label="Systems">
                {systems.length}
              </Descriptions.Item>
              <Descriptions.Item label="Flows">
                {enclosureFlows.length}
              </Descriptions.Item>
              <Descriptions.Item label="Plant Sites">
                {enclosureSites.length}
              </Descriptions.Item>
              <Descriptions.Item label="Plants">
                {enclosurePlants.length}
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Row gutter={[16, 16]}>
            <Col xs={24} sm={12} md={8}>
              <Card size="small">
                <Statistic
                  title="Systems"
                  value={systems.length}
                  prefix={<ExperimentOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={8}>
              <Card size="small">
                <Statistic
                  title="Flows"
                  value={enclosureFlows.length}
                  prefix={<DeploymentUnitOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={8}>
              <Card size="small">
                <Statistic
                  title="Plants"
                  value={enclosurePlants.length}
                  prefix={<AimOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={12}>
              <Card size="small" title="Latest Air Temperature">
                {latestTemp ? (
                  <Statistic
                    value={latestTemp.TemperatureC}
                    precision={1}
                    suffix="°C"
                  />
                ) : (
                  <Text type="secondary">No readings yet</Text>
                )}
                {latestTemp && (
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    <br />
                    {fmtTime(nullTime(latestTemp.Timestamp))}
                  </Text>
                )}
                <TemperatureChart
                  label="Temperature"
                  data={temperatures.map((x) => {
                    return {
                      timestamp: x.Timestamp.Time,
                      temperature: x.TemperatureC,
                    };
                  })}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={12}>
              <Card size="small" title="Latest Air Humidity">
                {latestHumidity ? (
                  <Statistic
                    value={latestHumidity.HumidityRh}
                    precision={1}
                    suffix="% RH"
                  />
                ) : (
                  <Text type="secondary">No readings yet</Text>
                )}
                {latestHumidity && (
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    <br />
                    {fmtTime(nullTime(latestHumidity.Timestamp))}
                  </Text>
                )}
                <HumidityChart
                  label="Humidity"
                  data={humidities.map((x) => {
                    return {
                      timestamp: x.Timestamp.Time,
                      humidity: x.HumidityRh,
                    };
                  })}
                />
              </Card>
            </Col>
          </Row>

          <Card size="small" title={`Notes (${enclosureNotes.length})`}>
            <Table
              rowKey="ID"
              columns={noteColumns}
              dataSource={enclosureNotes}
              pagination={false}
              size="small"
              locale={{ emptyText: "No notes" }}
            />
          </Card>

          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              <SystemListCard systems={systems} />
            </Col>

            <Col xs={24} lg={12}>
              <Card size="small" title={`Flows (${enclosureFlows.length})`}>
                <Table
                  rowKey="ID"
                  columns={flowColumns}
                  dataSource={enclosureFlows}
                  pagination={false}
                  size="small"
                  locale={{ emptyText: "No flows in this enclosure" }}
                />
              </Card>
            </Col>
          </Row>

          <PlantListCard plants={enclosurePlants} />
        </Space>
      </Spin>

      <AddNoteModal
        open={noteModalOpen}
        title="Add Enclosure Note"
        onClose={() => setNoteModalOpen(false)}
        onSubmit={(content, timestamp) =>
          api.createEnclosureNote(content, timestamp).then(() => notes.reload())
        }
      />
    </Space>
  );
};

export default EnclosureDetail;
