import React from "react";
import { Row, Col, Card, Typography, Empty, Button, Space } from "antd";
import { CameraOutlined, PlayCircleOutlined } from "@ant-design/icons";
import { useAsync } from "../hooks/useAsync";
import { api, Timelapse } from "../api";

const { Title, Text, Paragraph } = Typography;

const Photos: React.FC = () => {
  const timelapseResults = useAsync<Timelapse[]>(
    () => api.listTimelapses(),
    [],
  );
  const timelapses = timelapseResults.data ?? [];
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
          <Title level={3} style={{ marginBottom: 4 }}>
            Photos
          </Title>
          <Text type="secondary">
            Timelapse capture of whole systems and roots.
          </Text>
        </div>
        <Button type="primary" icon={<CameraOutlined />}>
          New Timelapse
        </Button>
      </div>

      <Row gutter={[16, 16]}>
        {timelapses.map((p) => (
          <Col xs={24} sm={12} md={8} key={p.name}>
            <Card
              cover={
                <div
                  style={{
                    height: 180,
                    background: "linear-gradient(135deg, #1f3a22, #3a7d34)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    color: "#fff",
                  }}
                >
                  <CameraOutlined style={{ fontSize: 48 }} />
                </div>
              }
              actions={[
                <a href={`/api/timelapse/${p.name}/timelapse.mjpeg`}>
                  <PlayCircleOutlined key="play" />
                </a>,
              ]}
            >
              <Card.Meta title={p.name} />
            </Card>
          </Col>
        ))}
      </Row>

      <Card>
        <Empty description="Select a timelapse to preview the latest frame." />
        <Paragraph style={{ marginTop: 12 }} type="secondary">
          Frames are stored on disk and compiled into videos on demand.
        </Paragraph>
      </Card>
    </Space>
  );
};

export default Photos;
