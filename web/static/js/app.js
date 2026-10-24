import { t, setLang, detectLang, getLang, LANGS, locale } from "./i18n.js";
import { Chart } from "./chart.js";

// ---- State ----------------------------------------------------------------------

const PALETTE = ["#6c5ce7", "#f59e0b", "#10b981", "#ef4444", "#3b82f6", "#ec4899", "#14b8a6", "#a855f7",
  "#f97316", "#84cc16", "#06b6d4", "#e11d48", "#64748b", "#eab308"];
const RANGES = ["1h", "6h", "24h", "7d", "30d", "90d", "1y", "5y"];
const THEMES = ["system", "light", "dark", "midnight", "ocean", "sepia"];
const THEME_PREVIEW = {
  system: ["#f3f4f8", "#0e1117", "#5b4bdb"], light: ["#f3f4f8", "#ffffff", "#5b4bdb"], dark: ["#0e1117", "#161a23", "#8d7dff"],
  midnight: ["#000000", "#0a0a0c", "#00d1ff"], ocean: ["#0a1626", "#0f1f35", "#2dd4bf"], sepia: ["#f2e9d8", "#fbf5e9", "#b05e1e"],
};
const CONDITIONS = ["rise_pct", "fall_pct", "move_pct", "rise_abs", "fall_abs", "above", "below"];
const WINDOWS = ["since", "5m", "15m", "1h", "4h", "24h", "7d"];

const S = {
  meta: null,
  settings: null,
  views: new Map(),
  services: [],
  rules: [],
  events: [],
  es: null,
  page: null,
  sourceInfo: new Map(), // id -> { name, type, url, fee, color }
  stopped: false,
};

const $ = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));
const esc = (s) => String(s ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]);

// ---- API ---------------------------------------------------------------------------

async function apiGet(path) {
  const r = await fetch(path, { cache: "no-store" });
  if (!r.ok) throw new Error((await r.json().catch(() => null))?.error || r.statusText);
  return r.json();
}

