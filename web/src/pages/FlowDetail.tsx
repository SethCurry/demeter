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
  HomeOutlined,
  DeploymentUnitOutlined,
  AimOutlined,
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
  type Flow,
  type System,
  type Enclosure,
  type PlantSite,
  type PlantWithFlow,
  type SimpleNote,
} from "../api";
import PlantListCard from "../components/PlantListCard";

const { Title, Text } = Typography;

const FlowDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const flowId = Number(id);

  const [noteModalOpen, setNoteModalOpen] = useState(false);

  const validId = Number.isFinite(flowId) && flowId > 0;

  const flw = useAsync<Flow>(
    () =>
      validId ? api.getFlow(flowId) : Promise.reject(new Error("invalid id")),
    [flowId],
  );
  const allFlows = useAsync<Flow[]>(() => api.listFlows(), []);
  const sys = useAsync<System[]>(() => api.listSystems(), []);
  const enc = useAsync<Enclosure[]>(() => api.listEnclosures(), []);
  const sites = useAsync<PlantSite[]>(() => api.listPlantSites(), []);
  const plants = useAsync<PlantWithFlow[]>(() => api.flowPlants(flowId), []);
  const notes = useAsync<SimpleNote[]>(() => api.listFlowNotes(), []);

  const flow = flw.data;
  const flows = allFlows.data ?? [];
  const systems = sys.data ?? [];
  const enclosures = enc.data ?? [];
  const allPlantSites = sites.data ?? [];
  const plantSites = allPlantSites.filter((s) => s.FlowID === flowId);
  const allPlants = plants.data ?? [];
  const allNotes = notes.data ?? [];

  const loading =
    flw.loading ||
    allFlows.loading ||
    sys.loading ||
    enc.loading ||
    sites.loading ||
    plants.loading ||
    notes.loading;
  const error =
    flw.error ||
    allFlows.error ||
    sys.error ||
    enc.error ||
    sites.error ||
    plants.error ||
    notes.error;

  const systemById = new Map(systems.map((s) => [s.ID, s]));
  const enclosureById = new Map(enclosures.map((e) => [e.ID, e]));
  const flowById = new Map(flows.map((f) => [f.ID, f]));

  // Child flows: flows whose ParentFlowID points at this flow.
  const childFlows = flows.filter((f) => nullInt(f.ParentFlowID) === flowId);

  const siteIds = new Set(plantSites.map((s) => s.ID));
  const flowPlants = allPlants.filter((p) => siteIds.has(p.PlantSiteID));

  // Notes belong to this flow when their content mentions it; the backend
  // currently exposes a flat list, so we cannot attribute them precisely.
  const flowNotes = allNotes;

  const reloadAll = () => {
    flw.reload();
    allFlows.reload();
    sys.reload();
    enc.reload();
    sites.reload();
    plants.reload();
    notes.reload();
  };

  if (!validId) {
    return (
      <Space direction="vertical" size="large" style={{ width: "100%" }}>
        <Alert type="error" message="Invalid flow id" showIcon />
        <Button icon={<RollbackOutlined />} onClick={() => navigate("/flows")}>
          Back to flows
        </Button>
      </Space>
    );
  }

  if (flw.error) {
    return (
      <Space direction="vertical" size="large" style={{ width: "100%" }}>
        <Alert
          type="error"
          message="Failed to load flow"
          description={flw.error}
          showIcon
        />
        <Button icon={<RollbackOutlined />} onClick={() => navigate("/flows")}>
          Back to flows
        </Button>
      </Space>
    );
  }

  const system = flow ? systemById.get(flow.SystemID) : undefined;
  const enclosure = system ? enclosureById.get(system.EnclosureID) : undefined;
  const parentFlowId = flow ? nullInt(flow.ParentFlowID) : null;
  const parentFlow =
    parentFlowId != null ? flowById.get(parentFlowId) : undefined;
  const fmtTime = (t: string | null) =>
    t ? new Date(t).toLocaleString() : "—";

  const childFlowColumns = [
    { title: "ID", dataIndex: "ID", key: "id", width: 60 },
    {
      title: "Name",
      key: "name",
      render: (_: unknown, r: Flow) => (
        <Button
          type="link"
          style={{ padding: 0 }}
          onClick={() => navigate(`/flows/${r.ID}`)}
        >
          {r.Name}
        </Button>
      ),
    },
    {
      title: "Plant Sites",
      key: "sites",
      render: (_: unknown, r: Flow) => (
        <Tag color="green">
          {allPlantSites.filter((s) => s.FlowID === r.ID).length}
        </Tag>
      ),
    },
  ];

  const siteColumns = [
    { title: "ID", dataIndex: "ID", key: "id", width: 60 },
    {
      title: "Position",
      key: "pos",
      render: (_: unknown, r: PlantSite) => `(${r.X}, ${r.Y}, ${r.Z})`,
    },
    {
      title: "Plants",
      key: "plants",
      render: (_: unknown, r: PlantSite) => (
        <Tag color="green">
          {allPlants.filter((p) => p.PlantSiteID === r.ID).length}
        </Tag>
      ),
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
              onClick={() => navigate("/flows")}
              type="text"
            />
            <Title level={3} style={{ marginBottom: 4 }}>
              <DeploymentUnitOutlined /> {flow ? flow.Name : `Flow #${flowId}`}
            </Title>
          </Space>
          <Text type="secondary">
            Details, plant sites, and contents of this flow.
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
              title="Flow"
              column={{ xs: 1, sm: 2, md: 3 }}
              size="small"
            >
              <Descriptions.Item label="ID">
                {flow?.ID ?? "—"}
              </Descriptions.Item>
              <Descriptions.Item label="Name">
                {flow?.Name ?? "—"}
              </Descriptions.Item>
              <Descriptions.Item label="Parent Flow">
                {parentFlowId != null ? (
                  <Button
                    type="link"
                    style={{ padding: 0 }}
                    onClick={() => navigate(`/flows/${parentFlowId}`)}
                  >
                    <DeploymentUnitOutlined />{" "}
                    {parentFlow?.Name ?? `#${parentFlowId}`}
                  </Button>
                ) : (
                  <Tag>root</Tag>
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
                  "—"
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
                  "—"
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Child Flows">
                {childFlows.length}
              </Descriptions.Item>
              <Descriptions.Item label="Plant Sites">
                {plantSites.length}
              </Descriptions.Item>
              <Descriptions.Item label="Plants">
                {flowPlants.length}
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Row gutter={[16, 16]}>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="Child Flows"
                  value={childFlows.length}
                  prefix={<DeploymentUnitOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="Plant Sites"
                  value={plantSites.length}
                  prefix={<AimOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={8}>
              <Card size="small">
                <Statistic
                  title="Plants"
                  value={flowPlants.length}
                  prefix={<AimOutlined />}
                />
              </Card>
            </Col>
          </Row>

          <Card size="small" title={`Child Flows (${childFlows.length})`}>
            <Table
              rowKey="ID"
              columns={childFlowColumns}
              dataSource={childFlows}
              pagination={false}
              size="small"
              locale={{ emptyText: "No child flows" }}
            />
          </Card>

          <Card size="small" title={`Plant Sites (${plantSites.length})`}>
            <Table
              rowKey="ID"
              columns={siteColumns}
              dataSource={plantSites}
              pagination={false}
              size="small"
              locale={{ emptyText: "No plant sites in this flow" }}
            />
          </Card>

          <PlantListCard plants={flowPlants} />

          <Card size="small" title={`Notes (${flowNotes.length})`}>
            <Table
              rowKey="ID"
              columns={noteColumns}
              dataSource={flowNotes}
              pagination={false}
              size="small"
              locale={{ emptyText: "No notes" }}
            />
          </Card>
        </Space>
      </Spin>

      <AddNoteModal
        open={noteModalOpen}
        title="Add Flow Note"
        onClose={() => setNoteModalOpen(false)}
        onSubmit={(content, timestamp) =>
          api.createFlowNote(content, timestamp).then(() => notes.reload())
        }
      />
    </Space>
  );
};

export default FlowDetail;
