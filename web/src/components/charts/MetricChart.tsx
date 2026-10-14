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
import { MetricPoint } from "../../api";
import { useEffect, useState } from "react";

interface MetricChartProps {
  label: string;
  loader: () => Promise<MetricPoint[]>;
  borderColor?: string;
  backgroundColor?: string;
  tickCallback?: (value: any, index: any, ticks: any) => any;
  minY?: number;
  maxY?: number;
}

export default function MetricChart({
  borderColor,
  backgroundColor,
  tickCallback,
  loader,
  label,
  minY,
  maxY,
}: MetricChartProps) {
  Chart.register(
    CategoryScale,
    Title,
    Legend,
    LinearScale,
    PointElement,
    LineElement,
  );

  const [data, setData] = useState<MetricPoint[]>([]);

  useEffect(() => {
    loader().then((newData) => {
      setData(newData);
    });
  }, [loader, setData]);

  const chartData = {
    labels: data.map((x) => x.Timestamp),
    datasets: [
      {
        label: label,
        data: data.map((x) => x.Value),
        fill: true,
      },
    ],
  };

  const defaultTickCallback = (value: any, index: any, ticks: any) => {
    return value;
  };

  const options = {
    responsive: true,
    borderColor: borderColor || "rgba(0, 150, 199, 0.5)",
    backgroundColor: backgroundColor || "rgba(72, 202, 228, 0.5)",
    scales: {
      y: {
        min: minY,
        max: maxY,
        ticks: {
          callback: tickCallback || defaultTickCallback,
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
