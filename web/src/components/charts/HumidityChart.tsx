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
  humidity: number;
}

interface BaseChartProps {
  label: string;
  data: DataPoint[];
}

export default function HumidityChart({ data, label }: BaseChartProps) {
  Chart.register(
    CategoryScale,
    Title,
    Legend,
    LinearScale,
    PointElement,
    LineElement,
  );
  console.log(data);

  const chartData = {
    labels: data.map((x) => x.timestamp),
    datasets: [
      {
        label: label,
        data: data.map((x) => x.humidity),
        fill: true,
      },
    ],
  };

  const options = {
    responsive: true,
    borderColor: "rgba(0, 150, 199, 0.5)",
    backgroundColor: "rgba(72, 202, 228, 0.5)",
    scales: {
      y: {
        min: 0,
        max: 100,
        ticks: {
          callback: function (value: any, index: any, ticks: any) {
            return value + "%";
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
