import React, { useEffect, useMemo, useRef, useState } from "react";
import { Empty, Typography } from "antd";
import Plotly from "plotly.js-dist-min";
import type { Config, Data, Layout } from "plotly.js-dist-min";
import { nullTime, type Flow, type Plant, type PlantSite } from "../api";

const { Text } = Typography;

/**
 * 3D Plotly scatter diagram of plant site positions.
 *
 * Site coordinates follow the database schema conventions:
 *   x: left -> right, y: bottom -> top, z: back -> front.
 *
 * Each flow gets its own colored legend entry. Sites occupied by at least
 * one plant are drawn as filled markers in the flow's color; empty sites
 * are drawn in gray. Hovering a marker shows the site id, position, flow,
 * and the plants living there.
 */

// Distinct palette cycled across flows so legend entries stay readable.
const FLOW_COLORS = ["rgba(52,78,65,0.95)"];

const EMPTY_COLOR = "rgba(100, 100, 100, 0.5)";

export interface PlantSites3DProps {
  /** Flows to render; each flow becomes one colored legend entry. */
  flows: Flow[];
  /** Plant sites belonging to those flows. */
  plantSites: PlantSite[];
  /** Plants used to determine which sites are occupied. */
  plants: Plant[];
  /** Plot height in pixels. */
  height?: number;
  /**
   * When set, the diagram switches to highlight mode: the site with this id
   * is drawn in red and every other site is drawn in grey, ignoring the
   * per-flow color palette. Used by the plant detail view.
   */
  highlightSiteId?: number;
}

const fmtPlanted = (p: Plant): string => {
  const t = nullTime(p.PlantedOn);
  return t ? new Date(t).toLocaleDateString() : "unknown date";
};

const HIGHLIGHT_COLOR = "rgba(52,78,65,0.95)";