async function apiSend(method, path, body) {
  const r = await fetch(path, {
    method,
    headers: { "Content-Type": "application/json", "X-Cryptomon": "1" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!r.ok) throw new Error((await r.json().catch(() => null))?.error || r.statusText);
  return r.status === 204 ? null : r.json();
}

// ---- Formatting ------------------------------------------------------------------

const nfCache = new Map();
function nf(opts) {
  const key = locale() + JSON.stringify(opts);
  let f = nfCache.get(key);
  if (!f) {
    try {
      f = new Intl.NumberFormat(locale(), opts);
    } catch {
      f = new Intl.NumberFormat("en-US", opts);
    }
    nfCache.set(key, f);
  }
  return f;
}

function priceOpts(v) {
  const a = Math.abs(v);
  if (a >= 100000) return { minimumFractionDigits: 0, maximumFractionDigits: 0 };
  if (a >= 100) return { minimumFractionDigits: 2, maximumFractionDigits: 2 };
  if (a >= 1) return { minimumFractionDigits: 2, maximumFractionDigits: 4 };
  return { minimumSignificantDigits: 3, maximumSignificantDigits: 5 };
}

function fmtPrice(v, vs) {
  if (v == null || !isFinite(v) || v === 0) return "—";
  return nf({ style: "currency", currency: vs, currencyDisplay: "narrowSymbol", ...priceOpts(v) }).format(v);
}

function fmtNum(v, opts) {
  if (v == null || !isFinite(v)) return "—";
  return nf(opts || priceOpts(v)).format(v);
}

function fmtPct(v, digits = 2) {
  if (v == null || !isFinite(v)) return "—";
  return nf({ minimumFractionDigits: digits, maximumFractionDigits: digits, signDisplay: "exceptZero" }).format(v) + "%";
}

function fmtVolume(v) {
  if (!v) return "—";
  return nf({ notation: "compact", maximumFractionDigits: 2 }).format(v);
}

function timeAgo(ms) {
  if (!ms) return "—";
  const s = Math.max(0, Math.round((Date.now() - ms) / 1000));
  if (s < 5) return t("time.now");
  if (s < 60) return t("time.sec", { n: s });
  if (s < 3600) return t("time.min", { n: Math.round(s / 60) });
  if (s < 86400) return t("time.hour", { n: Math.round(s / 3600) });
  return t("time.day", { n: Math.round(s / 86400) });
}

const dfCache = new Map();
function df(opts) {
  const key = locale() + JSON.stringify(opts);
  if (!dfCache.has(key)) dfCache.set(key, new Intl.DateTimeFormat(locale(), opts));
  return dfCache.get(key);
}
const fmtDateTime = (ms) => df({ dateStyle: "medium", timeStyle: "short" }).format(ms);
function fmtTick(ms, step) {
  if (step < 86400e3) return df({ hour: "2-digit", minute: "2-digit" }).format(ms);
  if (step < 30 * 86400e3) return df({ day: "numeric", month: "short" }).format(ms);
  return df({ month: "short", year: "numeric" }).format(ms);
}
function axisFormatter(v, step) {
  const d = step ? Math.min(8, Math.max(0, -Math.floor(Math.log10(step)))) : 2;
  return nf({ minimumFractionDigits: d, maximumFractionDigits: d }).format(v);
}

function changeClass(v) {
  return v == null ? "" : v > 0 ? "up" : v < 0 ? "down" : "";
}

function assetName(code) {
  return t("asset." + code);
}

function sourceName(id) {
  return S.sourceInfo.get(id)?.name || id;
}

function sourceColor(id, code) {
  if (id === "agg") return S.meta?.assets.find((a) => a.code === code)?.color || "#6c5ce7";
  return S.sourceInfo.get(id)?.color || "#888";
}

function windowText(w) {
  if (w === "since") return t("msg.since");
  const unit = { m: "unit.min", h: "unit.h", d: "unit.d" }[w.slice(-1)];
  return t("msg.in", { w: w.slice(0, -1) + " " + t(unit) });
}

// alertMessage renders an alert event in the current language.
function alertMessage(ev) {
  if (ev.condition === "above" || ev.condition === "below") {
    return t("msg." + ev.condition, { asset: ev.asset, level: fmtPrice(ev.value, ev.vs) });
  }
  let key = "move";
  if (ev.condition === "rise_pct" || ev.condition === "rise_abs" || (ev.condition === "move_pct" && ev.changePct > 0)) key = "up";
  if (ev.condition === "fall_pct" || ev.condition === "fall_abs" || (ev.condition === "move_pct" && ev.changePct < 0)) key = "down";
  // "rose by"/"fell by" already carry the direction, so the number is unsigned.
  const pct = Math.abs(ev.changePct);
  let chg = fmtPct(pct, pct < 0.1 ? 4 : 2).replace("+", "");
  if (ev.condition === "rise_abs" || ev.condition === "fall_abs") chg = fmtPrice(Math.abs(ev.changeAbs), ev.vs);
  return t("msg." + key, { asset: ev.asset, chg, window: windowText(ev.window) });
}

// ---- Settings, theme, scale ---------------------------------------------------------

function cacheUI() {
  try {
    localStorage.setItem("cryptomon.ui", JSON.stringify({ theme: S.settings.theme, scale: S.settings.scale, dir: document.documentElement.dir }));
  } catch {}
}

function resolvedTheme(theme) {
  if (theme !== "system") return theme;
  return matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function applyTheme() {
  document.documentElement.dataset.theme = resolvedTheme(S.settings.theme);
  cacheUI();
  redrawCharts();
}

function applyScale() {
  const d = document.documentElement;
  if (S.settings.scale) {
    d.dataset.scale = "fixed";
    d.style.setProperty("--scale", S.settings.scale);
  } else {
    delete d.dataset.scale;
  }
  cacheUI();
  redrawCharts();
}

const charts = new Set();
function redrawCharts() {
  requestAnimationFrame(() => charts.forEach((c) => c.draw()));
}
function makeChart(canvas, opts) {
  const c = new Chart(canvas, { formatTick: fmtTick, formatTime: fmtDateTime, formatValue: axisFormatter, ...opts });
  charts.add(c);
  return c;
}
function dropChart(c) {
  if (!c) return;
  c.destroy();
  charts.delete(c);
}

let saveTimer = null;
function saveSettings(patch, { reconnect = false } = {}) {
  Object.assign(S.settings, patch);
  cacheUI();
  clearTimeout(saveTimer);
  saveTimer = setTimeout(async () => {
    try {
      S.settings = await apiSend("PUT", "/api/settings", S.settings);
    } catch (e) {
      toast(t("toast.error", { e: e.message }), "error");
    }
  }, 250);
  if (reconnect) {
    S.views.clear();
    connect();
    render();
  }
}

function applyLang() {
  const lang = S.settings.lang || detectLang();
  setLang(lang);
  nfCache.clear();
  dfCache.clear();
  cacheUI();
  $$("[data-i18n]").forEach((el) => (el.textContent = t(el.dataset.i18n)));
  $("#themeBtn").title = t("settings.theme");
  $("#vsSelect").title = t("settings.currency");
  $("#langSelect").title = t("settings.language");
  fillHeaderSelects();
  setConn(S.connState || "waiting");
}

function fillHeaderSelects() {
  const vs = $("#vsSelect");
  vs.innerHTML = S.meta.fiat.map((c) => `<option value="${c}" title="${esc(assetName(c))}">${c}</option>`).join("");
  vs.value = S.settings.vs;
  const ls = $("#langSelect");
  ls.innerHTML = LANGS.map((l) => `<option value="${l.code}">${esc(l.name)}</option>`).join("");
  ls.value = getLang();
}

// ---- Live stream --------------------------------------------------------------------

function setConn(state) {
  S.connState = state;
  $("#connDot").className = "dot " + state;
  $("#connText").textContent = t(state === "up" ? "conn.live" : state === "down" ? "conn.offline" : "conn.connecting");
}

function connect() {
  if (S.es) S.es.close();
  setConn("waiting");
  const es = new EventSource("/api/stream?vs=" + encodeURIComponent(S.settings.vs));
  S.es = es;
  es.onopen = () => setConn("up");
  es.onerror = () => {
    if (!S.stopped) setConn("down");
  };
  es.addEventListener("views", (e) => {
    for (const v of JSON.parse(e.data)) S.views.set(v.code, v);
    S.page?.onViews?.();
  });
  es.addEventListener("view", (e) => {
    const v = JSON.parse(e.data);
    const prev = S.views.get(v.code);
    S.views.set(v.code, v);
    S.page?.onView?.(v, prev);
  });
  es.addEventListener("status", (e) => {
    S.services = JSON.parse(e.data);
    S.page?.onStatus?.();
  });
  es.addEventListener("rules", (e) => {
    S.rules = JSON.parse(e.data);
    updateBadge();
    S.page?.onRules?.();
  });
  es.addEventListener("alert", (e) => onAlert(JSON.parse(e.data)));
  es.addEventListener("bye", () => showStopped());
}

function updateBadge() {
  const n = S.rules.filter((r) => r.enabled).length;
  const b = $("#alertBadge");
  b.textContent = n;
  b.classList.toggle("hidden", n === 0);
}

function onAlert(ev) {
  S.events.unshift(ev);
  const text = alertMessage(ev);
  const title = `${ev.asset}/${ev.vs}${ev.source && ev.source !== "agg" ? " · " + sourceName(ev.source) : ""}`;
  toast(`<b>${esc(title)}</b>${esc(text)}<br><span class="muted">${esc(fmtPrice(ev.price, ev.vs))}</span>`, "alert", 12000, true);
  if (S.settings.sound) beep();
  if (S.settings.browserNotify && "Notification" in window && Notification.permission === "granted") {
    try {
      new Notification("Cryptomon · " + title, { body: text + "\n" + fmtPrice(ev.price, ev.vs), tag: ev.id });
    } catch {}
  }
  S.page?.onEvent?.();
}

let audioCtx = null;
function beep() {
  try {
    audioCtx = audioCtx || new (window.AudioContext || window.webkitAudioContext)();
    const now = audioCtx.currentTime;
    [880, 1320].forEach((f, i) => {
      const o = audioCtx.createOscillator();
      const g = audioCtx.createGain();
      o.frequency.value = f;
      o.type = "sine";
      g.gain.setValueAtTime(0.0001, now + i * 0.16);
      g.gain.exponentialRampToValueAtTime(0.18, now + i * 0.16 + 0.02);
      g.gain.exponentialRampToValueAtTime(0.0001, now + i * 0.16 + 0.15);
      o.connect(g).connect(audioCtx.destination);
      o.start(now + i * 0.16);
      o.stop(now + i * 0.16 + 0.16);
    });
  } catch {}
}

function toast(html, kind = "", ms = 4000, isHTML = false) {
  const el = document.createElement("div");
  el.className = "toast " + kind;
  if (isHTML) el.innerHTML = html;
  else el.textContent = html;
  $("#toasts").appendChild(el);
  setTimeout(() => el.remove(), ms);
  el.addEventListener("click", () => el.remove());
}

function showStopped() {
  S.stopped = true;
  S.es?.close();
  setConn("down");
  S.page?.destroy?.();
  S.page = null;
  $("#view").innerHTML = `<div class="stopped">${esc(t("settings.stopped"))}</div>`;
}

// ---- Router -------------------------------------------------------------------------------

function parseRoute() {
  const parts = (location.hash.replace(/^#\/?/, "") || "").split("/").filter(Boolean);
  const name = parts[0] || "dashboard";
  return { name, arg: parts[1] ? decodeURIComponent(parts[1]).toUpperCase() : null };
}

function render() {
  if (S.stopped) return;
  const r = parseRoute();
  S.page?.destroy?.();
  const pages = { dashboard: Dashboard, asset: AssetPage, markets: MarketsPage, alerts: AlertsPage, settings: SettingsPage };
  const Page = pages[r.name] && (r.name !== "asset" || S.meta.assets.some((a) => a.code === r.arg)) ? pages[r.name] : Dashboard;
  $$("#nav a").forEach((a) => a.classList.toggle("active", a.dataset.route === (r.name === "asset" ? "dashboard" : r.name)));
  const view = $("#view");
  view.innerHTML = "";
  S.page = Page(r.arg);
  S.page.mount(view);
  window.scrollTo(0, 0);
}

// ---- Shared bits ---------------------------------------------------------------------------

function coinBadge(code, size = "") {
  const a = S.meta.assets.find((x) => x.code === code);
  return `<div class="coin ${size}" style="background:${a?.color || "#888"}">${esc(a?.symbol || code[0])}</div>`;
}

function historyURL(code, range, sources, points) {
  return `/api/history/${code}?vs=${encodeURIComponent(S.settings.vs)}&range=${range}&sources=${encodeURIComponent(sources.join(","))}&points=${points}`;
}

function flashPrice(el, v, prev) {
  if (prev == null || v == null || v === prev) return;
  el.classList.remove("flash-up", "flash-down");
  void el.offsetWidth;
  el.classList.add(v > prev ? "flash-up" : "flash-down");
  setTimeout(() => el.classList.remove("flash-up", "flash-down"), 700);
}

let ticker = null;
function everySecond(fn) {
  clearInterval(ticker);
  ticker = setInterval(fn, 1000);
}

// ---- Dashboard ------------------------------------------------------------------------------

function Dashboard() {
  const sparks = new Map();
  let root;

  function card(a) {
    return `<div class="card" data-code="${a.code}" tabindex="0">
      <div class="card-head">${coinBadge(a.code)}
        <div class="card-title"><div class="name">${esc(assetName(a.code))}</div><div class="code" data-f="code">${a.code}</div></div>
        <span class="pill num" data-f="chg">—</span>
      </div>
      <div class="price num" data-f="price">—</div>
      <canvas class="spark" data-f="spark"></canvas>
      <div class="best" data-f="best"></div>
      <div class="foot"><i class="dot waiting" data-f="dot"></i><span class="grow" data-f="sources"></span><span data-f="ago"></span></div>
    </div>`;
  }

  function update(v, prev) {
    const el = root?.querySelector(`.card[data-code="${v.code}"]`);
    if (!el) return;
    const f = (n) => el.querySelector(`[data-f="${n}"]`);
    const priceEl = f("price");
    priceEl.textContent = v.price ? fmtPrice(v.price, v.vs) : v.status === "down" ? t("card.down") : t("card.waiting");
    if (prev && prev.vs === v.vs) flashPrice(priceEl, v.price, prev.price);
    const chg = f("chg");
    chg.textContent = fmtPct(v.change24h);
    chg.className = "pill num " + changeClass(v.change24h);
    f("code").textContent = v.vs !== S.settings.vs ? `${v.code} · ${t("card.vsNote", { vs: v.vs })}` : `${v.code}/${v.vs}`;
    const b = v.best || {};
    f("best").innerHTML = b.buy
      ? `<span class="tag buy">${esc(t("card.buy"))}</span><span>${esc(sourceName(b.buy.source))}</span><span class="num">${esc(fmtPrice(b.buy.effective, v.vs))}</span>
         <span class="tag sell">${esc(t("card.sell"))}</span><span>${esc(sourceName(b.sell.source))}</span><span class="num">${esc(fmtPrice(b.sell.effective, v.vs))}</span>`
      : "";
    f("dot").className = "dot " + v.status;
    f("sources").textContent = v.total ? t("card.sources", { n: v.online, m: v.total }) : "";
    f("ago").textContent = timeAgo(v.time);
    el.dataset.time = v.time || "";
    const sp = sparks.get(v.code);
    if (sp && v.price && sp.vs === v.vs) sp.chart.append("agg", v.time, v.price);
  }

  function stats() {
    if (!root) return;
    const crypto = [...S.views.values()].filter((v) => v.kind === "crypto" && v.best?.buy);
    let best = null;
    for (const v of crypto) if (!best || v.best.spreadPct > best.best.spreadPct) best = v;
    if (best) {
      $("#stArb").innerHTML = `<span class="${changeClass(best.best.spreadPct)}">${esc(fmtPct(best.best.spreadPct, 3))}</span> <span class="muted" style="font-size:1rem">${best.code}</span>`;
      $("#stArbSub").textContent = best.best.spreadPct > 0
        ? t("detail.arbHint", { buy: sourceName(best.best.buy.source), sell: sourceName(best.best.sell.source) })
        : t("detail.noArb");
    }
    const up = S.services.filter((s) => s.status === "up").length;
    $("#stSvc").textContent = `${up} / ${S.services.length || S.meta.assets.length}`;
    const online = [...S.views.values()].reduce((n, v) => n + (v.online || 0), 0);
    const total = [...S.views.values()].reduce((n, v) => n + (v.total || 0), 0);
    $("#stSvcSub").textContent = t("card.sources", { n: online, m: total });
    $("#stAlerts").textContent = S.rules.filter((r) => r.enabled).length;
    $("#stAlertsSub").textContent = S.events[0] ? `${alertMessage(S.events[0])} · ${timeAgo(S.events[0].time)}` : t("alerts.noEvents");
  }

  async function loadSparks() {
    const vs = S.settings.vs;
    await Promise.all(S.meta.assets.map(async (a) => {
      const sp = sparks.get(a.code);
      try {
        let h = await apiGet(historyURL(a.code, "24h", ["agg"], 160));
        if ((h.series.agg || []).length < 12) h = await apiGet(historyURL(a.code, "7d", ["agg"], 160));
        if (!root || S.settings.vs !== vs) return;
        sp.vs = h.vs;
        sp.chart.setSeries([{ id: "agg", label: a.code, color: a.color, points: h.series.agg || [] }]);
      } catch {}
    }));
  }

  return {
    mount(view) {
      const uniq = new Set();
      S.meta.assets.forEach((a) => a.sources.forEach((s) => uniq.add(s.id)));
      view.innerHTML = `
        <div class="page-head"><div><h1>${esc(t("nav.dashboard"))}</h1><div class="muted">${esc(t("dash.subtitle", { n: uniq.size }))}</div></div></div>
        <div class="grid stats">
          <div class="panel stat"><div class="label">${esc(t("dash.arbitrage"))}</div><div class="value num" id="stArb">—</div><div class="sub" id="stArbSub"></div></div>
          <div class="panel stat"><div class="label">${esc(t("dash.services"))}</div><div class="value num" id="stSvc">—</div><div class="sub" id="stSvcSub"></div></div>
          <div class="panel stat"><div class="label">${esc(t("dash.alerts"))}</div><div class="value num" id="stAlerts">—</div><div class="sub" id="stAlertsSub"></div></div>
        </div>
        <h2>${esc(t("section.crypto"))}</h2>
        <div class="grid cards">${S.meta.assets.filter((a) => a.kind === "crypto").map(card).join("")}</div>
        <h2>${esc(t("section.fiat"))}</h2>
        <div class="grid cards fiat">${S.meta.assets.filter((a) => a.kind === "fiat").map(card).join("")}</div>`;
      root = view;
      for (const el of $$(".card", view)) {
        const code = el.dataset.code;
        const go = () => (location.hash = "#/asset/" + code);
        el.addEventListener("click", go);
        el.addEventListener("keydown", (e) => e.key === "Enter" && go());
        sparks.set(code, { chart: makeChart(el.querySelector(".spark"), { sparkline: true }), vs: null });
      }
      S.views.forEach((v) => update(v));
      stats();
      loadSparks();
      everySecond(() => {
        $$(".card", root).forEach((el) => {
          const ago = el.querySelector('[data-f="ago"]');
          if (el.dataset.time) ago.textContent = timeAgo(+el.dataset.time);
        });
      });
    },
    destroy() {
      sparks.forEach((s) => dropChart(s.chart));
      root = null;
    },
    onViews() {
      S.views.forEach((v) => update(v));
      stats();
    },
    onView(v, prev) {
      update(v, prev);
      stats();
    },
    onStatus: stats,
    onRules: stats,
    onEvent: stats,
  };
}

// ---- Asset detail ---------------------------------------------------------------------------

function AssetPage(code) {
  const asset = S.meta.assets.find((a) => a.code === code);
  const isCrypto = asset.kind === "crypto";
  let range = S.settings.range || "24h";
  let selected = new Set((S.settings.watch && S.settings.watch[code]) || ["agg"]);
  let percent = false;
  let chart = null;
  let data = {}; // id -> points
  let histVs = null;
  let convMode = "buy";
  let root;
  let loadSeq = 0;

  const ids = () => ["agg", ...asset.sources.map((s) => s.id)];

  function seriesList() {
    return ids().filter((id) => selected.has(id)).map((id) => ({
      id, label: id === "agg" ? t("detail.aggregate") : sourceName(id), color: sourceColor(id, code),
      points: data[id] || [], width: id === "agg" ? 2.4 : 1.8,
    }));
  }

  async function load() {
    const seq = ++loadSeq;
    $("#overlay").textContent = t("detail.loading");
    try {
      const h = await apiGet(historyURL(code, range, [...selected], 700));
      if (seq !== loadSeq || !root) return;
      data = h.series;
      histVs = h.vs;
      chart.setOptions({ from: Date.now() - rangeMs(range) });
      chart.setSeries(seriesList());
      const empty = Object.values(data).every((p) => !p.length);
      $("#overlay").textContent = empty ? t("detail.noHistory") : "";
    } catch (e) {
      if (seq === loadSeq) $("#overlay").textContent = e.message;
    }
  }

  function chips() {
    $("#chips").innerHTML = ids().map((id) => {
      const v = S.views.get(code);
      const sv = v?.sources.find((s) => s.id === id);
      const dim = id !== "agg" && sv && sv.state !== "ok" && sv.state !== "excluded";
      return `<span class="chip${selected.has(id) ? " on" : ""}${dim ? " off" : ""}" data-id="${esc(id)}" style="--c:${sourceColor(id, code)}"><i></i>${esc(id === "agg" ? t("detail.aggregate") : sourceName(id))}</span>`;
    }).join("");
  }

  function toggle(id) {
    if (selected.has(id)) {
      if (selected.size === 1) return;
      selected.delete(id);
    } else {
      selected.add(id);
    }
    const watch = { ...(S.settings.watch || {}), [code]: [...selected] };
    saveSettings({ watch });
    chips();
    load();
  }

  function head(v) {
    const p = $("#dPrice");
    const prevText = p.dataset.v ? +p.dataset.v : null;
    p.textContent = v?.price ? fmtPrice(v.price, v.vs) : v?.status === "down" ? t("card.down") : t("card.waiting");
    if (v?.price) {
      flashPrice(p, v.price, prevText);
      p.dataset.v = v.price;
    }
    $("#dPair").textContent = `${code}/${v?.vs || S.settings.vs}`;
    const c = $("#dChg");
    c.textContent = v ? `${fmtPct(v.change24h)} · 24h` : "";
    c.className = "num " + changeClass(v?.change24h);
  }

  function best(v) {
    const b = v.best || {};
    const vs = v.vs;
    const fee = (id) => v.sources.find((s) => s.id === id)?.fee || 0;
    const row = (label, bp, hint) => bp
      ? `<div class="k">${esc(label)}</div><div class="v num">${esc(fmtPrice(bp.effective, vs))}<small>${esc(sourceName(bp.source))}${isCrypto ? " · " + esc(fmtPrice(bp.price, vs)) + " + " + esc(fmtPct(fee(bp.source) * 100, 2).replace("+", "")) : ""}${hint ? " " + esc(hint) : ""}</small></div>`
      : "";
    $("#bestKV").innerHTML = row(t("detail.bestBuy"), b.buy) + row(t("detail.bestSell"), b.sell) +
      (b.worstBuy ? `<div class="k">${esc(t("detail.saving"))}</div><div class="v num">${esc(fmtPct(b.savingPct, 3).replace("+", ""))}<small>${esc(t("detail.savingHint"))} (${esc(sourceName(b.worstBuy.source))})</small></div>` : "");
    const arb = $("#arb");
    if (b.buy && b.sell) {
      const good = b.spreadPct > 0 && b.buy.source !== b.sell.source;
      arb.className = "arb" + (good ? " good" : "");
      arb.innerHTML = `<b>${esc(t("detail.arbitrage"))}: <span class="num ${changeClass(b.spreadPct)}">${esc(fmtPct(b.spreadPct, 3))}</span></b><br>` +
        esc(good ? t("detail.arbHint", { buy: sourceName(b.buy.source), sell: sourceName(b.sell.source) }) : t("detail.noArb"));
    } else {
      arb.className = "arb";
      arb.textContent = t("card.waiting");
    }
  }

  function converter(v) {
    const amount = parseFloat($("#convAmount").value.replace(",", ".")) || 0;
    const b = v.best || {};
    const bestP = convMode === "buy" ? b.buy : b.sell;
    const worstP = convMode === "buy" ? b.worstBuy : b.worstSell;
    if (!bestP || !worstP) {
      $("#convKV").innerHTML = "";
      return;
    }
    const bestV = amount * bestP.effective, worstV = amount * worstP.effective;
    $("#convKV").innerHTML =
      `<div class="k">${esc(convMode === "buy" ? t("detail.cost") : t("detail.receive"))} · ${esc(t("detail.bestAt", { src: sourceName(bestP.source) }))}</div><div class="v num up">${esc(fmtPrice(bestV, v.vs))}</div>` +
      `<div class="k">${esc(t("detail.worstAt", { src: sourceName(worstP.source) }))}</div><div class="v num">${esc(fmtPrice(worstV, v.vs))}</div>` +
      `<div class="k">${esc(t("detail.diff"))}</div><div class="v num">${esc(fmtPrice(Math.abs(worstV - bestV), v.vs))}</div>`;
  }

  function table(v) {
    const cols = isCrypto
      ? ["source", "pair", "bid", "ask", "spread", "fee", "effBuy", "effSell", "change", "updated", "status"]
      : ["source", "pair", "rate", "spread", "updated", "status"];
    $("#dHead").innerHTML = "<tr>" + cols.map((c) => `<th>${esc(c === "spread" && !isCrypto ? "Δ" : t("col." + c))}</th>`).join("") + "</tr>";
    $("#dRows").innerHTML = v.sources.map((s) => {
      const st = `<span class="state ${s.state}" title="${esc(s.error || "")}">${esc(t("state." + s.state))}</span>`;
      const name = `<div class="src-name"><span class="sw" style="background:${sourceColor(s.id, code)}"></span><span>${esc(s.name)}<br><small>${esc(t("type." + s.type))}</small></span></div>`;
      const ok = s.state === "ok" || s.state === "excluded";
      const dev = ok && v.price ? (s.rate / v.price - 1) * 100 : null;
      const cells = {
        source: name,
        pair: `<span class="muted">${esc(s.pair || "")}</span>`,
        bid: ok ? fmtNum(s.bid) : "—",
        ask: ok ? fmtNum(s.ask) : "—",
        rate: ok ? fmtPrice(s.rate, v.vs) : "—",
        spread: isCrypto ? (ok ? fmtPct(s.spreadPct, 4).replace("+", "") : "—") : (dev == null ? "—" : `<span class="${changeClass(dev)}">${fmtPct(dev, 3)}</span>`),
        fee: fmtPct(s.fee * 100, 2).replace("+", "") + (s.feeCustom ? "*" : ""),
        effBuy: ok ? fmtNum(s.effBuy) : "—",
        effSell: ok ? fmtNum(s.effSell) : "—",
        change: `<span class="${changeClass(s.change24h)}">${fmtPct(s.change24h)}</span>`,
        updated: `<span class="muted">${esc(timeAgo(isCrypto ? s.fetched : s.time || s.fetched))}</span>`,
        status: st,
      };
      const cls = (c) => (c === "effBuy" && s.bestBuy ? "best-buy" : c === "effSell" && s.bestSell ? "best-sell" : c === "rate" && s.bestBuy ? "best-buy" : "");
      return `<tr class="${ok ? "" : "dim"}" data-id="${esc(s.id)}" style="cursor:pointer">` +
        cols.map((c) => `<td class="num ${cls(c)}">${cells[c]}</td>`).join("") + "</tr>";
    }).join("");
  }

  function update(v) {
    if (!root || !v) return;
    head(v);
    best(v);
    converter(v);
    table(v);
    if (v.vs === histVs && rangeMs(range) <= 7 * 86400e3) {
      for (const id of selected) {
        const val = id === "agg" ? v.price : v.sources.find((s) => s.id === id && s.state !== "error")?.rate;
        if (val) chart.append(id, v.time, val);
      }
    }
  }

  return {
    mount(view) {
      root = view;
      view.innerHTML = `
        <a class="btn small" href="#/">${document.documentElement.dir === "rtl" ? "→" : "←"} ${esc(t("detail.back"))}</a>
        <div class="detail-head" style="margin-top:.9rem">
          ${coinBadge(code, "lg")}
          <div class="titles">
            <h1>${esc(assetName(code))} <span class="muted" id="dPair" style="font-weight:500;font-size:1rem"></span></h1>
            <div class="big num" id="dPrice">—</div><div id="dChg" class="num"></div>
          </div>
          <button class="btn primary" id="dAlert" type="button">＋ ${esc(t("detail.createAlert"))}</button>
        </div>
        <div class="toolbar">
          <div class="seg" id="ranges">${RANGES.map((r) => `<button type="button" data-r="${r}" class="${r === range ? "on" : ""}">${esc(t("range." + r))}</button>`).join("")}</div>
          <div class="seg" id="styleSeg"><button type="button" data-s="area">${esc(t("detail.area"))}</button><button type="button" data-s="line">${esc(t("detail.line"))}</button></div>
          <div class="seg"><button type="button" id="pctBtn">${esc(t("detail.percent"))}</button></div>
        </div>
        <div class="chips" id="chips"></div>
        <div class="panel">
          <div class="chart-box"><canvas id="dChart"></canvas><div class="overlay" id="overlay"></div></div>
          <div class="muted" style="font-size:.8rem;margin-top:.4rem">${esc(t("detail.chartHint"))}</div>
        </div>
        <div class="grid two">
          <div class="panel"><h3>${esc(t("detail.best"))}</h3><div class="kv" id="bestKV"></div><div class="arb" id="arb"></div></div>
          <div class="panel"><h3>${esc(t("detail.converter"))}</h3>
            <div class="conv">
              <div class="seg" id="convSeg"><button type="button" data-m="buy" class="on">${esc(t("detail.buy"))}</button><button type="button" data-m="sell">${esc(t("detail.sell"))}</button></div>
              <input type="text" inputmode="decimal" id="convAmount" value="1" aria-label="${esc(t("detail.amount"))}"><span class="muted">${code}</span>
            </div>
            <div class="kv" id="convKV"></div>
          </div>
        </div>
        <h2>${esc(t("detail.sources"))}</h2>
        <div class="table-wrap"><table><thead id="dHead"></thead><tbody id="dRows"></tbody></table></div>`;

      chart = makeChart($("#dChart"), { area: S.settings.chartStyle !== "line" });
      $$("#styleSeg button").forEach((b) => b.classList.toggle("on", b.dataset.s === (S.settings.chartStyle || "area")));
      chips();
      update(S.views.get(code));
      load();

      $("#ranges").addEventListener("click", (e) => {
        const r = e.target.closest("button")?.dataset.r;
        if (!r) return;
        range = r;
        $$("#ranges button").forEach((b) => b.classList.toggle("on", b.dataset.r === r));
        saveSettings({ range: r });
        load();
      });
      $("#styleSeg").addEventListener("click", (e) => {
        const s = e.target.closest("button")?.dataset.s;
        if (!s) return;
        $$("#styleSeg button").forEach((b) => b.classList.toggle("on", b.dataset.s === s));
        chart.setOptions({ area: s === "area" });
        saveSettings({ chartStyle: s });
      });
      $("#pctBtn").addEventListener("click", (e) => {
        percent = !percent;
        e.target.classList.toggle("on", percent);
        chart.setOptions({ percent });
      });
      $("#chips").addEventListener("click", (e) => {
        const id = e.target.closest(".chip")?.dataset.id;
        if (id) toggle(id);
      });
      $("#dRows").addEventListener("click", (e) => {
        const id = e.target.closest("tr")?.dataset.id;
        if (id) toggle(id);
      });
      $("#convSeg").addEventListener("click", (e) => {
        const m = e.target.closest("button")?.dataset.m;
        if (!m) return;
        convMode = m;
        $$("#convSeg button").forEach((b) => b.classList.toggle("on", b.dataset.m === m));
        converter(S.views.get(code));
      });
      $("#convAmount").addEventListener("input", () => converter(S.views.get(code)));
      $("#dAlert").addEventListener("click", () => openAlertDialog({ asset: code, vs: S.views.get(code)?.vs || S.settings.vs }));
      everySecond(() => {
        const v = S.views.get(code);
        if (v) table(v);
      });
    },
    destroy() {
      dropChart(chart);
      root = null;
    },
    onViews() {
      update(S.views.get(code));
    },
    onView(v) {
      if (v.code === code) update(v);
    },
  };
}

function rangeMs(r) {
  const n = parseInt(r, 10);
  const unit = r.slice(-1);
  return n * ({ h: 3600e3, d: 86400e3, y: 365 * 86400e3 }[unit] || 3600e3);
}

// ---- Markets ------------------------------------------------------------------------------

function MarketsPage() {
  let root, chart;
  let range = "30d";
  let picked = new Set(["BTC", "ETH", "TON"]);
  let loadSeq = 0;

  function matrix(kind) {
    const assets = S.meta.assets.filter((a) => a.kind === kind);
    const cols = [];
    assets.forEach((a) => a.sources.forEach((s) => !cols.includes(s.id) && cols.push(s.id)));
    const head = `<tr><th></th>${cols.map((c) => `<th>${esc(sourceName(c))}</th>`).join("")}</tr>`;
    const rows = assets.map((a) => {
      const v = S.views.get(a.code);
      const cells = cols.map((c) => {
        const s = v?.sources.find((x) => x.id === c);
        const ok = s && (s.state === "ok" || s.state === "excluded");
        const cls = !ok ? "muted" : s.bestBuy && s.bestSell ? "best-both" : s.bestBuy ? "best-buy" : s.bestSell ? "best-sell" : "";
        return `<td class="num ${cls}" data-cell="${a.code}:${c}">${ok ? esc(fmtNum(s.rate)) : "—"}</td>`;
      }).join("");
      return `<tr><td><div class="src-name">${coinBadge(a.code, "sm")}<span>${a.code}<br><small>${esc(v?.vs || "")}</small></span></div></td>${cells}</tr>`;
    }).join("");
    return { head, rows };
  }

  function renderMatrices() {
    for (const kind of ["crypto", "fiat"]) {
      const m = matrix(kind);
      $(`#m-${kind} thead`).innerHTML = m.head;
      const body = $(`#m-${kind} tbody`);
      const old = new Map($$("td[data-cell]", body).map((td) => [td.dataset.cell, td.textContent]));
      body.innerHTML = m.rows;
      $$("td[data-cell]", body).forEach((td) => {
        const prev = old.get(td.dataset.cell);
        if (prev !== undefined && prev !== td.textContent) td.classList.add("flash");
      });
    }
  }

  async function loadCompare() {
    const seq = ++loadSeq;
    const list = [...picked];
    const res = await Promise.all(list.map((c) => apiGet(historyURL(c, range, ["agg"], 400)).catch(() => null)));
    if (seq !== loadSeq || !root) return;
    chart.setOptions({ from: Date.now() - rangeMs(range) });
    chart.setSeries(list.map((c, i) => {
      const a = S.meta.assets.find((x) => x.code === c);
      return { id: c, label: c, color: a.color, points: res[i]?.series.agg || [] };
    }));
  }

  function compareChips() {
    $("#cmpChips").innerHTML = S.meta.assets.map((a) =>
      `<span class="chip${picked.has(a.code) ? " on" : ""}" data-id="${a.code}" style="--c:${a.color}"><i></i>${a.code}</span>`).join("");
  }

  return {
    mount(view) {
      root = view;
      view.innerHTML = `
        <div class="page-head"><div><h1>${esc(t("markets.title"))}</h1><div class="muted">${esc(t("markets.hint"))}</div></div></div>
        <div class="table-wrap" id="m-crypto"><table><thead></thead><tbody></tbody></table></div>
        <h2>${esc(t("markets.fiat"))}</h2>
        <div class="table-wrap" id="m-fiat"><table><thead></thead><tbody></tbody></table></div>
        <h2>${esc(t("markets.compare"))} <span class="muted" style="font-weight:400;font-size:.85rem">${esc(t("markets.compareHint"))}</span></h2>
        <div class="toolbar" style="margin-top:0"><div class="seg" id="cmpRanges">${RANGES.map((r) => `<button type="button" data-r="${r}" class="${r === range ? "on" : ""}">${esc(t("range." + r))}</button>`).join("")}</div></div>
        <div class="chips" id="cmpChips"></div>
        <div class="panel"><div class="chart-box"><canvas id="cmpChart"></canvas></div></div>`;
      chart = makeChart($("#cmpChart"), { percent: true, area: false });
      renderMatrices();
      compareChips();
      loadCompare();
      $("#cmpRanges").addEventListener("click", (e) => {
        const r = e.target.closest("button")?.dataset.r;
        if (!r) return;
        range = r;
        $$("#cmpRanges button").forEach((b) => b.classList.toggle("on", b.dataset.r === r));
        loadCompare();
      });
      $("#cmpChips").addEventListener("click", (e) => {
        const id = e.target.closest(".chip")?.dataset.id;
        if (!id) return;
        if (picked.has(id)) {
          if (picked.size > 1) picked.delete(id);
        } else picked.add(id);
        compareChips();
        loadCompare();
      });
      everySecond(() => {});
    },
    destroy() {
      dropChart(chart);
      root = null;
    },
    onViews: renderMatrices,
    onView: renderMatrices,
  };
}

// ---- Alerts --------------------------------------------------------------------------------

function ruleText(r) {
  const label = t("cond." + r.condition);
  let cond;
  if (r.condition.endsWith("_pct")) {
    // Put the number where the label has its "%" sign, whatever the word order.
    const num = fmtNum(r.value, { maximumFractionDigits: 4 });
    cond = /±?%/.test(label) ? label.replace(/(±?)%/, (_, pm) => pm + num + "%") : `${label} ${num}%`;
  } else {
    cond = `${label} ${fmtPrice(r.value, r.vs)}`;
  }
  const win = r.condition === "above" || r.condition === "below" ? "" : " · " + t("win." + r.window);
  return `${r.asset}/${r.vs} — ${cond}${win}`;
}

function AlertsPage() {
  let root;

  function rules() {
    const list = $("#rules");
    if (!S.rules.length) {
      list.innerHTML = `<div class="panel empty">${esc(t("alerts.empty"))}</div>`;
      return;
    }
    list.innerHTML = S.rules.map((r) => {
      let progress = "";
      if (r.lastChange != null && r.condition !== "above" && r.condition !== "below") {
        let p = 0;
        const ch = r.lastChange;
        if (r.condition.endsWith("_pct")) {
          const dir = r.condition === "rise_pct" ? ch : r.condition === "fall_pct" ? -ch : Math.abs(ch);
          p = dir / r.value;
        } else {
          const ref = r.lastPrice / (1 + ch / 100);
          const abs = r.lastPrice - ref;
          p = (r.condition === "rise_abs" ? abs : -abs) / r.value;
        }
        progress = `<div class="progress"><span style="width:${Math.max(0, Math.min(100, p * 100)).toFixed(1)}%"></span></div>`;
      }
      const src = r.source && r.source !== "agg" ? " · " + sourceName(r.source) : "";
      const meta = [
        `<span class="${r.enabled ? "up" : ""}">${esc(r.enabled ? t("alerts.active") : t("alerts.paused"))}</span>`,
        r.triggers ? esc(t("alerts.fired", { n: r.triggers })) : "",
        r.lastPrice ? esc(t("alerts.now", { v: fmtPrice(r.lastPrice, r.vs) })) + (r.lastChange != null && r.condition.endsWith("pct") ? ` (${esc(fmtPct(r.lastChange, 3))})` : "") : "",
        r.window === "since" && r.baseline && !(r.condition === "above" || r.condition === "below") ? esc(t("alerts.baseline", { v: fmtPrice(r.baseline, r.vs) })) : "",
      ].filter(Boolean).join("<span>·</span>");
      return `<div class="panel rule${r.enabled ? "" : " off"}" data-id="${esc(r.id)}">
        ${coinBadge(r.asset)}
        <div><div class="what">${esc(ruleText(r))}${esc(src)}</div>${r.note ? `<div class="muted">${esc(r.note)}</div>` : ""}<div class="meta num">${meta}</div>${progress}</div>
        <div class="actions">
          <button class="btn small" data-act="toggle" type="button">${esc(r.enabled ? t("alerts.disable") : t("alerts.enable"))}</button>
          <button class="btn small icon" data-act="edit" type="button" title="${esc(t("alerts.edit"))}"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z"/></svg></button>
          <button class="btn small icon danger" data-act="delete" type="button" title="${esc(t("alerts.delete"))}"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6"/></svg></button>
        </div></div>`;
    }).join("");
  }

  function events() {
    const el = $("#events");
    if (!S.events.length) {
      el.innerHTML = `<div class="empty">${esc(t("alerts.noEvents"))}</div>`;
      return;
    }
    el.innerHTML = S.events.slice(0, 100).map((ev) => `<div class="event">${coinBadge(ev.asset)}<div>${esc(alertMessage(ev))}${ev.source && ev.source !== "agg" ? ` <span class="muted">· ${esc(sourceName(ev.source))}</span>` : ""}<div class="muted" style="font-size:.8rem">${esc(fmtDateTime(ev.time))}</div></div><div class="num">${esc(fmtPrice(ev.price, ev.vs))}</div></div>`).join("");
  }

  return {
    mount(view) {
      root = view;
      view.innerHTML = `
        <div class="page-head"><h1>${esc(t("alerts.title"))}</h1><button class="btn primary" id="newRule" type="button">＋ ${esc(t("alerts.new"))}</button></div>
        <div class="rules" id="rules"></div>
        <h2>${esc(t("alerts.history"))} <span class="spacer"></span><button class="btn small" id="clearEvents" type="button">${esc(t("alerts.clear"))}</button></h2>
        <div class="panel events" id="events"></div>`;
      rules();
      events();
      apiGet("/api/alerts/events").then((evs) => {
        S.events = evs;
        if (root) events();
      }).catch(() => {});
      $("#newRule").addEventListener("click", () => openAlertDialog({}));
      $("#clearEvents").addEventListener("click", async () => {
        await apiSend("DELETE", "/api/alerts/events");
        S.events = [];
        events();
      });
      $("#rules").addEventListener("click", async (e) => {
        const btn = e.target.closest("button");
        const id = e.target.closest(".rule")?.dataset.id;
        const r = S.rules.find((x) => x.id === id);
        if (!btn || !r) return;
        try {
          if (btn.dataset.act === "toggle") {
            const saved = await apiSend("PUT", "/api/alerts/" + id, { ...r, enabled: !r.enabled });
            S.rules = S.rules.map((x) => (x.id === id ? saved : x));
          } else if (btn.dataset.act === "edit") {
            openAlertDialog(r);
            return;
          } else if (btn.dataset.act === "delete") {
            if (!confirm(t("alerts.deleteConfirm"))) return;
            await apiSend("DELETE", "/api/alerts/" + id);
            S.rules = S.rules.filter((x) => x.id !== id);
          }
          updateBadge();
          rules();
        } catch (err) {
          toast(t("toast.error", { e: err.message }), "error");
        }
      });
    },
    destroy() {
      root = null;
    },
    onRules: () => root && rules(),
    onEvent: () => root && (events(), rules()),
    onView: () => {},
  };
}

function openAlertDialog(rule) {
  const dlg = $("#alertDialog");
  const form = $("#alertForm");
  const editing = !!rule.id;
  const r = Object.assign({ asset: "BTC", vs: S.settings.vs, source: "agg", condition: "rise_pct", value: 5, window: "since", repeat: false, note: "", enabled: true }, rule);
  if (r.asset === r.vs) r.vs = r.asset === "USD" ? "EUR" : "USD";

  form.innerHTML = `
    <h2>${esc(editing ? t("alerts.edit") : t("alerts.new"))}</h2>
    <div class="form-grid">
      <div class="field"><label>${esc(t("alerts.asset"))}</label><select name="asset">${S.meta.assets.map((a) => `<option value="${a.code}">${a.code} — ${esc(assetName(a.code))}</option>`).join("")}</select></div>
      <div class="field"><label>${esc(t("alerts.vs"))}</label><select name="vs"></select></div>
      <div class="field full"><label>${esc(t("alerts.source"))}</label><select name="source"></select></div>
      <div class="field"><label>${esc(t("alerts.condition"))}</label><select name="condition">${CONDITIONS.map((c) => `<option value="${c}">${esc(t("cond." + c))}</option>`).join("")}</select></div>
      <div class="field"><label>${esc(t("alerts.value"))}</label><div class="input-unit"><input type="text" inputmode="decimal" name="value" required><span id="valUnit">%</span></div></div>
      <div class="field full" id="winField"><label>${esc(t("alerts.window"))}</label><select name="window">${WINDOWS.map((w) => `<option value="${w}">${esc(t("win." + w))}</option>`).join("")}</select></div>
      <div class="field full"><label>${esc(t("alerts.note"))}</label><input type="text" name="note" maxlength="200" placeholder="${esc(t("alerts.notePh"))}"></div>
    </div>
    <label class="check"><input type="checkbox" name="repeat"> ${esc(t("alerts.repeat"))}</label>
    <div class="hint num" id="alertHint"></div>
    <div class="dialog-actions">
      <button class="btn" value="cancel" type="button" id="dlgCancel">${esc(t("alerts.cancel"))}</button>
      <button class="btn primary" type="submit">${esc(t("alerts.save"))}</button>
    </div>`;

  const f = form.elements;
  const fillVs = () => {
    const a = f.asset.value;
    f.vs.innerHTML = S.meta.fiat.filter((c) => c !== a).map((c) => `<option value="${c}">${c}</option>`).join("");
    if (f.vs.dataset.want && f.vs.dataset.want !== a) f.vs.value = f.vs.dataset.want;
    const asset = S.meta.assets.find((x) => x.code === a);
    f.source.innerHTML = `<option value="agg">${esc(t("detail.aggregate"))}</option>` +
      asset.sources.map((s) => `<option value="${esc(s.id)}">${esc(s.name)}</option>`).join("");
  };
  const currentPrice = () => {
    const v = S.views.get(f.asset.value);
    if (!v || v.vs !== f.vs.value) return null;
    if (f.source.value === "agg") return v.price;
    return v.sources.find((s) => s.id === f.source.value)?.rate || null;
  };
  const refresh = () => {
    const cond = f.condition.value;
    const level = cond === "above" || cond === "below";
    $("#winField").classList.toggle("hidden", level);
    $("#valUnit").textContent = cond.endsWith("_pct") ? "%" : f.vs.value;
    const price = currentPrice();
    const value = parseFloat(String(f.value.value).replace(",", "."));
    let hint = price ? t("alerts.current", { v: fmtPrice(price, f.vs.value) }) : "";
    if (price && value > 0 && (level || f.window.value === "since")) {
      const fire = { rise_pct: price * (1 + value / 100), fall_pct: price * (1 - value / 100), rise_abs: price + value, fall_abs: price - value, above: value, below: value }[cond];
      if (fire) hint += " · " + t("alerts.fireAt", { v: fmtPrice(fire, f.vs.value) });
      if (cond === "move_pct") hint += " · " + t("alerts.fireAt", { v: fmtPrice(price * (1 - value / 100), f.vs.value) + " / " + fmtPrice(price * (1 + value / 100), f.vs.value) });
    }
    $("#alertHint").textContent = hint;
    $("#alertHint").classList.toggle("hidden", !hint);
  };

  f.asset.value = r.asset;
  f.vs.dataset.want = r.vs;
  fillVs();
  f.vs.value = r.vs;
  f.source.value = r.source || "agg";
  f.condition.value = r.condition;
  f.value.value = r.value;
  f.window.value = r.window;
  f.repeat.checked = !!r.repeat;
  f.note.value = r.note || "";
  refresh();

  f.asset.addEventListener("change", () => {
    f.vs.dataset.want = f.vs.value;
    fillVs();
    refresh();
  });
  f.condition.addEventListener("change", () => {
    const price = currentPrice();
    if (price && (f.condition.value === "above" || f.condition.value === "below")) {
      const k = f.condition.value === "above" ? 1.05 : 0.95;
      f.value.value = +(price * k).toPrecision(5);
    } else if (f.condition.value.endsWith("_pct") && parseFloat(f.value.value) > 100) {
      f.value.value = 5;
    }
    refresh();
  });
  ["vs", "source", "value", "window"].forEach((n) => f[n].addEventListener("input", refresh));
  $("#dlgCancel").onclick = () => dlg.close();
  form.onsubmit = async (e) => {
    e.preventDefault();
    const body = {
      ...r, asset: f.asset.value, vs: f.vs.value, source: f.source.value, condition: f.condition.value,
      value: parseFloat(String(f.value.value).replace(",", ".")), window: f.window.value, repeat: f.repeat.checked, note: f.note.value.trim(),
    };
    try {
      const saved = editing ? await apiSend("PUT", "/api/alerts/" + r.id, body) : await apiSend("POST", "/api/alerts", body);
      S.rules = editing ? S.rules.map((x) => (x.id === saved.id ? saved : x)) : [...S.rules, saved];
      updateBadge();
      dlg.close();
      toast(t("toast.saved"));
      S.page?.onRules?.();
      if (S.settings.browserNotify && "Notification" in window && Notification.permission === "default") Notification.requestPermission();
    } catch (err) {
      toast(t("toast.error", { e: err.message }), "error");
    }
  };
  dlg.showModal();
}

// ---- Settings -----------------------------------------------------------------------------

function SettingsPage() {
  let root;

  function services() {
    const el = $("#svcRows");
    if (!el) return;
    el.innerHTML = S.services.map((s) => `<tr>
      <td><div class="src-name">${coinBadge(s.code, "sm")}<span>${s.code}</span></div></td>
      <td class="muted ltr">${esc(s.url)}</td>
      <td><span class="state ${s.status === "up" ? "ok" : s.status === "down" ? "error" : "waiting"}" title="${esc(s.error || "")}">${esc(t("svc." + s.status))}</span></td>
      <td class="num">${s.online}/${s.sources}</td>
      <td class="muted">${esc(timeAgo(s.lastSeen))}</td></tr>`).join("");
  }

  function sourcesTable() {
    const seen = new Map();
    for (const a of S.meta.assets) for (const s of a.sources) if (!seen.has(s.id)) seen.set(s.id, s);
    const disabled = new Set(S.settings.disabled || []);
    $("#srcRows").innerHTML = [...seen.values()].map((s) => {
      const custom = S.settings.fees && S.settings.fees[s.id] != null;
      const fee = custom ? S.settings.fees[s.id] : s.fee;
      return `<tr data-id="${esc(s.id)}">
        <td><div class="src-name"><span class="sw" style="background:${sourceColor(s.id)}"></span><span>${esc(s.name)}<br><small>${esc(t("type." + s.type))}</small></span></div></td>
        <td><input type="checkbox" data-k="use" ${disabled.has(s.id) ? "" : "checked"}></td>
        <td>${s.type === "exchange" && s.fee > 0
          ? `<input class="fee-input num" type="text" inputmode="decimal" data-k="fee" value="${(fee * 100).toFixed(3).replace(/\.?0+$/, "")}"> % ${custom ? `<button class="btn small" data-k="reset" type="button">${esc(t("settings.reset"))}</button>` : ""}`
          : '<span class="muted">—</span>'}</td>
        <td><a href="${esc(s.url)}" target="_blank" rel="noopener" class="muted ltr">${esc(s.url.replace(/^https?:\/\//, ""))}</a></td></tr>`;
    }).join("");
  }

  return {
    mount(view) {
      root = view;
      const st = S.settings;
      view.innerHTML = `
        <div class="page-head"><h1>${esc(t("settings.title"))}</h1></div>
        <div class="grid settings">
          <div class="panel">
            <h3>${esc(t("settings.appearance"))}</h3>
            <div class="field"><label>${esc(t("settings.language"))}</label><select id="sLang">${LANGS.map((l) => `<option value="${l.code}">${esc(l.name)}</option>`).join("")}</select></div>
            <div class="field"><div class="label">${esc(t("settings.theme"))}</div><div class="themes" id="sThemes">${THEMES.map((th) => {
              const [a, b, c] = THEME_PREVIEW[th];
              return `<div class="theme-opt${st.theme === th ? " on" : ""}" data-theme="${th}"><div class="pv"><b style="background:${a}"></b><b style="background:${b}"></b><b style="background:${c};flex:.4"></b></div>${esc(t("theme." + th))}</div>`;
            }).join("")}</div></div>
            <div class="field"><div class="label">${esc(t("settings.scale"))}</div>
              <label class="check" style="margin:0"><input type="checkbox" id="sScaleAuto" ${st.scale ? "" : "checked"}> ${esc(t("settings.scaleAuto"))}</label>
              <div class="range-row"><input type="range" id="sScale" min="0.6" max="2" step="0.05" value="${st.scale || 1}" ${st.scale ? "" : "disabled"}><span class="num" id="sScaleVal">${Math.round((st.scale || 1) * 100)}%</span></div>
            </div>
            <div class="field"><label>${esc(t("settings.chartStyle"))}</label><div class="seg" id="sStyle"><button type="button" data-s="area" class="${st.chartStyle !== "line" ? "on" : ""}">${esc(t("detail.area"))}</button><button type="button" data-s="line" class="${st.chartStyle === "line" ? "on" : ""}">${esc(t("detail.line"))}</button></div></div>
            <div class="field"><label>${esc(t("settings.currency"))}</label><select id="sVs">${S.meta.fiat.map((c) => `<option value="${c}">${c} — ${esc(assetName(c))}</option>`).join("")}</select></div>
          </div>
          <div class="panel">
            <h3>${esc(t("settings.notifications"))}</h3>
            <div class="switch"><div>${esc(t("settings.osNotify"))}<small>${esc(t("settings.osNotifyHint"))}</small></div><label class="toggle"><input type="checkbox" id="sOS" ${st.osNotify ? "checked" : ""}><span></span></label></div>
            <div class="switch"><div>${esc(t("settings.browserNotify"))}<small id="sPerm"></small></div><label class="toggle"><input type="checkbox" id="sBrowser" ${st.browserNotify ? "checked" : ""}><span></span></label></div>
            <div class="switch"><div>${esc(t("settings.sound"))}</div><label class="toggle"><input type="checkbox" id="sSound" ${st.sound ? "checked" : ""}><span></span></label></div>
            <button class="btn" id="sTest" type="button" style="margin-top:.6rem">🔔 ${esc(t("settings.test"))}</button>
            <h3 style="margin-top:1.6rem">${esc(t("settings.about"))}</h3>
            <p class="muted" style="margin-top:0">Cryptomon ${esc(S.meta.version)} — ${esc(t("settings.aboutText"))}</p>
            <button class="btn danger" id="sQuit" type="button">⏻ ${esc(t("settings.quit"))}</button>
          </div>
        </div>
        <h2>${esc(t("settings.sources"))}</h2>
        <p class="muted" style="margin-top:-.4rem">${esc(t("settings.sourcesHint"))}</p>
        <div class="table-wrap"><table><thead><tr><th>${esc(t("col.source"))}</th><th>${esc(t("settings.include"))}</th><th>${esc(t("col.fee"))}</th><th></th></tr></thead><tbody id="srcRows"></tbody></table></div>
        <h2>${esc(t("settings.services"))}</h2>
        <div class="table-wrap"><table><thead><tr><th></th><th>URL</th><th>${esc(t("col.status"))}</th><th>${esc(t("detail.sources"))}</th><th>${esc(t("col.updated"))}</th></tr></thead><tbody id="svcRows"></tbody></table></div>`;

      $("#sLang").value = getLang();
      $("#sVs").value = st.vs;
      const perm = () => {
        const el = $("#sPerm");
        if (!("Notification" in window)) return;
        el.textContent = Notification.permission === "denied" ? t("settings.permDenied") : "";
      };
      perm();
      sourcesTable();
      services();

      $("#sLang").addEventListener("change", (e) => {
        saveSettings({ lang: e.target.value });
        applyLang();
        render();
      });
      $("#sThemes").addEventListener("click", (e) => {
        const th = e.target.closest(".theme-opt")?.dataset.theme;
        if (!th) return;
        saveSettings({ theme: th });
        applyTheme();
        $$(".theme-opt").forEach((o) => o.classList.toggle("on", o.dataset.theme === th));
      });
      const scale = $("#sScale");
      $("#sScaleAuto").addEventListener("change", (e) => {
        scale.disabled = e.target.checked;
        saveSettings({ scale: e.target.checked ? 0 : +scale.value });
        applyScale();
      });
      scale.addEventListener("input", () => {
        $("#sScaleVal").textContent = Math.round(scale.value * 100) + "%";
      });
      scale.addEventListener("change", () => {
        saveSettings({ scale: +scale.value });
        applyScale();
      });
      $("#sStyle").addEventListener("click", (e) => {
        const s = e.target.closest("button")?.dataset.s;
        if (!s) return;
        $$("#sStyle button").forEach((b) => b.classList.toggle("on", b.dataset.s === s));
        saveSettings({ chartStyle: s });
      });
      $("#sVs").addEventListener("change", (e) => {
        $("#vsSelect").value = e.target.value;
        saveSettings({ vs: e.target.value }, { reconnect: false });
        S.views.clear();
        connect();
      });
      $("#sOS").addEventListener("change", (e) => saveSettings({ osNotify: e.target.checked }));
      $("#sSound").addEventListener("change", (e) => {
        saveSettings({ sound: e.target.checked });
        if (e.target.checked) beep();
      });
      $("#sBrowser").addEventListener("change", async (e) => {
        saveSettings({ browserNotify: e.target.checked });
        if (e.target.checked && "Notification" in window && Notification.permission === "default") await Notification.requestPermission();
        perm();
      });
      $("#sTest").addEventListener("click", async () => {
        try {
          if (S.settings.osNotify) await apiSend("POST", "/api/notify/test");
          if (S.settings.browserNotify && "Notification" in window && Notification.permission === "granted") new Notification("Cryptomon", { body: t("settings.testSent") });
          if (S.settings.sound) beep();
          toast(t("settings.testSent"));
        } catch (err) {
          toast(t("toast.error", { e: err.message }), "error");
        }
      });
      $("#sQuit").addEventListener("click", async () => {
        if (!confirm(t("settings.quitConfirm"))) return;
        try {
          await apiSend("POST", "/api/shutdown");
        } catch {}
        showStopped();
      });
      $("#srcRows").addEventListener("change", (e) => {
        const id = e.target.closest("tr")?.dataset.id;
        if (!id) return;
        if (e.target.dataset.k === "use") {
          const set = new Set(S.settings.disabled || []);
          e.target.checked ? set.delete(id) : set.add(id);
          saveSettings({ disabled: [...set] });
        } else if (e.target.dataset.k === "fee") {
          const v = parseFloat(e.target.value.replace(",", "."));
          if (!(v >= 0 && v <= 10)) return;
          saveSettings({ fees: { ...(S.settings.fees || {}), [id]: v / 100 } });
          sourcesTable();
        }
      });
      $("#srcRows").addEventListener("click", (e) => {
        if (e.target.dataset.k !== "reset") return;
        const id = e.target.closest("tr").dataset.id;
        const fees = { ...(S.settings.fees || {}) };
        delete fees[id];
        saveSettings({ fees });
        sourcesTable();
      });
      everySecond(services);
    },
    destroy() {
      root = null;
    },
    onStatus: services,
  };
}

// ---- Boot ------------------------------------------------------------------------------------

async function boot() {
  try {
    [S.meta, S.settings] = await Promise.all([apiGet("/api/meta"), apiGet("/api/settings")]);
  } catch (e) {
    $("#view").innerHTML = `<div class="stopped">${esc(e.message)}</div>`;
    setTimeout(boot, 2000);
    return;
  }
  // Stable colours per source across all pages.
  let i = 0;
  for (const a of S.meta.assets) {
    for (const s of a.sources) {
      if (!S.sourceInfo.has(s.id)) S.sourceInfo.set(s.id, { ...s, color: PALETTE[i++ % PALETTE.length] });
    }
  }
  applyLang();
  applyTheme();
  applyScale();

  $("#vsSelect").addEventListener("change", (e) => saveSettings({ vs: e.target.value }, { reconnect: true }));
  $("#langSelect").addEventListener("change", (e) => {
    saveSettings({ lang: e.target.value });
    applyLang();
    render();
  });
  $("#themeBtn").addEventListener("click", () => {
    const next = THEMES[(THEMES.indexOf(S.settings.theme) + 1) % THEMES.length];
    saveSettings({ theme: next });
    applyTheme();
    toast(t("settings.theme") + ": " + t("theme." + next), "", 1500);
    if (parseRoute().name === "settings") render();
  });
  matchMedia("(prefers-color-scheme: dark)").addEventListener("change", () => S.settings.theme === "system" && applyTheme());
  window.addEventListener("hashchange", render);

  connect();
  apiGet("/api/alerts").then((r) => {
    S.rules = r;
    updateBadge();
  }).catch(() => {});
  apiGet("/api/alerts/events").then((e) => (S.events = e)).catch(() => {});
  render();
}

boot();
