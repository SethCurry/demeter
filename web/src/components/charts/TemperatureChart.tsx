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
import { useState } from "react";

interface DataPoint {
  timestamp: string;
  temperature: number;
}

interface BaseChartProps {
  label: string;
  data: DataPoint[];
}

export default function TemperatureChart({ data, label }: BaseChartProps) {
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
        data: data.map((x) => x.temperature),
        fill: false,
      },
    ],
  };

  const options = {
    responsive: true,
    borderColor: "rgba(249, 65, 68, 0.5)",
    backgroundColor: "rgba(249, 132, 74, 0.5)",
    scales: {
      y: {
        min: 0,
        max: 50,
        ticks: {
          callback: function (value: any, index: any, ticks: any) {
            return value + " C";
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
