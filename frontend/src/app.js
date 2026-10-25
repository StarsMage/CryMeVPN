const BACKEND_SUPPORTS_SSH_PORT = true;

const DEFAULT_DOMAIN = 'www.openssl.org';
const DEFAULT_DEST_PORT = '443';

const DEFAULT_HTTP_PORT = '10809';
const DEFAULT_SOCKS_PORT = '10808';

const FINGERPRINTS = [
    { value: 'firefox', label: 'Firefox' },
    { value: 'chrome',  label: 'Chrome' },
    { value: 'safari',  label: 'Safari' },
    { value: 'edge',    label: 'Edge' },
];
const DNS_OPTIONS = [
    { value: '1.1.1.1',      label: 'Cloudflare (1.1.1.1)' },
    { value: '9.9.9.9',      label: 'Quad9 (9.9.9.9)' },
    { value: '94.140.14.14', label: 'AdGuard (94.140.14.14)' },
];
const DEFAULT_FINGERPRINT = 'firefox';
const DEFAULT_DNS = '1.1.1.1';

const hasBackend = () => !!(window.go && window.go.main && window.go.main.App);
const api = () => window.go.main.App;
const $ = (id) => document.getElementById(id);

const state = {
    servers: [],
    selectedId: null,
    connected: false,
    busy: false,
    connecting: false,
    deployToken: 0,
    menuServerId: null
};

const portOk = (p) => /^\d+$/.test(p) && Number(p) >= 1 && Number(p) <= 65535;

const localPortOk = (p) => portOk(p) && Number(p) >= 1024;

const domainOk = (d) =>
    d.length <= 253 &&
    /^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$/.test(d);

const inList = (list, v) => list.some(o => o.value === v);

function fillSelect(id, options, selected) {
    const sel = $(id);
    if (!sel) return;
    sel.innerHTML = '';
    options.forEach(o => {
        const opt = document.createElement('option');
        opt.value = o.value;
        opt.textContent = o.label;
        sel.appendChild(opt);
    });
    sel.value = selected;
}

async function loadServers() {
    if (!hasBackend()) {
        state.servers = [];
        return;
    }
    try {
        state.servers = (await api().GetSavedServer()) || [];
    } catch (e) {
        console.error('GetSavedServer failed:', e);
        state.servers = [];
        window.showModal('Error', 'Failed to load servers: ' + String(e));
    }
    if (!state.servers.find(s => s.id === state.selectedId)) {
        state.selectedId = state.servers[0]?.id || null;
    }
}

function showScreen(screenId) {
    document.querySelectorAll('.screen').forEach(s => {
        s.classList.remove('active');
        s.style.display = 'none';
    });
    const target = $(screenId);
    if (target) {
        target.classList.add('active');
        target.style.display = 'flex';
    }
}

window.showModal = function (title, message) {
    if ($('modal-title')) $('modal-title').innerText = title;
    if ($('modal-message')) $('modal-message').innerText = message;
    $('custom-modal')?.classList.remove('hidden');
};

function getSelected() {
    return state.servers.find(s => s.id === state.selectedId) || null;
}

function setAdvancedOpen(force) {
    const panel = $('advanced-panel');
    const btn = $('btn-advanced');
    if (!panel || !btn) return;
    const open = panel.classList.toggle('open', force);
    btn.classList.toggle('open', open);
    btn.setAttribute('aria-expanded', String(open));

    if (open) {
        setTimeout(() => {
            const screen = $('screen-setup');
            if (screen) screen.scrollTo({ top: screen.scrollHeight, behavior: 'smooth' });
        }, 360);
    }
}

function clearSetupForm() {
    ['input-ip', 'input-ssh-port', 'input-xray-port', 'input-user', 'input-pass']
        .forEach(id => { if ($(id)) $(id).value = ''; });
    if ($('input-ssh-port')) $('input-ssh-port').value = '22';
    if ($('input-xray-port')) $('input-xray-port').value = '443';
    if ($('input-user')) $('input-user').value = 'root';
    if ($('input-pass')) $('input-pass').type = 'password';

    if ($('input-domain')) $('input-domain').value = DEFAULT_DOMAIN;
    if ($('input-dest-port')) $('input-dest-port').value = DEFAULT_DEST_PORT;
    if ($('select-fingerprint')) $('select-fingerprint').value = DEFAULT_FINGERPRINT;
    if ($('select-dns')) $('select-dns').value = DEFAULT_DNS;
    if ($('input-http-port')) $('input-http-port').value = DEFAULT_HTTP_PORT;
    if ($('input-socks-port')) $('input-socks-port').value = DEFAULT_SOCKS_PORT;
    setAdvancedOpen(false);
}

