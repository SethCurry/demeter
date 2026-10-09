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
  ph: number;
}

interface BaseChartProps {
  label: string;
  data: DataPoint[];
}

export default function PhChart({ data, label }: BaseChartProps) {
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
        data: data.map((x) => x.ph),
        fill: false,
      },
    ],
  };

  const options = {
    responsive: true,
    borderColor: "rgba(114, 87, 220, 0.5)",
    backgroundColor: "rgba(158, 123, 255, 0.5)",
    scales: {
      y: {
        min: 0,
        max: 14,
        ticks: {
          callback: function (value: any, index: any, ticks: any) {
            return value + " pH";
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