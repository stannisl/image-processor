package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const indexHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ImageProcessor</title>
    <style>
        /* ── Reset & Base ── */
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        :root {
            --bg:        #0f1117;
            --surface:   #1a1d27;
            --surface2:  #22263a;
            --border:    #2e3248;
            --accent:    #6c63ff;
            --accent2:   #a78bfa;
            --success:   #22c55e;
            --warning:   #f59e0b;
            --danger:    #ef4444;
            --text:      #e2e8f0;
            --muted:     #64748b;
            --radius:    14px;
            --radius-sm: 8px;
            --shadow:    0 8px 32px rgba(0,0,0,.45);
            --transition: .2s ease;
        }

        body {
            font-family: 'Inter', system-ui, sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
        }

        /* ── Header ── */
        header {
            width: 100%;
            padding: 20px 40px;
            display: flex;
            align-items: center;
            gap: 12px;
            border-bottom: 1px solid var(--border);
            background: var(--surface);
            position: sticky;
            top: 0;
            z-index: 100;
            backdrop-filter: blur(12px);
        }

        .logo {
            width: 36px; height: 36px;
            background: linear-gradient(135deg, var(--accent), var(--accent2));
            border-radius: 10px;
            display: flex; align-items: center; justify-content: center;
            font-size: 18px;
        }

        header h1 {
            font-size: 1.2rem;
            font-weight: 700;
            background: linear-gradient(135deg, var(--accent2), #e2e8f0);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        header span {
            margin-left: auto;
            font-size: .8rem;
            color: var(--muted);
        }

        /* ── Main layout ── */
        main {
            width: 100%;
            max-width: 960px;
            padding: 40px 20px 80px;
            display: flex;
            flex-direction: column;
            gap: 32px;
        }

        /* ── Upload zone ── */
        .upload-card {
            background: var(--surface);
            border: 1px solid var(--border);
            border-radius: var(--radius);
            padding: 40px;
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 20px;
            box-shadow: var(--shadow);
        }

        .drop-zone {
            width: 100%;
            border: 2px dashed var(--border);
            border-radius: var(--radius);
            padding: 48px 24px;
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 12px;
            cursor: pointer;
            transition: border-color var(--transition), background var(--transition);
            position: relative;
        }

        .drop-zone:hover,
        .drop-zone.drag-over {
            border-color: var(--accent);
            background: rgba(108,99,255,.06);
        }

        .drop-zone input[type="file"] {
            position: absolute;
            inset: 0;
            opacity: 0;
            cursor: pointer;
            width: 100%;
            height: 100%;
        }

        .drop-icon {
            width: 56px; height: 56px;
            background: rgba(108,99,255,.12);
            border-radius: 50%;
            display: flex; align-items: center; justify-content: center;
            font-size: 26px;
        }

        .drop-zone p { color: var(--muted); font-size: .9rem; text-align: center; }
        .drop-zone strong { color: var(--text); font-size: 1rem; }

        /* Preview strip */
        .preview-strip {
            display: flex;
            gap: 12px;
            flex-wrap: wrap;
            justify-content: center;
            width: 100%;
        }

        .preview-thumb {
            width: 80px; height: 80px;
            border-radius: var(--radius-sm);
            object-fit: cover;
            border: 2px solid var(--accent);
            animation: pop .2s ease;
        }

        @keyframes pop {
            from { transform: scale(.8); opacity: 0; }
            to   { transform: scale(1);  opacity: 1; }
        }

        /* Upload btn */
        .btn {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            padding: 12px 28px;
            border-radius: var(--radius-sm);
            font-size: .95rem;
            font-weight: 600;
            cursor: pointer;
            border: none;
            transition: all var(--transition);
        }

        .btn-primary {
            background: linear-gradient(135deg, var(--accent), #8b5cf6);
            color: #fff;
            box-shadow: 0 4px 16px rgba(108,99,255,.35);
        }

        .btn-primary:hover:not(:disabled) {
            transform: translateY(-1px);
            box-shadow: 0 6px 20px rgba(108,99,255,.5);
        }

        .btn-primary:disabled {
            opacity: .5;
            cursor: not-allowed;
        }

        .btn-danger {
            background: rgba(239,68,68,.12);
            color: var(--danger);
            border: 1px solid rgba(239,68,68,.25);
        }

        .btn-danger:hover { background: rgba(239,68,68,.22); }

        .btn-ghost {
            background: transparent;
            color: var(--muted);
            border: 1px solid var(--border);
        }

        .btn-ghost:hover { background: var(--surface2); color: var(--text); }

        /* ── Progress bar ── */
        .upload-progress {
            width: 100%;
            height: 4px;
            background: var(--border);
            border-radius: 99px;
            overflow: hidden;
            display: none;
        }

        .upload-progress.active { display: block; }

        .upload-progress-bar {
            height: 100%;
            background: linear-gradient(90deg, var(--accent), var(--accent2));
            border-radius: 99px;
            width: 0%;
            transition: width .3s ease;
        }

        /* ── Section title ── */
        .section-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .section-header h2 {
            font-size: 1rem;
            font-weight: 600;
            color: var(--muted);
            text-transform: uppercase;
            letter-spacing: .08em;
        }

        .badge {
            background: var(--surface2);
            border: 1px solid var(--border);
            border-radius: 99px;
            padding: 2px 10px;
            font-size: .75rem;
            color: var(--muted);
        }

        /* ── Image grid ── */
        .image-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
            gap: 20px;
        }

        /* ── Image card ── */
        .image-card {
            background: var(--surface);
            border: 1px solid var(--border);
            border-radius: var(--radius);
            overflow: hidden;
            box-shadow: var(--shadow);
            transition: transform var(--transition), border-color var(--transition);
            animation: pop .25s ease;
        }

        .image-card:hover { transform: translateY(-2px); border-color: var(--accent); }

        /* Thumbnail area */
        .card-thumb {
            width: 100%;
            aspect-ratio: 16/10;
            background: var(--surface2);
            position: relative;
            overflow: hidden;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .card-thumb img {
            width: 100%;
            height: 100%;
            object-fit: cover;
            transition: opacity .3s;
        }

        /* Processing overlay */
        .processing-overlay {
            position: absolute;
            inset: 0;
            background: rgba(15,17,23,.85);
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: 12px;
            backdrop-filter: blur(4px);
        }

        .spinner {
            width: 36px; height: 36px;
            border: 3px solid var(--border);
            border-top-color: var(--accent);
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }

        @keyframes spin { to { transform: rotate(360deg); } }

        .processing-overlay p {
            font-size: .8rem;
            color: var(--muted);
        }

        /* Status pill */
        .status-pill {
            position: absolute;
            top: 10px; right: 10px;
            padding: 3px 10px;
            border-radius: 99px;
            font-size: .7rem;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: .06em;
        }

        .status-pill.pending  { background: rgba(245,158,11,.2);  color: var(--warning); }
        .status-pill.done     { background: rgba(34,197,94,.2);   color: var(--success); }
        .status-pill.error    { background: rgba(239,68,68,.2);   color: var(--danger); }

        /* Card body */
        .card-body {
            padding: 14px 16px;
            display: flex;
            flex-direction: column;
            gap: 10px;
        }

        .card-name {
            font-size: .9rem;
            font-weight: 600;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .card-meta {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }

        .meta-chip {
            background: var(--surface2);
            border: 1px solid var(--border);
            border-radius: 99px;
            padding: 2px 10px;
            font-size: .72rem;
            color: var(--muted);
            font-family: monospace;
        }

        .card-id {
            font-size: .68rem;
            color: var(--muted);
            font-family: monospace;
            cursor: pointer;
            padding: 4px 8px;
            background: var(--surface2);
            border-radius: var(--radius-sm);
            border: 1px solid var(--border);
            transition: all var(--transition);
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .card-id:hover { border-color: var(--accent); color: var(--accent2); }

        /* Card actions */
        .card-actions {
            display: flex;
            gap: 8px;
        }

        .card-actions .btn {
            flex: 1;
            justify-content: center;
            padding: 8px 12px;
            font-size: .8rem;
        }

        /* ── Empty state ── */
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: var(--muted);
            grid-column: 1 / -1;
        }

        .empty-state .empty-icon { font-size: 48px; margin-bottom: 12px; }
        .empty-state p { font-size: .9rem; }

        /* ── Toast ── */
        #toast-container {
            position: fixed;
            bottom: 24px; right: 24px;
            display: flex;
            flex-direction: column;
            gap: 10px;
            z-index: 999;
        }

        .toast {
            padding: 12px 18px;
            border-radius: var(--radius-sm);
            font-size: .85rem;
            font-weight: 500;
            color: #fff;
            box-shadow: 0 4px 16px rgba(0,0,0,.4);
            animation: slideIn .25s ease;
            display: flex;
            align-items: center;
            gap: 10px;
            min-width: 220px;
        }

        @keyframes slideIn {
            from { transform: translateX(60px); opacity: 0; }
            to   { transform: translateX(0);    opacity: 1; }
        }

        .toast.success { background: #166534; border: 1px solid var(--success); }
        .toast.error   { background: #7f1d1d; border: 1px solid var(--danger); }
        .toast.info    { background: #1e1b4b; border: 1px solid var(--accent); }

        /* ── Lightbox ── */
        .lightbox {
            display: none;
            position: fixed;
            inset: 0;
            background: rgba(0,0,0,.85);
            z-index: 200;
            align-items: center;
            justify-content: center;
            backdrop-filter: blur(8px);
            padding: 20px;
        }

        .lightbox.open { display: flex; }

        .lightbox-inner {
            max-width: 90vw;
            max-height: 90vh;
            position: relative;
        }

        .lightbox-inner img {
            max-width: 100%;
            max-height: 90vh;
            border-radius: var(--radius);
            box-shadow: 0 20px 60px rgba(0,0,0,.7);
        }

        .lightbox-close {
            position: absolute;
            top: -14px; right: -14px;
            width: 32px; height: 32px;
            background: var(--surface);
            border: 1px solid var(--border);
            border-radius: 50%;
            cursor: pointer;
            display: flex; align-items: center; justify-content: center;
            font-size: 16px;
            color: var(--text);
            transition: all var(--transition);
        }

        .lightbox-close:hover { background: var(--danger); border-color: var(--danger); }

        /* Responsive */
        @media (max-width: 600px) {
            .upload-card { padding: 24px 16px; }
            .image-grid  { grid-template-columns: 1fr; }
            header h1    { font-size: 1rem; }
        }
    </style>
</head>
<body>

<!-- Header -->
<header>
    <div class="logo">🖼</div>
    <h1>ImageProcessor</h1>
    <span id="counter-label">0 images</span>
</header>

<!-- Main -->
<main>

    <!-- Upload card -->
    <div class="upload-card">
        <div class="drop-zone" id="dropZone">
            <input type="file" id="fileInput" accept="image/jpeg,image/png,image/gif" multiple>
            <div class="drop-icon">📤</div>
            <strong>Перетащите изображение сюда</strong>
            <p>или нажмите для выбора файла<br>
               <small style="color:var(--muted)">JPG, PNG, GIF · до 20 MB</small>
            </p>
        </div>

        <div class="preview-strip" id="previewStrip"></div>

        <div class="upload-progress" id="uploadProgress">
            <div class="upload-progress-bar" id="uploadProgressBar"></div>
        </div>

        <button class="btn btn-primary" id="uploadBtn" disabled>
            <span>⚡</span> Загрузить и обработать
        </button>
    </div>

    <!-- Gallery -->
    <div class="section-header">
        <h2>Галерея</h2>
        <span class="badge" id="galleryBadge">0</span>
    </div>

    <div class="image-grid" id="imageGrid">
        <div class="empty-state">
            <div class="empty-icon">🌌</div>
            <p>Загрузите первое изображение</p>
        </div>
    </div>

</main>

<!-- Lightbox -->
<div class="lightbox" id="lightbox">
    <div class="lightbox-inner">
        <div class="lightbox-close" id="lightboxClose">✕</div>
        <img id="lightboxImg" src="" alt="preview">
    </div>
</div>

<!-- Toasts -->
<div id="toast-container"></div>

<script>
/* ═══════════════════════════════════════════════
   STATE
═══════════════════════════════════════════════ */
const state = {
    files: [],          // выбранные File[]
    images: [],         // { id, filename, status, size, mimeType, originalUrl, processedUrl }
    polling: {}         // id -> intervalId
};

/* ═══════════════════════════════════════════════
   DOM refs
═══════════════════════════════════════════════ */
const dropZone        = document.getElementById('dropZone');
const fileInput       = document.getElementById('fileInput');
const previewStrip    = document.getElementById('previewStrip');
const uploadProgress  = document.getElementById('uploadProgress');
const uploadProgressBar = document.getElementById('uploadProgressBar');
const uploadBtn       = document.getElementById('uploadBtn');
const imageGrid       = document.getElementById('imageGrid');
const galleryBadge    = document.getElementById('galleryBadge');
const counterLabel    = document.getElementById('counter-label');
const lightbox        = document.getElementById('lightbox');
const lightboxImg     = document.getElementById('lightboxImg');
const lightboxClose   = document.getElementById('lightboxClose');

/* ═══════════════════════════════════════════════
   TOAST
═══════════════════════════════════════════════ */
function toast(msg, type = 'info', duration = 3500) {
    const icons = { success: '✅', error: '❌', info: 'ℹ️' };
    const el = document.createElement('div');
    el.className = 'toast ' + type;
    el.innerHTML = '<span>' + icons[type] + '</span><span>' + msg + '</span>';
    document.getElementById('toast-container').appendChild(el);
    setTimeout(() => el.remove(), duration);
}

/* ═══════════════════════════════════════════════
   COPY TO CLIPBOARD
═══════════════════════════════════════════════ */
function copyId(id) {
    navigator.clipboard.writeText(id).then(() => toast('ID скопирован', 'success'));
}

/* ═══════════════════════════════════════════════
   FILE SELECTION
═══════════════════════════════════════════════ */
fileInput.addEventListener('change', () => handleFiles(fileInput.files));

dropZone.addEventListener('dragover', e => {
    e.preventDefault();
    dropZone.classList.add('drag-over');
});
dropZone.addEventListener('dragleave', () => dropZone.classList.remove('drag-over'));
dropZone.addEventListener('drop', e => {
    e.preventDefault();
    dropZone.classList.remove('drag-over');
    handleFiles(e.dataTransfer.files);
});

function handleFiles(fileList) {
    state.files = [];
    previewStrip.innerHTML = '';

    const allowed = ['image/jpeg', 'image/png', 'image/gif'];
    const maxSize = 20 * 1024 * 1024;

    Array.from(fileList).forEach(f => {
        if (!allowed.includes(f.type)) {
            toast('Неподдерживаемый формат: ' + f.name, 'error');
            return;
        }
        if (f.size > maxSize) {
            toast('Файл слишком большой: ' + f.name, 'error');
            return;
        }
        state.files.push(f);

        const img = document.createElement('img');
        img.className = 'preview-thumb';
        img.src = URL.createObjectURL(f);
        previewStrip.appendChild(img);
    });

    uploadBtn.disabled = state.files.length === 0;
}

/* ═══════════════════════════════════════════════
   UPLOAD
═══════════════════════════════════════════════ */
uploadBtn.addEventListener('click', () => {
    if (!state.files.length) return;
    uploadFiles(state.files);
});

async function uploadFiles(files) {
    uploadBtn.disabled = true;
    uploadProgress.classList.add('active');

    const total = files.length;
    let done = 0;

    for (const file of files) {
        await uploadSingleFile(file);
        done++;
        uploadProgressBar.style.width = (done / total * 100) + '%';
    }

    setTimeout(() => {
        uploadProgress.classList.remove('active');
        uploadProgressBar.style.width = '0%';
    }, 600);

    // Сброс выбора
    state.files = [];
    previewStrip.innerHTML = '';
    fileInput.value = '';
    uploadBtn.disabled = true;
}

async function uploadSingleFile(file) {
    const fd = new FormData();
    fd.append('photo', file);

    try {
        const res = await fetch('/upload', { method: 'POST', body: fd });
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const body = await res.json();
        // Ответ: { "image": { id, filename, status, ... } }
        const img = body.image;

        const entry = {
            id:           img.id,
            filename:     img.filename || file.name,
            status:       img.status || 'pending',
            size:         file.size,
            mimeType:     file.type,
            originalUrl:  null,
            processedUrl: null,
        };

        // Грузим превью оригинала
        entry.originalUrl = '/image/' + img.id + '/original';

        state.images.unshift(entry);
        renderGallery();
        toast('Загружено: ' + entry.filename, 'success');

        // Если не готово — начинаем polling
        if (entry.status !== 'done') {
            startPolling(entry.id);
        } else {
            entry.processedUrl = '/image/' + img.id;
            renderGallery();
        }

    } catch (e) {
        toast('Ошибка загрузки ' + file.name + ': ' + e.message, 'error');
    }
}

/* ═══════════════════════════════════════════════
   POLLING
═══════════════════════════════════════════════ */
function startPolling(id) {
    if (state.polling[id]) return;
    state.polling[id] = setInterval(() => checkStatus(id), 2500);
}

async function checkStatus(id) {
    try {
        // Пытаемся получить обработанное изображение
        // Если 200 — готово; если 4xx/5xx — ещё в очереди
        const res = await fetch('/image/' + id, { method: 'HEAD' });

        if (res.ok) {
            clearInterval(state.polling[id]);
            delete state.polling[id];

            const entry = state.images.find(i => i.id === id);
            if (entry) {
                entry.status       = 'done';
                entry.processedUrl = '/image/' + id;
                renderGallery();
                toast('Обработано: ' + entry.filename, 'success');
            }
        }
    } catch (_) { /* network error — ignore */ }
}

/* ═══════════════════════════════════════════════
   DELETE
═══════════════════════════════════════════════ */
async function deleteImage(id) {
    if (!confirm('Удалить изображение?')) return;

    try {
        const res = await fetch('/image/' + id, { method: 'DELETE' });
        if (res.ok || res.status === 204) {
            clearInterval(state.polling[id]);
            delete state.polling[id];

            state.images = state.images.filter(i => i.id !== id);
            renderGallery();
            toast('Изображение удалено', 'info');
        } else {
            toast('Ошибка удаления', 'error');
        }
    } catch (e) {
        toast('Сетевая ошибка: ' + e.message, 'error');
    }
}

/* ═══════════════════════════════════════════════
   LIGHTBOX
═══════════════════════════════════════════════ */
function openLightbox(url) {
    lightboxImg.src = url;
    lightbox.classList.add('open');
}

lightboxClose.addEventListener('click', () => lightbox.classList.remove('open'));
lightbox.addEventListener('click', e => {
    if (e.target === lightbox) lightbox.classList.remove('open');
});

/* ═══════════════════════════════════════════════
   FORMAT HELPERS
═══════════════════════════════════════════════ */
function fmtSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1024 / 1024).toFixed(2) + ' MB';
}

function fmtExt(mime) {
    const map = { 'image/jpeg': 'JPG', 'image/png': 'PNG', 'image/gif': 'GIF' };
    return map[mime] || mime.split('/')[1]?.toUpperCase() || '?';
}

function shortId(id) {
    return id ? id.slice(0, 8) + '…' : '—';
}

/* ═══════════════════════════════════════════════
   RENDER
═══════════════════════════════════════════════ */
function renderGallery() {
    const count = state.images.length;
    galleryBadge.textContent = count;
    counterLabel.textContent = count + ' image' + (count !== 1 ? 's' : '');

    if (count === 0) {
        imageGrid.innerHTML = '<div class="empty-state"><div class="empty-icon">🌌</div><p>Загрузите первое изображение</p></div>';
        return;
    }

    imageGrid.innerHTML = state.images.map(img => renderCard(img)).join('');

    // Навешиваем обработчики после render
    state.images.forEach(img => {
        const card = document.getElementById('card-' + img.id);
        if (!card) return;

        // Клик по превью — лайтбокс
        const thumb = card.querySelector('.card-thumb img');
        if (thumb) {
            thumb.addEventListener('click', () => {
                const url = img.processedUrl || img.originalUrl;
                if (url) openLightbox(url);
            });
        }

        // Копировать ID
        const idEl = card.querySelector('.card-id');
        if (idEl) idEl.addEventListener('click', () => copyId(img.id));

        // Удалить
        const delBtn = card.querySelector('.btn-danger');
        if (delBtn) delBtn.addEventListener('click', () => deleteImage(img.id));

        // Скачать
        const dlBtn = card.querySelector('.btn-download');
        if (dlBtn) {
            dlBtn.addEventListener('click', () => {
                const url = img.processedUrl || img.originalUrl;
                if (!url) return;
                const a = document.createElement('a');
                a.href = url;
                a.download = img.filename;
                a.click();
            });
        }
    });
}

function renderCard(img) {
    const isDone    = img.status === 'done';
    const isPending = img.status !== 'done' && img.status !== 'error';
    const isError   = img.status === 'error';

    const thumbUrl  = img.originalUrl || '';
    const pillClass = isDone ? 'done' : isError ? 'error' : 'pending';
    const pillLabel = isDone ? 'Готово' : isError ? 'Ошибка' : 'Обработка';

    const thumbHTML = thumbUrl
        ? '<img src="' + thumbUrl + '" alt="' + img.filename + '" style="cursor:zoom-in">'
        : '<div style="color:var(--muted);font-size:.8rem">нет превью</div>';

    const overlayHTML = isPending ? '<div class="processing-overlay"><div class="spinner"></div><p>В обработке…</p></div>' : '';

    const dlDisabled = !isDone ? 'disabled style="opacity:.4;cursor:not-allowed"' : '';

    return '<div class="image-card" id="card-' + img.id + '">' +
        '<div class="card-thumb">' +
            thumbHTML +
            overlayHTML +
            '<div class="status-pill ' + pillClass + '">' + pillLabel + '</div>' +
        '</div>' +
        '<div class="card-body">' +
            '<div class="card-name" title="' + img.filename + '">' + img.filename + '</div>' +
            '<div class="card-meta">' +
                '<span class="meta-chip">' + fmtExt(img.mimeType) + '</span>' +
                '<span class="meta-chip">' + fmtSize(img.size) + '</span>' +
            '</div>' +
            '<div class="card-id" title="Нажмите, чтобы скопировать ID">🔑 ' + shortId(img.id) + '</div>' +
            '<div class="card-actions">' +
                '<button class="btn btn-ghost btn-download" ' + dlDisabled + '>⬇ Скачать</button>' +
                '<button class="btn btn-danger">🗑 Удалить</button>' +
            '</div>' +
        '</div>' +
    '</div>';
}
</script>
</body>
</html>`

type UIHandler struct{}

func (u *UIHandler) ServeUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, indexHTML)
}
