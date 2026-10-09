import { Line } from "react-chartjs-2";
import {
  CategoryScale,
  LinearScale,
  Chart,
  Title,
  Legend,
  PointElement,
  LineElement,
} from "chart.js";

interface DataPoint {
  timestamp: string;
  ec: number;
}

interface BaseChartProps {
  label: string;
  data: DataPoint[];
}

export default function EcChart({ data, label }: BaseChartProps) {
  Chart.register(
    CategoryScale,
    Title,
    Legend,
    LinearScale,
    PointElement,
    LineElement,
  );

  const chartData = {
    labels: data.map((x) => x.timestamp),
    datasets: [
      {
        label: label,
        data: data.map((x) => x.ec),
        fill: true,
      },
    ],
  };

  const options = {
    responsive: true,
    borderColor: "rgba(72, 149, 83, 0.5)",
    backgroundColor: "rgba(122, 203, 133, 0.5)",
    scales: {
      y: {
        min: 0,
        max: 10,
        ticks: {
          callback: function (value: any, index: any, ticks: any) {
            return value + " mS/cm";
          },
        },
      },
    },
    plugins: {
      legend: {
        position: "top" as const,
      },
      title: {
        display: true,
        text: label,
      },
    },
  };

  return <Line options={options} data={chartData} />;
}