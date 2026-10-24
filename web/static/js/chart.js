// A small dependency-free canvas chart: multi-series line/area charts with
// time axis, crosshair tooltip, HiDPI rendering and theme-aware colours.

const TIME_STEPS = [
  60e3, 5 * 60e3, 15 * 60e3, 30 * 60e3, 3600e3, 3 * 3600e3, 6 * 3600e3, 12 * 3600e3,
  86400e3, 2 * 86400e3, 7 * 86400e3, 14 * 86400e3, 30 * 86400e3, 91 * 86400e3, 182 * 86400e3, 365 * 86400e3,
];

function cssVar(el, name, fallback) {
  const v = getComputedStyle(el).getPropertyValue(name).trim();
  return v || fallback;
}

function niceStep(range, count) {
  const raw = range / Math.max(1, count);
  const mag = Math.pow(10, Math.floor(Math.log10(raw)));
  const norm = raw / mag;
  const step = norm < 1.5 ? 1 : norm < 3 ? 2 : norm < 7 ? 5 : 10;
  return step * mag;
}

// Index of the last point with t <= x (binary search).
function bisect(points, x) {
  let lo = 0, hi = points.length - 1;
  if (hi < 0 || points[0][0] > x) return -1;
  while (lo < hi) {
    const mid = (lo + hi + 1) >> 1;
    if (points[mid][0] <= x) lo = mid; else hi = mid - 1;
  }
  return lo;
}

export class Chart {
  constructor(canvas, opts = {}) {
    this.canvas = canvas;
    this.ctx = canvas.getContext("2d");
    this.series = [];
    this.opts = Object.assign({
      sparkline: false,
      area: true,
      percent: false,
      formatValue: (v) => v.toFixed(2),
      formatTime: (t) => new Date(t).toLocaleTimeString(),
      formatTick: null,
      emptyText: "",
    }, opts);
    this.hover = null;
    this.pending = false;
    this.ro = new ResizeObserver(() => this.draw());
    this.ro.observe(canvas);
    if (!this.opts.sparkline) {
      const move = (e) => {
        const r = canvas.getBoundingClientRect();
        this.hover = { x: e.clientX - r.left, y: e.clientY - r.top };
        this.draw();
      };
      canvas.addEventListener("pointermove", move);
      canvas.addEventListener("pointerdown", move);
      canvas.addEventListener("pointerleave", () => { this.hover = null; this.draw(); });
    }
  }

  destroy() {
    this.ro.disconnect();
  }

  setOptions(o) {
    Object.assign(this.opts, o);
    this.draw();
  }

  // series: [{ id, label, color, points: [[t, v], ...], width }]
  setSeries(series) {
    this.series = series.map((s) => ({ ...s, points: (s.points || []).filter((p) => p[1] != null && isFinite(p[1])) }));
    this.draw();
  }

  // append adds a live point to a series, keeping the time order.
  append(id, t, v) {
    const s = this.series.find((x) => x.id === id);
    if (!s || v == null || !isFinite(v)) return;
    const last = s.points[s.points.length - 1];
    if (last && t <= last[0]) {
      if (t === last[0]) last[1] = v;
      return;
    }
    s.points.push([t, v]);
    this.draw();
  }

  draw() {
    if (this.pending) return;
    this.pending = true;
    requestAnimationFrame(() => {
      this.pending = false;
      this.render();
    });
  }

  transformed() {
    if (!this.opts.percent) return this.series;
    return this.series.map((s) => {
      const base = s.points.length ? s.points[0][1] : 0;
      return { ...s, points: base ? s.points.map(([t, v]) => [t, (v / base - 1) * 100]) : [] };
    });
  }

  render() {
    const { canvas, ctx, opts } = this;
    const dpr = window.devicePixelRatio || 1;
    const w = canvas.clientWidth, h = canvas.clientHeight;
    if (!w || !h) return;
    if (canvas.width !== Math.round(w * dpr) || canvas.height !== Math.round(h * dpr)) {
      canvas.width = Math.round(w * dpr);
      canvas.height = Math.round(h * dpr);
    }
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, w, h);

    const text = cssVar(canvas, "--muted", "#888");
    const grid = cssVar(canvas, "--grid", "rgba(128,128,128,.15)");
    const fg = cssVar(canvas, "--text", "#222");
    const surface = cssVar(canvas, "--surface", "#fff");
    const font = getComputedStyle(canvas).fontFamily || "sans-serif";
    const fs = Math.max(10, Math.round(parseFloat(getComputedStyle(document.documentElement).fontSize) * 0.75));

    const series = this.transformed().filter((s) => s.points.length > 0);
    if (!series.length) {
      if (opts.emptyText && !opts.sparkline) {
        ctx.fillStyle = text;
        ctx.font = `${fs + 2}px ${font}`;
        ctx.textAlign = "center";
        ctx.fillText(opts.emptyText, w / 2, h / 2);
      }
      return;
    }

