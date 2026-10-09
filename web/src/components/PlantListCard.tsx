import { Card, Table } from "antd";
import { nullString, nullTime, type Plant, type PlantWithFlow } from "../api";
import { useNavigate } from "react-router-dom";

interface PlantListCardProps {
  plants: PlantWithFlow[];
}

export default function PlantListCard({ plants }: PlantListCardProps) {
  const navigate = useNavigate();

  const plantColumns = [
    { title: "Plant ID", dataIndex: "ID", key: "id", width: 80 },
    {
      title: "Site",
      key: "site",
      render: (_: unknown, r: PlantWithFlow) => {
        return `#${r.PlantSiteID} (${r.X.Int64},${r.Y.Int64},${r.Z.Int64})`;
      },
    },
    {
      title: "Flow",
      key: "flow",
      render: (_: unknown, r: PlantWithFlow) => {
        return (
          <a href={`/flows/${r.FlowID}`} onClick={() => navigate(`/flows/${r.FlowID}`)}>
            {r.FlowName}
          </a>
        );
      },
    },
    {
      title: "Species",
      key: "species",
      render: (_: unknown, r: PlantWithFlow) => {
          var speciesId = r.SpeciesID;
          const speciesName = r.SpeciesName;
          const genusName = r.GenusName;
        return (
          <a
              href={`/plant-species/${speciesId}`}
            onClick={() => navigate(`/plant-species/${speciesId}`)}
          >
            <em>{genusName}</em> {speciesName}
          </a>
        );
      },
    },
    {
      title: "Planted",
      key: "planted",
      render: (_: unknown, r: Plant) => {
        const planted = nullTime(r.PlantedOn);
        return planted ? new Date(planted).toLocaleDateString() : "—";
      },
    },
  ];

  return (
    <Card size="small" title={`Plants (${plants.length})`}>
      <Table
        rowKey="ID"
        columns={plantColumns}
        dataSource={plants}
        pagination={false}
        size="small"
        locale={{ emptyText: "No plants in this enclosure" }}
      />
    </Card>
  );
}
