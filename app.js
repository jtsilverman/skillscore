let skills = [];
let filtered = [];
let activeGrade = null;
let sortBy = 'relevance';
let currentPage = 1;
const PAGE_SIZE = 50;

// BM25 parameters
const K1 = 1.5;
const B = 0.75;

// Precomputed BM25 index
let bm25 = { docs: [], avgDl: 0, df: {}, N: 0 };

function tokenize(text) {
  return (text || '').toLowerCase().replace(/[^a-z0-9]/g, ' ').split(/\s+/).filter(t => t.length > 1);
}

function buildBM25Index(skills) {
  const docs = skills.map(s => {
    // Weight name tokens 3x, description 2x, source 1x
    const nameTokens = tokenize(s.name);
    const descTokens = tokenize(s.description);
    const sourceTokens = tokenize(s.source);
    const categoryTokens = tokenize(s.category);
    const tokens = [...nameTokens, ...nameTokens, ...nameTokens, ...descTokens, ...descTokens, ...sourceTokens, ...categoryTokens];
    const tf = {};
    for (const t of tokens) tf[t] = (tf[t] || 0) + 1;
    return { tf, dl: tokens.length };
  });

  const N = docs.length;
  const avgDl = docs.reduce((s, d) => s + d.dl, 0) / (N || 1);
  const df = {};
  for (const doc of docs) {
    for (const term of Object.keys(doc.tf)) {
      df[term] = (df[term] || 0) + 1;
    }
  }

  bm25 = { docs, avgDl, df, N };
}

function scoreBM25(query) {
  const terms = tokenize(query);
  if (terms.length === 0) return null;

  const scores = new Float64Array(bm25.N);
  for (const term of terms) {
    const docFreq = bm25.df[term] || 0;
    if (docFreq === 0) continue;
    const idf = Math.log(1 + (bm25.N - docFreq + 0.5) / (docFreq + 0.5));

    for (let i = 0; i < bm25.N; i++) {
      const doc = bm25.docs[i];
      const tf = doc.tf[term] || 0;
      if (tf === 0) continue;
      scores[i] += idf * (tf * (K1 + 1)) / (tf + K1 * (1 - B + B * doc.dl / bm25.avgDl));
    }
  }
  return scores;
}

async function init() {
  try {
    const resp = await fetch('scored-index.json');
    const data = await resp.json();
    skills = data.skills || [];
    document.getElementById('total-count').textContent = skills.length;
    document.getElementById('source-count').textContent = (data.sources || []).length;

    // Build BM25 index
    buildBM25Index(skills);

    // Populate category filter
    const categories = [...new Set(skills.map(s => s.category).filter(Boolean))].sort();
    const catSelect = document.getElementById('category-filter');
    for (const cat of categories) {
      const opt = document.createElement('option');
      opt.value = cat;
      opt.textContent = `${cat} (${skills.filter(s => s.category === cat).length})`;
      catSelect.appendChild(opt);
    }

    applyFilters();
  } catch (e) {
    document.getElementById('skill-list').innerHTML =
      '<div class="empty-state">Failed to load index. Run <code>skillscore index</code> first.</div>';
  }
}

function applyFilters() {
  const query = document.getElementById('search').value.trim();
  const category = document.getElementById('category-filter').value;

  // Compute BM25 relevance scores
  const relevanceScores = scoreBM25(query);
  const hasQuery = relevanceScores !== null;

  // Default to relevance sort when there's a query, score sort otherwise
  const effectiveSort = sortBy === 'relevance' && !hasQuery ? 'score' : sortBy;

  filtered = [];
  for (let i = 0; i < skills.length; i++) {
    const s = skills[i];

    // Category filter
    if (category && s.category !== category) continue;

    // Grade filter
    if (activeGrade && gradePrefix(s.grade) !== activeGrade) continue;

    // Query filter: only include skills with positive BM25 score
    if (hasQuery && relevanceScores[i] <= 0) continue;

    filtered.push({
      skill: s,
      relevance: hasQuery ? relevanceScores[i] : 0,
      idx: i,
    });
  }

  // Sort
  filtered.sort((a, b) => {
    if (effectiveSort === 'relevance' && hasQuery) {
      // Blend: 70% relevance, 30% quality score (normalized)
      const ra = a.relevance * 0.7 + (a.skill.score / 100) * 0.3;
      const rb = b.relevance * 0.7 + (b.skill.score / 100) * 0.3;
      return rb - ra;
    }
    if (effectiveSort === 'score') return b.skill.score - a.skill.score;
    if (effectiveSort === 'name') return a.skill.name.localeCompare(b.skill.name);
    if (effectiveSort === 'repo') return (a.skill.repo || '').localeCompare(b.skill.repo || '');
    return 0;
  });

  currentPage = 1;
  updateStats();
  render();
  renderPagination();
}