    let tMin = Infinity, tMax = -Infinity, vMin = Infinity, vMax = -Infinity;
    for (const s of series) {
      for (const [t, v] of s.points) {
        if (t < tMin) tMin = t;
        if (t > tMax) tMax = t;
        if (v < vMin) vMin = v;
        if (v > vMax) vMax = v;
      }
    }
    if (opts.from && opts.from < tMin) tMin = opts.from;
    if (tMax === tMin) tMax = tMin + 60e3;
    if (vMax === vMin) {
      const d = Math.abs(vMax) * 0.001 || 1;
      vMax += d;
      vMin -= d;
    }
    const pad = (vMax - vMin) * 0.08;
    vMax += pad;
    vMin -= opts.sparkline ? pad * 0.3 : pad;

    // Layout.
    ctx.font = `${fs}px ${font}`;
    let left = 0, right = 0, top = 4, bottom = 2;
    let yStep = 0, yTicks = [];
    if (!opts.sparkline) {
      yStep = niceStep(vMax - vMin, Math.max(2, Math.floor(h / 55)));
      for (let v = Math.ceil(vMin / yStep) * yStep; v <= vMax; v += yStep) yTicks.push(v);
      const fmtY = (v) => (opts.percent ? (v > 0 ? "+" : "") + v.toFixed(Math.abs(yStep) < 1 ? 2 : 1) + "%" : opts.formatValue(v, yStep));
      const labelW = Math.max(...yTicks.map((v) => ctx.measureText(fmtY(v)).width));
      right = labelW + 14;
      left = 8;
      top = 10;
      bottom = fs + 14;
      this.fmtY = fmtY;
    }
    const pw = w - left - right, ph = h - top - bottom;
    const x = (t) => left + ((t - tMin) / (tMax - tMin)) * pw;
    const y = (v) => top + (1 - (v - vMin) / (vMax - vMin)) * ph;
    this.geom = { x, y, left, right, top, bottom, pw, ph, tMin, tMax, w, h };

    if (!opts.sparkline) {
      // Horizontal grid + y labels on the right.
      ctx.strokeStyle = grid;
      ctx.lineWidth = 1;
      ctx.fillStyle = text;
      ctx.textAlign = "left";
      ctx.textBaseline = "middle";
      for (const v of yTicks) {
        const yy = Math.round(y(v)) + 0.5;
        ctx.beginPath();
        ctx.moveTo(left, yy);
        ctx.lineTo(left + pw, yy);
        ctx.stroke();
        ctx.fillText(this.fmtY(v), left + pw + 8, yy);
      }
      // Time axis.
      const span = tMax - tMin;
      const maxTicks = Math.max(2, Math.floor(pw / (fs * 7)));
      const step = TIME_STEPS.find((s) => span / s <= maxTicks) || TIME_STEPS[TIME_STEPS.length - 1];
      const tz = new Date().getTimezoneOffset() * 60e3;
      ctx.textAlign = "center";
      ctx.textBaseline = "top";
      let t0 = step >= 86400e3 ? Math.ceil((tMin - tz) / step) * step + tz : Math.ceil(tMin / step) * step;
      for (let t = t0; t <= tMax; t += step) {
        const xx = Math.round(x(t)) + 0.5;
        ctx.beginPath();
        ctx.moveTo(xx, top);
        ctx.lineTo(xx, top + ph);
        ctx.stroke();
        const label = opts.formatTick ? opts.formatTick(t, step) : new Date(t).toLocaleString();
        ctx.fillText(label, xx, top + ph + 6);
      }
    }

    // Series.
    ctx.save();
    ctx.beginPath();
    ctx.rect(left, 0, pw, h);
    ctx.clip();
    const single = series.length === 1;
    for (const s of series) {
      const pts = s.points;
      ctx.beginPath();
      pts.forEach(([t, v], i) => (i ? ctx.lineTo(x(t), y(v)) : ctx.moveTo(x(t), y(v))));
      if (opts.area && single) {
        const g = ctx.createLinearGradient(0, top, 0, top + ph);
        g.addColorStop(0, hexAlpha(s.color, 0.32));
        g.addColorStop(1, hexAlpha(s.color, 0));
        ctx.save();
        ctx.lineTo(x(pts[pts.length - 1][0]), top + ph);
        ctx.lineTo(x(pts[0][0]), top + ph);
        ctx.closePath();
        ctx.fillStyle = g;
        ctx.fill();
        ctx.restore();
        ctx.beginPath();
        pts.forEach(([t, v], i) => (i ? ctx.lineTo(x(t), y(v)) : ctx.moveTo(x(t), y(v))));
      }
      ctx.strokeStyle = s.color;
      ctx.lineWidth = s.width || (opts.sparkline ? 1.6 : 2);
      ctx.lineJoin = "round";
      ctx.stroke();
    }
    ctx.restore();

