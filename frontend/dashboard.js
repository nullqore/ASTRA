function initThree() {
    if (typeof THREE === 'undefined') { console.error("THREE.js not loaded."); return; }
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 1, 10000);
    camera.position.z = 1000;
    const renderer = new THREE.WebGLRenderer({ canvas: document.getElementById('three-canvas'), alpha: true });
    renderer.setSize(window.innerWidth, window.innerHeight);
    const particles = new THREE.Points(new THREE.BufferGeometry(), new THREE.PointsMaterial({ size: 2, color: 0x00c3ff, blending: THREE.AdditiveBlending, transparent: true, opacity: 0.7 }));
    const positions = new Float32Array(5000 * 3);
    for (let i = 0; i < positions.length; i++) {
        positions[i] = Math.random() * 2000 - 1000;
    }
    particles.geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    scene.add(particles);
    let mouseX = 0;
    document.addEventListener('mousemove', (e) => { mouseX = e.clientX - window.innerWidth / 2; });
    window.addEventListener('resize', () => {
        camera.aspect = window.innerWidth / window.innerHeight;
        camera.updateProjectionMatrix();
        renderer.setSize(window.innerWidth, window.innerHeight);
    });
    function animate() {
        requestAnimationFrame(animate);
        camera.position.x += (mouseX - camera.position.x) * 0.05;
        camera.lookAt(scene.position);
        particles.rotation.y = Date.now() * 0.00005;
        renderer.render(scene, camera);
    }
    animate();
    console.log("Three.js background initialized successfully.");
}