function renderServers() {
    const container = $('servers-list');
    if (!container) return;
    container.innerHTML = '';

    if (state.servers.length === 0) {
        const empty = document.createElement('div');
        empty.className = 'no-servers';
        empty.textContent = 'No servers added yet';
        container.appendChild(empty);
        return;
    }

    state.servers.forEach(server => {
        const item = document.createElement('div');
        item.className = 'server-item' + (server.id === state.selectedId ? ' selected' : '');

        const info = document.createElement('div');
        info.className = 'server-info';

        const name = document.createElement('span');
        name.className = 'server-name';
        name.textContent = server.name || server.ip;

        const ip = document.createElement('span');
        ip.className = 'server-ip';
        ip.textContent = `${server.ip}:${server.port}`;

        info.append(name, ip);

        const dots = document.createElement('button');
        dots.className = 'btn-dots';
        dots.textContent = '•••';
        dots.addEventListener('click', (e) => {
            e.stopPropagation();
            openContextMenu(server.id, dots);
        });

        item.append(info, dots);
        item.addEventListener('click', () => selectServer(server.id));
        container.appendChild(item);
    });
}

function selectServer(id) {
    if (state.connected || state.busy) {
        window.showModal('Disconnect first', 'Disconnect before switching to another server.');
        return;
    }
    state.selectedId = id;
    renderServers();
}

function openContextMenu(serverId, anchor) {
    const menu = $('context-menu');
    if (!menu) return;
    state.menuServerId = serverId;
    menu.classList.remove('hidden');
    const rect = anchor.getBoundingClientRect();
    const left = Math.min(rect.right - menu.offsetWidth, window.innerWidth - menu.offsetWidth - 8);
    menu.style.left = Math.max(8, left) + 'px';
    menu.style.top = rect.bottom + 4 + 'px';
}

function closeContextMenu() {
    $('context-menu')?.classList.add('hidden');
}

async function deleteServer(id) {
    if (!hasBackend()) {
        window.showModal('Error', 'Backend is not available.');
        return;
    }
    if (state.busy) return;

    if (state.connected && state.selectedId === id) {
        await disconnect();
        if (state.connected) return;
    }

    try {
        await api().DeleteServer(id);
    } catch (err) {
        window.showModal('Delete Error', String(err));
        return;
    }

    await loadServers();
    renderServers();
    if (state.servers.length === 0) showScreen('screen-welcome');
}

async function deploy() {
    const ip = $('input-ip')?.value.trim() || '';
    const sshPort = $('input-ssh-port')?.value.trim() || '22';
    const xrayPort = $('input-xray-port')?.value.trim() || '443';
    const user = $('input-user')?.value.trim() || 'root';
    const pass = $('input-pass')?.value.trim() || '';

    const domain = ($('input-domain')?.value.trim().toLowerCase()) || DEFAULT_DOMAIN;
    const destPort = $('input-dest-port')?.value.trim() || DEFAULT_DEST_PORT;
    const fingerprint = $('select-fingerprint')?.value || DEFAULT_FINGERPRINT;
    const dns = $('select-dns')?.value || DEFAULT_DNS;
    const httpPort = $('input-http-port')?.value.trim() || DEFAULT_HTTP_PORT;
    const socksPort = $('input-socks-port')?.value.trim() || DEFAULT_SOCKS_PORT;

    if (!ip || !pass) {
        window.showModal('Error', 'Please fill in IP address and password');
        return;
    }
    if (!/^\d+$/.test(sshPort) || !/^\d+$/.test(xrayPort)) {
        window.showModal('Error', 'Ports must be numbers');
        return;
    }
    if (!portOk(sshPort) || !portOk(xrayPort)) {
        window.showModal('Error', 'Ports must be between 1 and 65535');
        return;
    }

    if (!domainOk(domain)) {
        setAdvancedOpen(true);
        window.showModal('Error', 'Invalid camouflage domain (example: amazon.com)');
        return;
    }
    if (!portOk(destPort)) {
        setAdvancedOpen(true);
        window.showModal('Error', 'Camouflage port must be a number from 1 to 65535');
        return;
    }
    if (!inList(FINGERPRINTS, fingerprint) || !inList(DNS_OPTIONS, dns)) {
        setAdvancedOpen(true);
        window.showModal('Error', 'Invalid fingerprint or DNS selection');
        return;
    }
    if (!localPortOk(httpPort) || !localPortOk(socksPort)) {
        setAdvancedOpen(true);
        window.showModal('Error', 'Local ports must be numbers from 1024 to 65535');
        return;
    }
    if (httpPort === socksPort) {
        setAdvancedOpen(true);
        window.showModal('Error', 'HTTP and SOCKS ports must be different');
        return;
    }

    if (!hasBackend()) {
        window.showModal('Error', 'Backend is not available.');
        return;
    }

    const token = ++state.deployToken;
    showScreen('screen-loading');
    const status = $('loading-status');
    if (status) status.textContent = 'Connecting via SSH & deploying Xray...';

    try {
        if (BACKEND_SUPPORTS_SSH_PORT) {
            await api().DeployServer(ip, sshPort, xrayPort, user, pass, domain, destPort, dns, fingerprint, httpPort, socksPort);
        } else {
            await api().DeployServer(ip, xrayPort, user, pass, domain, destPort, dns, fingerprint, httpPort, socksPort);
        }

        await loadServers();

        if (token !== state.deployToken) {
            renderServers();
            return;
        }

        if (state.servers.length > 0) {
            state.selectedId = state.servers[state.servers.length - 1].id;
        }
        renderServers();
        clearSetupForm();
        showScreen('screen-main');
    } catch (err) {
        if (token !== state.deployToken) return;
        window.showModal('Deployment Error', String(err));
        showScreen('screen-setup');
    }
}

