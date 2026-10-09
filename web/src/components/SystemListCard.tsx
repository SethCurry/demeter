import { Card, Table } from "antd";
import { useNavigate } from "react-router-dom";
import { type System } from "../api";

interface SystemListCardProps {
  systems: System[];
}

export default function SystemListCard({ systems }: SystemListCardProps) {
  const navigate = useNavigate();
  const systemColumns = [
    { title: "ID", dataIndex: "ID", key: "id", width: 60 },
    {
      title: "Name",
      dataIndex: "Name",
      key: "name",
      render: (name: string, r: System) => (
        <a onClick={() => navigate(`/systems/${r.ID}`)}>{name}</a>
      ),
    },
  ];

  return (
    <Card size="small" title={`Systems (${systems.length})`}>
      <Table
        rowKey="ID"
        columns={systemColumns}
        dataSource={systems}
        pagination={false}
        size="small"
        locale={{ emptyText: "No systems in this enclosure" }}
      />
    </Card>
  );
}
