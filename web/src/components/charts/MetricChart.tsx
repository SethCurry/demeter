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
import { Alert, Button, DatePicker, Segmented, Space, Spin } from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import dayjs, { type Dayjs } from "dayjs";
import { useEffect, useMemo, useState } from "react";
import { MetricPoint } from "../../api";

export interface MetricChartQuery {
  start: string;
  end: string;
  limit?: number;
}

type PresetValue = "1h" | "6h" | "24h" | "7d" | "30d" | "custom";

const PRESET_HOURS: Record<Exclude<PresetValue, "custom">, number> = {
  "1h": 1,
  "6h": 6,
  "24h": 24,
  "7d": 24 * 7,
  "30d": 24 * 30,
};

const PRESET_OPTIONS: { label: string; value: PresetValue }[] = [
  { label: "1h", value: "1h" },
  { label: "6h", value: "6h" },
  { label: "24h", value: "24h" },
  { label: "7d", value: "7d" },
  { label: "30d", value: "30d" },
  { label: "Custom", value: "custom" },
];

// Aim for a readable number of points regardless of the selected window,
// while keeping at least one minute of resolution.
const TARGET_BUCKETS = 120;
const MIN_INTERVAL_SECONDS = 60;

function limitForRange(start: Dayjs, end: Dayjs): number {
  const totalSeconds = Math.max(1, end.diff(start, "second"));
  const intervalSeconds = Math.max(
    MIN_INTERVAL_SECONDS,
    Math.ceil(totalSeconds / TARGET_BUCKETS),
  );
  return Math.max(1, Math.ceil(totalSeconds / intervalSeconds));
}

function rangeForHours(hours: number): [Dayjs, Dayjs] {
  const end = dayjs();
  return [end.subtract(hours, "hour"), end];
}

interface MetricChartProps {
  label: string;
  loader: (query: MetricChartQuery) => Promise<MetricPoint[]>;
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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [preset, setPreset] = useState<PresetValue>("24h");
  const [range, setRange] = useState<[Dayjs, Dayjs]>(() => rangeForHours(24));
  const [reloadNonce, setReloadNonce] = useState(0);

  const query = useMemo<MetricChartQuery>(
    () => ({
      start: range[0].toISOString(),
      end: range[1].toISOString(),
      limit: limitForRange(range[0], range[1]),
    }),
    [range],
  );

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    loader(query)
      .then((newData) => {
        if (!cancelled) setData(newData);
      })
      .catch((err) => {
        if (!cancelled) {
          setData([]);
          setError(err instanceof Error ? err.message : String(err));
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [loader, query, reloadNonce]);

  const sortedData = useMemo(
    () =>
      [...data].sort(
        (a, b) => Date.parse(a.Timestamp) - Date.parse(b.Timestamp),
      ),
    [data],
  );

  const spanHours = range[1].diff(range[0], "hour", true);
  const formatLabel = (timestamp: string) => {
    const parsed = dayjs(timestamp);
    if (!parsed.isValid()) return timestamp;
    if (spanHours > 24 * 30) return parsed.format("MMM D");
    if (spanHours > 24) return parsed.format("MMM D HH:mm");
    return parsed.format("HH:mm");
  };

  const chartData = {
    labels: sortedData.map((x) => formatLabel(x.Timestamp)),
    datasets: [
      {
        label: label,
        data: sortedData.map((x) => x.Value),
        fill: true,
        borderColor: borderColor || "rgba(0, 150, 199, 0.5)",
        backgroundColor: backgroundColor || "rgba(72, 202, 228, 0.5)",
        pointRadius: 0,
      },
    ],
  };

  const defaultTickCallback = (value: any, index: any, ticks: any) => {
    return value;
  };

  const options = {
    responsive: true,
    scales: {
      x: {
        ticks: {
          autoSkip: true,
          maxRotation: 0,
          maxTicksLimit: 12,
        },
      },
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
      tooltip: {
        callbacks: {
          title: (items: { dataIndex: number }[]) => {
            const point = sortedData[items[0]?.dataIndex ?? -1];
            return point
              ? dayjs(point.Timestamp).format("YYYY-MM-DD HH:mm:ss")
              : "";
          },
        },
      },
    },
  };

  const handlePresetChange = (value: PresetValue) => {
    setPreset(value);

    if (value === "custom") {
      // Keep the current window and let the range picker refine it.
      return;
    }

    setRange(rangeForHours(PRESET_HOURS[value]));
  };

  const handleRangeChange = (values: [Dayjs | null, Dayjs | null] | null) => {
    if (!values?.[0] || !values[1] || !values[0].isBefore(values[1])) {
      return;
    }

    setPreset("custom");
    setRange([values[0], values[1]]);
  };

  const refresh = () => {
    if (preset !== "custom") {
      setRange(rangeForHours(PRESET_HOURS[preset]));
    }
    setReloadNonce((n) => n + 1);
  };

  return (
    <>
      <Space wrap size="small" style={{ marginBottom: 12 }}>
        <Segmented
          size="small"
          options={PRESET_OPTIONS}
          value={preset}
          onChange={(value) => handlePresetChange(value as PresetValue)}
        />
        <DatePicker.RangePicker
          size="small"
          showTime
          allowClear={false}
          format="YYYY-MM-DD HH:mm"
          value={range}
          onChange={handleRangeChange}
        />
        <Button
          size="small"
          icon={<ReloadOutlined />}
          onClick={refresh}
          title="Refresh chart"
          aria-label="Refresh chart"
        />
      </Space>

      {error && (
        <Alert
          type="error"
          showIcon
          message="Failed to load metric data"
          description={error}
          style={{ marginBottom: 12 }}
        />
      )}

      <Spin spinning={loading}>
        <Line options={options} data={chartData} />
      </Spin>
    </>
  );
}
