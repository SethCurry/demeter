import React, { useState } from 'react';
import { Row, Col, Card, Statistic, Typography, Table, Tag, Space, Button, Modal, Form, Input, message, Alert, Spin } from 'antd';
import { PlusOutlined, ReloadOutlined, RightOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAsync } from '../hooks/useAsync';
import { api, type Enclosure, type System } from '../api';

const { Title, Text } = Typography;

const Enclosures: React.FC = () => {
  const navigate = useNavigate();
  const enc = useAsync<Enclosure[]>(() => api.listEnclosures(), []);
  const sys = useAsync<System[]>(() => api.listSystems(), []);

  const [modalOpen, setModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm<{ name: string }>();

  const enclosures = enc.data ?? [];
  const systems = sys.data ?? [];
  const loading = enc.loading || sys.loading;
  const error = enc.error || sys.error;

  const systemsByEnclosure = new Map<number, number>();
  systems.forEach((s) =>
    systemsByEnclosure.set(s.EnclosureID, (systemsByEnclosure.get(s.EnclosureID) ?? 0) + 1)
  );

  const submit = (values: { name: string }) => {
    setSubmitting(true);
    api
      .createEnclosure(values.name)
      .then(() => {
        message.success('Enclosure created');
        setModalOpen(false);
        form.resetFields();
        enc.reload();
      })
      .catch((e) => message.error(e instanceof Error ? e.message : String(e)))
      .finally(() => setSubmitting(false));
  };

  const columns = [
    { title: 'ID', dataIndex: 'ID', key: 'id', width: 60 },
    {
      title: 'Name',
      dataIndex: 'Name',
      key: 'name',
      render: (name: string, r: Enclosure) => (
        <Button type="link" style={{ padding: 0 }} onClick={() => navigate(`/enclosures/${r.ID}`)}>
          {name}
        </Button>
      ),
    },
    {
      title: 'Systems',
      key: 'systems',
      render: (_: unknown, r: Enclosure) => (
        <Tag color="green">{systemsByEnclosure.get(r.ID) ?? 0}</Tag>
      ),
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>Enclosures</Title>
          <Text type="secondary">Spaces that share air qualities.</Text>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => { enc.reload(); sys.reload(); }}>
            Refresh
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            New Enclosure
          </Button>
        </Space>
      </div>

      {error && <Alert type="error" message="Failed to load data" description={error} showIcon />}

      <Spin spinning={loading}>
        <Row gutter={[16, 16]}>
          {enclosures.map((e) => (
            <Col xs={24} sm={12} key={e.ID}>
              <Card
                title={e.Name}
                size="small"
                hoverable
                onClick={() => navigate(`/enclosures/${e.ID}`)}
                extra={<RightOutlined style={{ color: '#588157' }} />}
              >
                <Row gutter={16}>
                  <Col span={12}>
                    <Statistic title="Systems" value={systemsByEnclosure.get(e.ID) ?? 0} />
                  </Col>
                  <Col span={12}>
                    <Statistic title="ID" value={e.ID} />
                  </Col>
                </Row>
              </Card>
            </Col>
          ))}
        </Row>

        <Table rowKey="ID" columns={columns} dataSource={enclosures} pagination={false} />
      </Spin>

      <Modal
        title="New Enclosure"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        okText="Create"
        confirmLoading={submitting}
      >
        <Form form={form} layout="vertical" onFinish={submit}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="Enclosure name" />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default Enclosures;