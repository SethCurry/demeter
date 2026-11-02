import { Card, Table } from "antd";
import { Link } from "react-router-dom";
import { type System } from "../api";

interface SystemListCardProps {
  systems: System[];
}

export default function SystemListCard({ systems }: SystemListCardProps) {
  const systemColumns = [
    { title: "ID", dataIndex: "ID", key: "id", width: 60 },
    {
      title: "Name",
      dataIndex: "Name",
      key: "name",
      render: (name: string, r: System) => (
        <Link to={`/systems/${r.ID}`}>{name}</Link>
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