document.addEventListener('DOMContentLoaded', () => {
    // This is the final, correct fix. It ensures the canvas cannot block clicks.
    document.getElementById('three-canvas').style.pointerEvents = 'none';

    initThree();

    const backToProjectsBtn = document.getElementById('back-to-projects-btn');
    if (backToProjectsBtn) {
        backToProjectsBtn.addEventListener('click', () => {
            localStorage.removeItem('activeProjectName');
            window.location.href = 'index.html';
        });
    }

    const API_BASE_URL = 'http://localhost:8080';

    // --- DOM Element Declarations ---
    const dashboardTitle = document.getElementById('dashboard-title');
    const addTargetForm = document.getElementById('add-target-form');
    const targetInput = document.getElementById('target-input');
    const targetTypeSelect = document.getElementById('target-type');
    const targetList = document.getElementById('target-list');
    const outOfScopeList = document.getElementById('out-of-scope-list');
    const toggleOutOfScopeBtn = document.getElementById('toggle-out-of-scope-btn');
    const moduleListDiv = document.getElementById('module-list');
    const confirmModal = document.getElementById('custom-confirm-modal');
    const confirmMessage = document.getElementById('custom-confirm-message');
    const confirmYesBtn = document.getElementById('custom-confirm-yes-btn');
    const confirmNoBtn = document.getElementById('custom-confirm-no-btn');

    // --- Core Logic ---

    // Check for an active project. If none, redirect to the main page.
    const projectName = localStorage.getItem('activeProjectName');
    if (!projectName) {
        window.location.href = 'index.html';
        return; // Stop further execution if no project is active
    }

    // Set the dashboard title with the active project name.
    dashboardTitle.textContent = `Project: ${projectName}`;

    // --- Function Declarations ---

    async function fetchProjectDetails() {
        try {
            targetList.innerHTML = '';
            outOfScopeList.innerHTML = '';
            const project = await fetchApi(`${API_BASE_URL}/api/projects/${projectName}`);
            renderTargets(project.wildcards, targetList, 'wildcard');
            renderTargets(project.domains, targetList, 'domain');
            renderTargets(project.out_of_scope, outOfScopeList, 'out-of-scope');
        } catch (error) {
            console.error('Failed to load project details:', error);
            showNotification('Could not load project details.', true);
        }
    }

    async function fetchAndRenderModules() {
        try {
            const modules = await fetchApi(`${API_BASE_URL}/api/modules`);
            moduleListDiv.innerHTML = '';
            modules.forEach(module => {
                const moduleItem = document.createElement('div');
                moduleItem.className = 'module-item';
                moduleItem.innerHTML = `
                    <h4>${module.name}</h4>
                    <p>${module.description || 'No description available.'}</p>
                `;
                if (module.name === 'Recon') {
                    moduleItem.addEventListener('click', () => {
                        window.location.href = 'recon.html';
                    });
                }
                moduleListDiv.appendChild(moduleItem);
            });
        } catch (error) {
            console.error('Failed to load modules:', error);
            showNotification('Could not load modules.', true);
        }
    }

    function renderTargets(targets, listElement, type) {
        if (!targets || !Array.isArray(targets)) return;
        targets.forEach(target => {
            const item = createTargetItem(target, type);
            listElement.appendChild(item);
        });
    }

    function createTargetItem(target, type) {
        const item = document.createElement('div');
        item.className = 'target-item';
        const displayTarget = type === 'wildcard' ? `*.${target}` : target;
        item.innerHTML = `<span class="target-name">${displayTarget}</span>`;
        const deleteBtn = document.createElement('button');
        deleteBtn.innerHTML = '&#x274C;';
        deleteBtn.className = 'delete-target-btn';
        deleteBtn.addEventListener('click', () => {
            showConfirmation(`Are you sure you want to delete "${displayTarget}"?`, async () => {
                await deleteTarget(target, type);
                item.remove();
            });
        });
        item.appendChild(deleteBtn);
        return item;
    }

    async function addTarget(target, type) {
        try {
            await fetchApi(`${API_BASE_URL}/api/projects/${projectName}/targets`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ target, type })
            });
            const listElement = type === 'out-of-scope' ? outOfScopeList : targetList;
            const item = createTargetItem(target, type);
            listElement.appendChild(item);
            targetInput.value = '';
            showNotification('Target added successfully!');
        } catch (error) {
            showNotification(error.message, true);
        }
    }

    async function deleteTarget(target, type) {
        try {
            await fetchApi(`${API_BASE_URL}/api/projects/${projectName}/targets?target=${target}&type=${type}`, {
                method: 'DELETE'
            });
            showNotification('Target deleted successfully!');
        } catch (error) {
            showNotification(error.message, true);
        }
    }

    function showConfirmation(message, callback) {
        confirmMessage.textContent = message;
        confirmModal.classList.remove('hidden');
        confirmYesBtn.onclick = () => {
            callback();
            confirmModal.classList.add('hidden');
        };
        confirmNoBtn.onclick = () => {
            confirmModal.classList.add('hidden');
        };
    }

    addTargetForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const target = targetInput.value.trim();
        const type = targetTypeSelect.value;
        if (!target) return;
        await addTarget(target, type);
    });

    toggleOutOfScopeBtn.addEventListener('click', () => {
        outOfScopeList.classList.toggle('hidden-view');
        toggleOutOfScopeBtn.classList.toggle('open');
        const text = toggleOutOfScopeBtn.querySelector('.btn-text');
        if (toggleOutOfScopeBtn.classList.contains('open')) {
            text.textContent = 'Hide Out of Scope';
        } else {
            text.textContent = 'Show Out of Scope';
        }
    });

    // --- Initial Data Fetch ---
    fetchProjectDetails();
    fetchAndRenderModules();
});

async function fetchApi(url, options) {
    const response = await fetch(url, options);
    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `Request failed with status ${response.status}`);
    }
    return response.json();
}

function showNotification(message, isError = false) {
    const notification = document.getElementById('notification');
    if (!notification) return;
    notification.textContent = message;
    notification.className = `notification show ${isError ? 'error' : ''}`;
    setTimeout(() => {
        notification.className = 'notification';
    }, 3000);
}