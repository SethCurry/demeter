import React from 'react';
import { Typography, Form, Input, InputNumber, Switch, Button, Card, Space, Divider, message } from 'antd';

const { Title, Text } = Typography;

const Settings: React.FC = () => {
  const [form] = Form.useForm();

  const onSave = () => {
    form.validateFields().then(() => message.success('Settings saved'));
  };

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div>
        <Title level={3} style={{ marginBottom: 4 }}>Settings</Title>
        <Text type="secondary">Backend connection and capture preferences.</Text>
      </div>

      <Card title="Connection">
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            websocketUrl: 'ws://localhost:8080/api/websocket/sensors',
            captureInterval: 10,
            retentionDays: 90,
            autoCapture: true,
          }}
        >
          <Form.Item name="websocketUrl" label="Sensor WebSocket URL">
            <Input />
          </Form.Item>
          <Form.Item name="captureInterval" label="Photo interval (minutes)">
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="retentionDays" label="Data retention (days)">
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="autoCapture" label="Auto-capture timelapses" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Divider />
          <Button type="primary" onClick={onSave}>Save</Button>
        </Form>
      </Card>
    </Space>
  );
};

export default Settings;