const PlantSites3D: React.FC<PlantSites3DProps> = ({
  flows,
  plantSites,
  plants,
  height = 520,
  highlightSiteId,
}) => {
  const elRef = useRef<HTMLDivElement>(null);

  const [maxX, setMaxX] = useState(0);
  const [maxY, setMaxY] = useState(0);
  const [maxZ, setMaxZ] = useState(0);

  const traces = useMemo<Data[]>(() => {
    const flowById = new Map(flows.map((f) => [f.ID, f]));
    const plantsBySite = new Map<number, Plant[]>();
    plants.forEach((p) => {
      const list = plantsBySite.get(p.PlantSiteID);
      if (list) list.push(p);
      else plantsBySite.set(p.PlantSiteID, [p]);
    });

    const sitesByFlow = new Map<number, PlantSite[]>();
    plantSites.forEach((s) => {
      // Skip sites that reference flows we are not visualizing.
      if (!flowById.has(s.FlowID)) return;
      const list = sitesByFlow.get(s.FlowID);
      if (list) list.push(s);
      else sitesByFlow.set(s.FlowID, [s]);
    });

    const makeTrace = (
      flow: Flow,
      sites: PlantSite[],
      color: string,
      name: string,
      showlegend: boolean,
    ): Data => {
      const xs: number[] = [];
      const ys: number[] = [];
      const zs: number[] = [];
      const hovertext: string[] = [];
      sites.forEach((s) => {
        const sitePlants = plantsBySite.get(s.ID) ?? [];
        const plantInfo = sitePlants.length
          ? `Plants: ${sitePlants.map((p) => `#${p.ID} (planted ${fmtPlanted(p)})`).join(", ")}`
          : "Empty (no plants)";
        // Scale plotted coordinates by 4 so adjacent sites have visible
        // gaps between markers. Hover text still reports the true DB coords.
        const newX = s.X * 4;
        const newY = s.Y * 4;
        const newZ = s.Z * 4;

        xs.push(newX);
        if (newX > maxX) {
          setMaxX(newX);
        }
        ys.push(newY);
        if (newY > maxY) {
          setMaxY(newY);
        }
        zs.push(newZ);
        if (newZ > maxZ) {
          setMaxZ(newZ);
        }
        hovertext.push(
          `Site #${s.ID} \u2014 ${flow.Name}<br>Position: (${s.X}, ${s.Y}, ${s.Z})<br>${plantInfo}`,
        );
      });
      return {
        type: "scatter3d",
        mode: "markers",
        name,
        showlegend,
        x: xs,
        y: ys,
        z: zs,
        text: hovertext,
        hoverinfo: "text",
        marker: {
          size: 7,
          color,
          line: { color: "#ffffff", width: 1 },
        },
      };
    };

    const result: Data[] = [];

    if (highlightSiteId !== undefined) {
      // Highlight mode: a single red marker for the chosen site, everything
      // else grey. Legend has just two entries.
      const highlightFlow = plantSites.find(
        (s) => s.ID === highlightSiteId,
      )?.FlowID;
      const highlightFlowObj =
        highlightFlow != null ? flowById.get(highlightFlow) : undefined;
      const highlighted: PlantSite[] = [];
      const others: PlantSite[] = [];
      [...plantSites]
        .sort((a, b) => a.X - b.X || a.Y - b.Y || a.Z - b.Z)
        .forEach((s) => {
          if (s.ID === highlightSiteId) highlighted.push(s);
          else others.push(s);
        });
      const dummyFlow: Flow = highlightFlowObj ?? {
        ID: -1,
        Name: "Site",
        SystemID: 0,
        ParentFlowID: { Int64: 0, Valid: false },
      };
      result.push(
        makeTrace(dummyFlow, highlighted, HIGHLIGHT_COLOR, "This plant", true),
      );
      result.push(
        makeTrace(
          dummyFlow,
          others,
          EMPTY_COLOR,
          "Other sites",
          others.length > 0,
        ),
      );
      return result;
    }

    [...flows]
      .sort((a, b) => a.ID - b.ID)
      .forEach((flow, idx) => {
        const color = FLOW_COLORS[idx % FLOW_COLORS.length];
        const sites = (sitesByFlow.get(flow.ID) ?? []).sort(
          (a, b) => a.X - b.X || a.Y - b.Y || a.Z - b.Z,
        );
        const occupied = sites.filter((s) => plantsBySite.has(s.ID));
        const empty = sites.filter((s) => !plantsBySite.has(s.ID));
        result.push(
          makeTrace(flow, occupied, color, flow.Name, occupied.length > 0),
        );
        result.push(
          makeTrace(flow, empty, EMPTY_COLOR, `${flow.Name} (empty)`, false),
        );
      });
    return result;
  }, [flows, plantSites, plants, highlightSiteId, maxX, maxY, maxZ]);

  // Tear down the plot (and its event listeners) only on unmount.
  useEffect(() => {
    const el = elRef.current;
    return () => {
      if (el) Plotly.purge(el);
    };
  }, []);

  useEffect(() => {
    const el = elRef.current;
    if (!el || plantSites.length === 0) return;

    const layout: Partial<Layout> = {
      margin: { l: 0, r: 0, b: 0, t: 0 },
      showlegend: flows.length > 1 || highlightSiteId !== undefined,
      hovermode: "closest",
      // Preserve the user's camera angle/zoom across data refreshes.
      uirevision: "plant-sites-3d",
      scene: {
        aspectmode: "data",
        camera: {
          eye: { x: 0, y: 0, z: -2 },
          up: { x: 0, y: 1, z: 0 },
          //center: { x: maxX / 2, y: maxY / 2, z: maxZ / 2 },
        },
        bgcolor: "#a3b18a",
        //rgba(52,78,65,0.95)
        xaxis: { title: { text: "X (left \u2192 right)" } },
        yaxis: { title: { text: "Y (bottom \u2192 top)" } },
        zaxis: { title: { text: "Z (back \u2192 front)" } },
      },
    };

    const config: Partial<Config> = {
      responsive: true,
      displaylogo: false,
    };

    // Plotly.react diffs efficiently, so refreshing data keeps the scene.
    Plotly.react(el, traces, layout, config).catch(() => {
      /* render can race with unmount; nothing to do */
    });
  }, [traces, plantSites.length, flows.length, highlightSiteId]);

  if (plantSites.length === 0) {
    return <Empty description="No plant sites to visualize" />;
  }

  return (
    <div>
      <div ref={elRef} style={{ width: "100%", height }} />
      <Text type="secondary" style={{ display: "block", marginTop: 8 }}>
        {highlightSiteId !== undefined
          ? "Drag to rotate, scroll to zoom, double-click to reset the view. The red marker is this plant\u2019s site; all other sites are shown in grey."
          : "Drag to rotate, scroll to zoom, double-click to reset the view. Colored markers hold plants; gray markers are empty sites. Each color is a separate flow."}
      </Text>
    </div>
  );
};

export default PlantSites3D;
