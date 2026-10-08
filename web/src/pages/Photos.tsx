import React from 'react';
import { Row, Col, Card, Typography, Empty, Button, Tag, Space } from 'antd';
import { CameraOutlined, PlayCircleOutlined } from '@ant-design/icons';

const { Title, Text, Paragraph } = Typography;

interface PhotoSet {
  id: number;
  name: string;
  target: 'Whole system' | 'Roots';
  interval: string;
  latest: string;
  frames: number;
}

const photoSets: PhotoSet[] = [
  { id: 1, name: 'Greenhouse A - Wide', target: 'Whole system', interval: '10 min', latest: '2025-10-03 15:50', frames: 432 },
  { id: 2, name: 'Lettuce NFT - Roots', target: 'Roots', interval: '30 min', latest: '2025-10-03 15:30', frames: 144 },
  { id: 3, name: 'Tomato DWC - Roots', target: 'Roots', interval: '30 min', latest: '2025-10-03 15:30', frames: 140 },
];

const Photos: React.FC = () => {
  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>Photos</Title>
          <Text type="secondary">Timelapse capture of whole systems and roots.</Text>
        </div>
        <Button type="primary" icon={<CameraOutlined />}>New Timelapse</Button>
      </div>

      <Row gutter={[16, 16]}>
        {photoSets.map((p) => (
          <Col xs={24} sm={12} md={8} key={p.id}>
            <Card
              cover={
                <div
                  style={{
                    height: 180,
                    background: 'linear-gradient(135deg, #1f3a22, #3a7d34)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: '#fff',
                  }}
                >
                  <CameraOutlined style={{ fontSize: 48 }} />
                </div>
              }
              actions={[<PlayCircleOutlined key="play" />]}
            >
              <Card.Meta
                title={p.name}
                description={
                  <Space direction="vertical" size={2}>
                    <span><Tag color={p.target === 'Roots' ? 'gold' : 'green'}>{p.target}</Tag></span>
                    <Text type="secondary">Every {p.interval} · {p.frames} frames</Text>
                    <Text type="secondary">Latest: {p.latest}</Text>
                  </Space>
                }
              />
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