function setConnectUI(mode) {
    const dot = $('status-dot');
    const text = $('status-text');
    const btn = $('btn-connect');
    const server = getSelected();

    dot?.classList.toggle('connected', mode === 'connected');
    if (btn) {
        btn.classList.toggle('active-connect', mode === 'connected');
        btn.disabled = (mode === 'connecting' || mode === 'disconnecting');
    }

    switch (mode) {
        case 'connecting':
            if (text) text.textContent = 'Connecting...';
            if (btn) btn.textContent = '...';
            break;
        case 'disconnecting':
            if (text) text.textContent = 'Disconnecting...';
            if (btn) btn.textContent = '...';
            break;
        case 'connected':
            if (text) text.textContent = server ? `Connected · ${server.ip}` : 'Connected';
            if (btn) btn.textContent = 'DISCONNECT';
            break;
        default:
            if (text) text.textContent = 'Disconnected';
            if (btn) btn.textContent = 'CONNECT';
    }
}

function listenXrayStatus() {
    if (!window.runtime || !window.runtime.EventsOn) return;

    window.runtime.EventsOn('xray:status', (msg) => {
        if (!state.connecting) return;
        const text = $('status-text');
        if (text) text.textContent = msg || 'Connecting...';
    });
}

async function connect() {
    const server = getSelected();
    if (!server) {
        window.showModal('No server', 'Add or select a server first.');
        return;
    }
    if (!hasBackend()) {
        window.showModal('Error', 'Backend is not available.');
        return;
    }
    state.busy = true;
    state.connecting = true;
    setConnectUI('connecting');
    try {
        await api().StartCore(server.id);
        state.connecting = false;
        state.connected = true;
        setConnectUI('connected');
    } catch (err) {
        state.connecting = false;
        setConnectUI('disconnected');
        window.showModal('Connection Error', String(err));
    } finally {
        state.connecting = false;
        state.busy = false;
    }
}

async function disconnect() {
    if (!hasBackend()) return;
    state.busy = true;
    setConnectUI('disconnecting');
    try {
        await api().StopCore();
        state.connected = false;
        setConnectUI('disconnected');
    } catch (err) {
        setConnectUI('connected');
        window.showModal('Disconnect Error', String(err));
    } finally {
        state.busy = false;
    }
}

async function initApp() {
    fillSelect('select-fingerprint', FINGERPRINTS, DEFAULT_FINGERPRINT);
    fillSelect('select-dns', DNS_OPTIONS, DEFAULT_DNS);
    listenXrayStatus();

    await loadServers();
    renderServers();
    showScreen(state.servers.length > 0 ? 'screen-main' : 'screen-welcome');

    $('btn-get-started')?.addEventListener('click', () => {
        clearSetupForm();
        showScreen('screen-setup');
    });

    $('btn-back-setup')?.addEventListener('click', () => {
        clearSetupForm();
        showScreen(state.servers.length > 0 ? 'screen-main' : 'screen-welcome');
    });

    $('btn-add-new')?.addEventListener('click', () => {
        clearSetupForm();
        showScreen('screen-setup');
    });

    $('btn-cancel-deploy')?.addEventListener('click', () => {
        state.deployToken++;
        showScreen(state.servers.length > 0 ? 'screen-main' : 'screen-welcome');
    });

    $('btn-toggle-pass')?.addEventListener('click', () => {
        const p = $('input-pass');
        if (p) p.type = p.type === 'password' ? 'text' : 'password';
    });

    $('btn-advanced')?.addEventListener('click', () => setAdvancedOpen());

    $('btn-deploy')?.addEventListener('click', deploy);

    $('btn-connect')?.addEventListener('click', () => {
        if (state.busy) return;
        state.connected ? disconnect() : connect();
    });

    $('modal-close')?.addEventListener('click', () => {
        $('custom-modal')?.classList.add('hidden');
    });

    $('menu-delete')?.addEventListener('click', () => {
        const id = state.menuServerId;
        closeContextMenu();
        if (id) deleteServer(id);
    });

    const menuEdit = $('menu-edit');
    if (menuEdit) menuEdit.style.display = 'none';

    document.addEventListener('click', (e) => {
        if (!e.target.closest('.btn-dots') && !e.target.closest('#context-menu')) {
            closeContextMenu();
        }
    });

    document.querySelectorAll('#screen-setup input').forEach(inp => {
        inp.addEventListener('keydown', (e) => { if (e.key === 'Enter') deploy(); });
    });

    if (!hasBackend()) {
        console.warn('Wails backend not found (window.go.main.App).');
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initApp);
} else {
    initApp();
}