const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

const statsEl = document.getElementById("stats");
const assetsListEl = document.getElementById("assets-list");
const scanResultsEl = document.getElementById("scan-results");
const createForm = document.getElementById("create-form");
const createMsg = document.getElementById("create-msg");
const reloadBtn = document.getElementById("reload-btn");

document.addEventListener("DOMContentLoaded", async () => {
  bindEvents();
  await loadDashboard();
  await loadAssets();
});

function bindEvents() {
  createForm.addEventListener("submit", onCreateAsset);
  reloadBtn.addEventListener("click", async () => {
    await loadDashboard();
    await loadAssets();
  });
}

async function api(path, options = {}) {
  const res = await fetch(`${API_URL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {})
    },
    ...options
  });

  let data = null;
  const text = await res.text();
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = text;
  }

  if (!res.ok) {
    throw new Error(data?.error || data?.message || `HTTP ${res.status}`);
  }

  return data;
}

async function loadDashboard() {
  try {
    const stats = await api("/assets/stats");
    statsEl.innerHTML = `
      <div class="stats-grid">
        <div class="stat-box">
          <div class="small">Total</div>
          <div>${stats.total ?? 0}</div>
        </div>
        <div class="stat-box">
          <div class="small">By Type</div>
          <pre>${JSON.stringify(stats.by_type ?? {}, null, 2)}</pre>
        </div>
        <div class="stat-box">
          <div class="small">By Status</div>
          <pre>${JSON.stringify(stats.by_status ?? {}, null, 2)}</pre>
        </div>
      </div>
    `;
  } catch (err) {
    statsEl.innerHTML = `<div class="err">Failed to load stats: ${err.message}</div>`;
  }
}

async function loadAssets() {
  try {
    const assets = await api("/assets");
    if (!Array.isArray(assets) || assets.length === 0) {
      assetsListEl.innerHTML = `<div>No assets yet</div>`;
      return;
    }

    assetsListEl.innerHTML = assets.map(asset => renderAsset(asset)).join("");

    document.querySelectorAll(".delete-btn").forEach(btn => {
      btn.addEventListener("click", async () => {
        const id = btn.dataset.id;
        await deleteAsset(id);
      });
    });

    document.querySelectorAll(".scan-btn").forEach(btn => {
      btn.addEventListener("click", async () => {
        const id = btn.dataset.id;
        const select = document.getElementById(`scan-type-${id}`);
        await startScan(id, select.value);
      });
    });

    document.querySelectorAll(".jobs-btn").forEach(btn => {
      btn.addEventListener("click", async () => {
        const id = btn.dataset.id;
        await loadAssetScans(id);
      });
    });

  } catch (err) {
    assetsListEl.innerHTML = `<div class="err">Failed to load assets: ${err.message}</div>`;
  }
}

function renderAsset(asset) {
  return `
    <div class="asset-item">
      <h3>${escapeHtml(asset.name)}</h3>
      <div><b>ID:</b> ${escapeHtml(asset.id)}</div>
      <div><b>Type:</b> ${escapeHtml(asset.type)}</div>
      <div><b>Status:</b> ${escapeHtml(asset.status)}</div>
      <div><b>Created:</b> ${escapeHtml(asset.created_at || "")}</div>

      <div class="asset-actions">
        <select id="scan-type-${asset.id}">
          ${getScanOptions(asset.type)}
        </select>
        <button class="scan-btn" data-id="${asset.id}">Start Scan</button>
        <button class="jobs-btn" data-id="${asset.id}">View Scans</button>
        <button class="delete-btn" data-id="${asset.id}">Delete</button>
      </div>
    </div>
  `;
}

function getScanOptions(assetType) {
  if (assetType === "domain") {
    return `
      <option value="dns">dns</option>
      <option value="whois">whois</option>
      <option value="subdomain">subdomain</option>
      <option value="cert_trans">cert_trans</option>
      <option value="ssl">ssl</option>
      <option value="tech">tech</option>
      <option value="all" selected>all</option>
    `;
  }
  if (assetType === "ip") {
    return `
      <option value="ip">ip</option>
      <option value="port" selected>port</option>
      <option value="asn">asn</option>
    `;
  }
  return `
    <option value="tech" selected>tech</option>
  `;
}

async function onCreateAsset(e) {
  e.preventDefault();
  createMsg.textContent = "";
  createMsg.className = "msg";

  const name = document.getElementById("asset-name").value.trim();
  const type = document.getElementById("asset-type").value;

  try {
    await api("/assets", {
      method: "POST",
      body: JSON.stringify({ name, type })
    });

    createForm.reset();
    createMsg.textContent = "Asset created successfully";
    createMsg.classList.add("ok");

    await loadDashboard();
    await loadAssets();
  } catch (err) {
    createMsg.textContent = err.message;
    createMsg.classList.add("err");
  }
}

async function deleteAsset(id) {
  try {
    await api(`/assets/${id}`, { method: "DELETE" });
    scanResultsEl.innerHTML = `<div class="ok">Deleted asset ${escapeHtml(id)}</div>`;
    await loadDashboard();
    await loadAssets();
  } catch (err) {
    scanResultsEl.innerHTML = `<div class="err">Delete failed: ${err.message}</div>`;
  }
}

async function startScan(assetId, scanType) {
  try {
    const job = await api(`/assets/${assetId}/scan`, {
      method: "POST",
      body: JSON.stringify({ scan_type: scanType })
    });

    scanResultsEl.innerHTML = `
      <div class="ok">Scan started</div>
      <pre>${JSON.stringify(job, null, 2)}</pre>
      <button id="load-results-btn">Load Results For This Job</button>
    `;

    document.getElementById("load-results-btn").addEventListener("click", async () => {
      await loadScanResults(job.id);
    });
  } catch (err) {
    scanResultsEl.innerHTML = `<div class="err">Start scan failed: ${err.message}</div>`;
  }
}

async function loadAssetScans(assetId) {
  try {
    const data = await api(`/assets/${assetId}/scans`);
    const scans = data.scans || [];

    if (scans.length === 0) {
      scanResultsEl.innerHTML = `<div>No scan jobs for asset ${escapeHtml(assetId)}</div>`;
      return;
    }

    scanResultsEl.innerHTML = `
      <h3>Scan Jobs</h3>
      ${scans.map(job => `
        <div class="asset-item">
          <div><b>Job ID:</b> ${escapeHtml(job.id)}</div>
          <div><b>Type:</b> ${escapeHtml(job.scan_type)}</div>
          <div><b>Status:</b> ${escapeHtml(job.status)}</div>
          <div><b>Results:</b> ${escapeHtml(String(job.results))}</div>
          <button class="job-result-btn" data-job-id="${job.id}">View Results</button>
        </div>
      `).join("")}
    `;

    document.querySelectorAll(".job-result-btn").forEach(btn => {
      btn.addEventListener("click", async () => {
        await loadScanResults(btn.dataset.jobId);
      });
    });
  } catch (err) {
    scanResultsEl.innerHTML = `<div class="err">Load scans failed: ${err.message}</div>`;
  }
}

async function loadScanResults(jobId) {
  try {
    const data = await api(`/scan-jobs/${jobId}/results`);
    scanResultsEl.innerHTML = `
      <h3>Scan Results</h3>
      <div><b>Job ID:</b> ${escapeHtml(data.job_id || jobId)}</div>
      <div><b>Scan Type:</b> ${escapeHtml(data.scan_type || "")}</div>
      <pre>${JSON.stringify(data.results || [], null, 2)}</pre>
    `;
  } catch (err) {
    scanResultsEl.innerHTML = `<div class="err">Load results failed: ${err.message}</div>`;
  }
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}