function updateStats() {
  const start = Math.min((currentPage - 1) * PAGE_SIZE + 1, filtered.length);
  const end = Math.min(currentPage * PAGE_SIZE, filtered.length);
  if (filtered.length === 0) {
    document.getElementById('stats-text').innerHTML = '<strong>0</strong> skills';
  } else {
    document.getElementById('stats-text').innerHTML =
      `Showing <strong>${start}-${end}</strong> of <strong>${filtered.length}</strong> skills`;
  }
}

function gradePrefix(grade) {
  return grade ? grade[0] : 'F';
}

function gradeClass(grade) {
  const p = gradePrefix(grade);
  return ['A', 'B', 'C', 'D', 'F'].includes(p) ? 'grade-' + p : 'grade-F';
}

function barColor(val) {
  if (val >= 90) return 'var(--green)';
  if (val >= 80) return 'var(--light-green)';
  if (val >= 70) return 'var(--yellow)';
  if (val >= 60) return 'var(--orange)';
  return 'var(--red)';
}

function sourceClass(source) {
  if (!source) return 'source-community';
  const lower = source.toLowerCase();
  if (lower.includes('official') || lower.includes('anthropic')) return 'source-official';
  if (lower.includes('local') || lower === 'jtsilverman') return 'source-local';
  return 'source-community';
}

function sourceLabel(source) {
  if (!source) return '';
  if (source.length > 20) return source.slice(0, 18) + '..';
  return source;
}

function render() {
  const list = document.getElementById('skill-list');
  const start = (currentPage - 1) * PAGE_SIZE;
  const page = filtered.slice(start, start + PAGE_SIZE);

  if (page.length === 0) {
    list.innerHTML = '<div class="empty-state">No skills match your search.</div>';
    return;
  }

  list.innerHTML = page.map((item, i) => {
    const s = item.skill;
    return `
    <div class="skill-card" onclick="toggle(this)" data-idx="${start + i}">
      <div class="skill-header">
        <span class="grade-badge ${gradeClass(s.grade)}">${s.grade}</span>
        <span class="skill-name">${esc(s.name)}</span>
        <span class="skill-score">${Math.round(s.score)}/100</span>
      </div>
      ${s.description ? `<div class="skill-desc">${esc(truncate(s.description, 120))}</div>` : ''}
      <div class="skill-meta">
        ${s.category ? `<span class="category-badge">${esc(s.category)}</span>` : ''}
        ${s.source ? `<span class="source-badge ${sourceClass(s.source)}">${esc(sourceLabel(s.source))}</span>` : ''}
        ${s.repo ? `<a href="https://github.com/${s.repo}" target="_blank">${s.repo}</a>` : ''}
      </div>
      <div class="skill-details">
        ${renderDims(s.dimensions)}
        ${renderSuggestions(s.suggestions)}
      </div>
    </div>`;
  }).join('');
}

function renderDims(d) {
  if (!d) return '';
  const dims = [
    ['Structure', d.structure],
    ['Description', d.description],
    ['Content', d.content],
    ['Engineering', d.engineering],
    ['Packaging', d.packaging],
  ];
  return dims.map(([label, val]) => `
    <div class="dim-row">
      <span class="dim-label">${label}</span>
      <div class="dim-bar">
        <div class="dim-bar-fill" style="width:${val}%;background:${barColor(val)}"></div>
      </div>
      <span class="dim-score">${Math.round(val)}</span>
    </div>
  `).join('');
}

function renderSuggestions(suggestions) {
  if (!suggestions || suggestions.length === 0) return '';
  const top = suggestions.slice(0, 3);
  return '<div style="margin-top:0.6rem;font-size:0.85rem;color:var(--text-dim)">' +
    top.map(s => {
      const icon = s.priority === 'high' ? '\u25cf' : s.priority === 'medium' ? '\u25d0' : '\u25cb';
      return `<div>${icon} ${esc(s.message)}</div>`;
    }).join('') + '</div>';
}

function renderPagination() {
  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const pag = document.getElementById('pagination');

  if (totalPages <= 1) {
    pag.innerHTML = '';
    return;
  }

  pag.innerHTML = `
    <button onclick="goPage(${currentPage - 1})" ${currentPage === 1 ? 'disabled' : ''}>Prev</button>
    <span class="page-info">Page ${currentPage} of ${totalPages}</span>
    <button onclick="goPage(${currentPage + 1})" ${currentPage === totalPages ? 'disabled' : ''}>Next</button>
  `;
}

function goPage(page) {
  const totalPages = Math.ceil(filtered.length / PAGE_SIZE);
  if (page < 1 || page > totalPages) return;
  currentPage = page;
  updateStats();
  render();
  renderPagination();
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

function toggle(el) {
  el.classList.toggle('expanded');
}

function setGradeFilter(grade, btn) {
  if (activeGrade === grade) {
    activeGrade = null;
    btn.classList.remove('active');
  } else {
    activeGrade = grade;
    document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
  }
  applyFilters();
}

function setSort(val) {
  sortBy = val;
  applyFilters();
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}

function truncate(str, len) {
  if (!str || str.length <= len) return str;
  return str.slice(0, len) + '...';
}

document.addEventListener('DOMContentLoaded', init);