    if (opts.sparkline) {
      const s = series[0];
      const [lt, lv] = s.points[s.points.length - 1];
      ctx.fillStyle = s.color;
      ctx.beginPath();
      ctx.arc(Math.min(x(lt), w - 3), y(lv), 2.6, 0, Math.PI * 2);
      ctx.fill();
      return;
    }

    // Last-value badges on the y axis.
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";
    for (const s of series) {
      const [, lv] = s.points[s.points.length - 1];
      const label = this.fmtY(lv);
      const yy = Math.min(Math.max(y(lv), top + fs / 2), top + ph - fs / 2);
      const tw = ctx.measureText(label).width;
      ctx.fillStyle = s.color;
      roundRect(ctx, left + pw + 3, yy - fs / 2 - 3, tw + 10, fs + 6, 4);
      ctx.fill();
      ctx.fillStyle = contrast(s.color);
      ctx.fillText(label, left + pw + 8, yy);
    }

    // Crosshair + tooltip.
    if (this.hover && this.hover.x >= left && this.hover.x <= left + pw) {
      const t = tMin + ((this.hover.x - left) / pw) * (tMax - tMin);
      const rows = [];
      let snapX = null;
      for (const s of series) {
        const i = bisect(s.points, t);
        let j = i;
        if (i < 0) j = 0;
        else if (i + 1 < s.points.length && Math.abs(s.points[i + 1][0] - t) < Math.abs(s.points[i][0] - t)) j = i + 1;
        const p = s.points[j];
        if (!p) continue;
        if (snapX === null) snapX = p[0];
        rows.push({ s, p });
        ctx.fillStyle = s.color;
        ctx.beginPath();
        ctx.arc(x(p[0]), y(p[1]), 3.5, 0, Math.PI * 2);
        ctx.fill();
      }
      if (snapX !== null) {
        const xx = Math.round(x(snapX)) + 0.5;
        ctx.strokeStyle = text;
        ctx.setLineDash([3, 3]);
        ctx.beginPath();
        ctx.moveTo(xx, top);
        ctx.lineTo(xx, top + ph);
        ctx.stroke();
        ctx.setLineDash([]);

        const lines = [opts.formatTime(snapX)].concat(rows.map((r) => `${r.s.label}: ${this.fmtY(r.p[1])}`));
        ctx.font = `${fs + 1}px ${font}`;
        const bw = Math.max(...lines.map((l) => ctx.measureText(l).width)) + 34;
        const lh = fs + 7;
        const bh = lines.length * lh + 10;
        let bx = xx + 14;
        if (bx + bw > left + pw) bx = xx - 14 - bw;
        let by = Math.min(Math.max(this.hover.y - bh / 2, top), top + ph - bh);
        ctx.fillStyle = hexAlpha(surface, 0.94);
        ctx.strokeStyle = grid;
        roundRect(ctx, bx, by, bw, bh, 8);
        ctx.fill();
        ctx.stroke();
        ctx.textBaseline = "middle";
        lines.forEach((l, i) => {
          const ly = by + 5 + lh * i + lh / 2;
          if (i === 0) {
            ctx.fillStyle = text;
            ctx.fillText(l, bx + 10, ly);
          } else {
            ctx.fillStyle = rows[i - 1].s.color;
            ctx.fillRect(bx + 10, ly - 4, 8, 8);
            ctx.fillStyle = fg;
            ctx.fillText(l, bx + 24, ly);
          }
        });
      }
    }
  }
}

function roundRect(ctx, x, y, w, h, r) {
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.arcTo(x + w, y, x + w, y + h, r);
  ctx.arcTo(x + w, y + h, x, y + h, r);
  ctx.arcTo(x, y + h, x, y, r);
  ctx.arcTo(x, y, x + w, y, r);
  ctx.closePath();
}

function parseColor(c) {
  c = String(c).trim();
  if (c.startsWith("#")) {
    let h = c.slice(1);
    if (h.length === 3) h = h.split("").map((x) => x + x).join("");
    const n = parseInt(h.slice(0, 6), 16);
    return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
  }
  const m = c.match(/rgba?\(([^)]+)\)/);
  if (m) return m[1].split(",").slice(0, 3).map((x) => parseFloat(x));
  return [128, 128, 128];
}

export function hexAlpha(c, a) {
  const [r, g, b] = parseColor(c);
  return `rgba(${r},${g},${b},${a})`;
}

function contrast(c) {
  const [r, g, b] = parseColor(c);
  return (r * 299 + g * 587 + b * 114) / 1000 > 150 ? "#111" : "#fff";
}
