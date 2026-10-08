import React, { useState } from 'react';
import { Modal, Form, Input, DatePicker, message } from 'antd';
import dayjs from 'dayjs';

/**
 * A small dialog for creating a free-form text note.
 *
 * The dialog collects an optional timestamp (defaults to "now" server-side
 * when omitted) and the note content, then calls `onSubmit` and closes once
 * it resolves. Callers are responsible for reloading their note list after a
 * successful create.
 */
interface AddNoteModalProps {
  open: boolean;
  title?: string;
  onClose: () => void;
  onSubmit: (content: string, timestamp?: string) => Promise<void>;
}

const AddNoteModal: React.FC<AddNoteModalProps> = ({ open, title, onClose, onSubmit }) => {
  const [form] = Form.useForm<{ content: string; timestamp?: dayjs.Dayjs }>();
  const [submitting, setSubmitting] = useState(false);

  const handleOk = () => form.submit();

  const handleFinish = (values: { content: string; timestamp?: dayjs.Dayjs }) => {
    setSubmitting(true);
    onSubmit(values.content, values.timestamp ? values.timestamp.format('YYYY-MM-DDTHH:mm:ssZ') : undefined)
      .then(() => {
        message.success('Note added');
        form.resetFields();
        onClose();
      })
      .catch((e) => message.error(e instanceof Error ? e.message : String(e)))
      .finally(() => setSubmitting(false));
  };

  return (
    <Modal
      title={title ?? 'Add Note'}
      open={open}
      onCancel={onClose}
      onOk={handleOk}
      okText="Add"
      confirmLoading={submitting}
      destroyOnClose
    >
      <Form form={form} layout="vertical" onFinish={handleFinish}>
        <Form.Item name="timestamp" label="Timestamp">
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item
          name="content"
          label="Content"
          rules={[{ required: true, message: 'Please enter note content' }]}
        >
          <Input.TextArea rows={4} placeholder="Note content" autoFocus />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default AddNoteModal;